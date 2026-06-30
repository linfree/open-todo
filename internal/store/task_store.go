package store

import (
	"time"
	"github.com/google/uuid"
)

type Task struct {
	ID          string  `json:"id"`
	UserID      string  `json:"user_id"`
	Title       string  `json:"title"`
	Description *string `json:"description"`
	Completed   bool    `json:"completed"`
	Priority    string  `json:"priority"`
	Status      string  `json:"status"`
	ListID      string  `json:"list_id"`
	Tags        string  `json:"tags"`
	SubTasks    string  `json:"sub_tasks"`
	Reminders   string  `json:"reminders"`
	DueDate     *string `json:"due_date"`
	Deleted     bool    `json:"deleted"`
	DeletedAt   *string `json:"deleted_at"`
	OrderNum    int     `json:"order_num"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

func (s *Store) GetTasks() ([]Task, error) {
	rows, err := s.db.Query("SELECT id, user_id, title, description, completed, priority, status, list_id, tags, sub_tasks, reminders, due_date, deleted, deleted_at, order_num, created_at, updated_at FROM tasks WHERE deleted = 0 ORDER BY order_num ASC, created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		var completed int
		var deleted int
		if err := rows.Scan(&t.ID, &t.UserID, &t.Title, &t.Description, &completed, &t.Priority, &t.Status, &t.ListID, &t.Tags, &t.SubTasks, &t.Reminders, &t.DueDate, &deleted, &t.DeletedAt, &t.OrderNum, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		t.Completed = completed != 0
		t.Deleted = deleted != 0
		tasks = append(tasks, t)
	}
	if tasks == nil {
		tasks = []Task{}
	}
	return tasks, rows.Err()
}

func (s *Store) SaveTask(t *Task) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	if t.CreatedAt == "" {
		t.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}
	t.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	completed := 0
	if t.Completed {
		completed = 1
	}
	deleted := 0
	if t.Deleted {
		deleted = 1
	}

	_, err := s.db.Exec(
		`INSERT OR REPLACE INTO tasks (id, user_id, title, description, completed, priority, status, list_id, tags, sub_tasks, reminders, due_date, deleted, deleted_at, order_num, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		t.ID, t.UserID, t.Title, t.Description, completed, t.Priority, t.Status, t.ListID, t.Tags, t.SubTasks, t.Reminders, t.DueDate, deleted, t.DeletedAt, t.OrderNum, t.CreatedAt, t.UpdatedAt,
	)
	if err != nil {
		return err
	}
	s.logChange("tasks", t.ID, "insert", t.UpdatedAt)
	return nil
}

func (s *Store) SoftDeleteTask(id string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.Exec("UPDATE tasks SET deleted = 1, deleted_at = ?, updated_at = ? WHERE id = ?", now, now, id)
	if err != nil {
		return err
	}
	s.logChange("tasks", id, "delete", now)
	return nil
}

func (s *Store) logChange(table, recordID, action, timestamp string) {
	s.db.Exec("INSERT INTO sync_log (table_name, record_id, action, timestamp, synced) VALUES (?, ?, ?, ?, 0)",
		table, recordID, action, timestamp)
}
