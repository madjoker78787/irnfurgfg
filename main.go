// main.go
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
)

// Task - структура задачи
type Task struct {
	ID    int    `json:"id"`    // уникальный id задачи
	Title string `json:"title"` // заголовок задачи
	Done  bool   `json:"done"`  // выполнена или нет
}

// Создадим простое in-memory хранилище задач с синхронизацией
type TaskStore struct {
	mu     sync.Mutex
	tasks  map[int]Task
	nextID int
}

func NewTaskStore() *TaskStore {
	return &TaskStore{
		tasks:  make(map[int]Task),
		nextID: 1,
	}
}

// Добавляем новую задачу, возвращаем созданную
func (s *TaskStore) CreateTask(title string) Task {
	s.mu.Lock()
	defer s.mu.Unlock()
	task := Task{
		ID:    s.nextID,
		Title: title,
		Done:  false,
	}
	s.tasks[s.nextID] = task
	s.nextID++
	return task
}

// Получаем все задачи
func (s *TaskStore) GetAllTasks() []Task {
	s.mu.Lock()
	defer s.mu.Unlock()
	tasks := make([]Task, 0, len(s.tasks))
	for _, t := range s.tasks {
		tasks = append(tasks, t)
	}
	return tasks
}

// Обновляем состояние задачи (например, mark done)
func (s *TaskStore) UpdateTask(id int, done bool) (Task, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	task, exists := s.tasks[id]
	if !exists {
		return Task{}, false
	}
	task.Done = done
	s.tasks[id] = task
	return task, true
}

// Удаляем задачу
func (s *TaskStore) DeleteTask(id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.tasks[id]; !exists {
		return false
	}
	delete(s.tasks, id)
	return true
}

func main() {
	store := NewTaskStore()

	http.HandleFunc("/tasks", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			// Возвращаем все задачи
			tasks := store.GetAllTasks()
			jsonResponse(w, tasks, http.StatusOK)
		case "POST":
			// Создаем новую задачу – читаем из тела JSON с ключом "title"
			var input struct {
				Title string `json:"title"`
			}
			if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.Title == "" {
				http.Error(w, "Invalid input", http.StatusBadRequest)
				return
			}
			task := store.CreateTask(input.Title)
			jsonResponse(w, task, http.StatusCreated)
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	// Эндпоинт для обновления и удаления задачи по ID
	http.HandleFunc("/tasks/", func(w http.ResponseWriter, r *http.Request) {
		// URL вида /tasks/{id}
		idStr := r.URL.Path[len("/tasks/"):]
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid task id", http.StatusBadRequest)
			return
		}
		switch r.Method {
		case "PUT":
			// Обновляем поле Done
			var input struct {
				Done bool `json:"done"`
			}
			if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
				http.Error(w, "Invalid input", http.StatusBadRequest)
				return
			}
			task, ok := store.UpdateTask(id, input.Done)
			if !ok {
				http.Error(w, "Task not found", http.StatusNotFound)
				return
			}
			jsonResponse(w, task, http.StatusOK)
		case "DELETE":
			if !store.DeleteTask(id) {
				http.Error(w, "Task not found", http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	log.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// helper для ответа с JSON
func jsonResponse(w http.ResponseWriter, data interface{}, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}
