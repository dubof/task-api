package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid" // генератор уникальных ID
)

type Task struct {
	ID        string      `json:"id"`
	Status    string      `json:"status"`
	Result    interface{} `json:"result,omitempty"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at,omitempty"`
}

type TaskManager struct {
	tasks map[string]*Task // хранилище задачч
	mu    sync.RWMutex     // защита от одновременного доступа
}

// создание нового менеджера задач
func NewTaskManager() *TaskManager {
	return &TaskManager{
		tasks: make(map[string]*Task), // инициализация хранилища
	}
}

// новая задача
func (tm *TaskManager) CreateTask() *Task {
	taskID := uuid.New().String() // генератор айди

	task := &Task{
		ID:        taskID,
		Status:    "ожидает",
		CreatedAt: time.Now(),
	}

	// безопасное добавление задачи
	tm.mu.Lock()
	tm.tasks[taskID] = task
	tm.mu.Unlock()

	// запуск обработки в фоне
	go tm.processTask(taskID)

	return task
}

// имитация долгой задачи
func (tm *TaskManager) processTask(taskID string) {
	// время имитации в минутах
	time.Sleep(1 * time.Minute)

	tm.mu.Lock()
	defer tm.mu.Unlock()

	// обновление
	if task, exists := tm.tasks[taskID]; exists {
		task.Status = "завершено"
		task.Result = map[string]string{
			"сообщение": "Задача успешно выполнена",
			"данные":    "Пример результата",
		}
		task.UpdatedAt = time.Now()
	}
}

// получение задачи по id
func (tm *TaskManager) GetTask(taskID string) (*Task, bool) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	task, exists := tm.tasks[taskID]
	return task, exists
}

func main() {
	// инициализация менеджера
	taskManager := NewTaskManager()

	// роутер для создания задач
	http.HandleFunc("/tasks", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
			return
		}

		// создание и возврат задачи
		task := taskManager.CreateTask()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(task)
	})

	// проверка статуса
	http.HandleFunc("/tasks/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
			return
		}

		// извлечение id ищ url
		taskID := r.URL.Path[len("/tasks/"):]
		if taskID == "" {
			http.Error(w, "Необходим ID задачи", http.StatusBadRequest)
			return
		}

		// получение и возврат задачи
		task, exists := taskManager.GetTask(taskID)
		if !exists {
			http.Error(w, "Задача не найдена", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(task)
	})

	// запуск сервера
	fmt.Println("Сервер запущен на :8080")
	http.ListenAndServe(":8080", nil)
}
