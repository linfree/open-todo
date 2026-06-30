package store

import (
	"testing"
)

func setupStore(t *testing.T) *Store {
	t.Helper()
	s, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	return s
}

func TestSaveAndGetTask(t *testing.T) {
	s := setupStore(t)
	defer s.Close()

	task := &Task{Title: "测试任务", ListID: "default", Priority: "medium"}
	if err := s.SaveTask(task); err != nil {
		t.Fatalf("save: %v", err)
	}

	tasks, err := s.GetTasks()
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if tasks[0].Title != "测试任务" {
		t.Errorf("title = %s, want 测试任务", tasks[0].Title)
	}
	if tasks[0].ID == "" {
		t.Error("task id is empty")
	}
}

func TestSoftDeleteTask(t *testing.T) {
	s := setupStore(t)
	defer s.Close()

	task := &Task{Title: "待删除", ListID: "default"}
	s.SaveTask(task)
	s.SoftDeleteTask(task.ID)

	tasks, _ := s.GetTasks()
	if len(tasks) != 0 {
		t.Errorf("expected 0 active tasks, got %d", len(tasks))
	}
}

func TestGetUnsyncedChanges(t *testing.T) {
	s := setupStore(t)
	defer s.Close()

	task := &Task{Title: "同步测试", ListID: "default"}
	s.SaveTask(task)

	changes, err := s.GetUnsyncedChanges()
	if err != nil {
		t.Fatalf("get changes: %v", err)
	}
	if len(changes) == 0 {
		t.Error("expected unsynced changes")
	}
	if changes[0].TableName != "tasks" {
		t.Errorf("table = %s, want tasks", changes[0].TableName)
	}
}
