package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
)

type Task struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

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

func (s *TaskStore) GetAllTasks() []Task {
	s.mu.Lock()
	defer s.mu.Unlock()
	tasks := make([]Task, len(s.tasks))
	for _, t := range s.tasks {
		tasks = append(tasks, t)
	}
	return tasks
}

func (s *TaskStore) UpdateTask(id int, done bool) (Task, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	task, ok := s.tasks[id]
	if !ok {
		return Task{}, false
	}
	task.Done = done
	s.tasks[id] = task
	return task, true
}

func (s *TaskStore) DeleteTask(id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.tasks[id]; !exists {
		return false
	}
	delete(s.tasks, id)
	return true
}

func jsonResponse(w http.ResponseWriter, data interface{}, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

func main() {
	store := NewTaskStore()
	http.HandleFunc("/tasks", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			//возвращаем все задачи
			tasks := store.GetAllTasks()
			jsonResponse(w, tasks, http.StatusOK)
		case http.MethodPost:
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
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/tasks/", func(w http.ResponseWriter, r *http.Request) {
		idStr := r.URL.Path[len("/tasks/"):]
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid task id", http.StatusBadRequest)
			return
		}
		switch r.Method {
		case http.MethodPut:
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
		case http.MethodDelete:
			if !store.DeleteTask(id) {
				http.Error(w, "Task not allowed", http.StatusMethodNotAllowed)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	log.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
