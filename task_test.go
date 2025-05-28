package main

import (
	"testing"
)

func TestCreateAndGetAllTasks(t *testing.T) {
	s := NewTaskStore() // создаём новый инстанс
	task := s.CreateTask("Test task")

	tasks := s.GetAllTasks()
	if len(tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(tasks))
	}
	if tasks[0].ID != task.ID {
		t.Errorf("Expected task ID %d, got %d", task.ID, tasks[0].ID)
	}
}
