package database

import (
	"encoding/json"
	"fmt"
	"time"
)

type taskPayload struct {
	ID          string  `json:"id"`
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

func taskDefaults(p *taskPayload) {
	if p.Title == "" {
		p.Title = ""
	}
	if p.Priority == "" {
		p.Priority = "none"
	}
	if p.Status == "" {
		p.Status = "todo"
	}
	if p.Tags == "" {
		p.Tags = "[]"
	}
	if p.SubTasks == "" {
		p.SubTasks = "[]"
	}
	if p.Reminders == "" {
		p.Reminders = "[]"
	}
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func (d *DB) StoreChanges(userID string, changes []SyncRecord) error {
	for _, c := range changes {
		switch c.TableName {
		case "tasks":
			var p taskPayload
			if c.Data != nil && len(c.Data) > 0 {
				if err := json.Unmarshal(c.Data, &p); err != nil {
					return fmt.Errorf("store task %s: parse data: %w", c.RecordID, err)
				}
			}
			taskDefaults(&p)
			if p.CreatedAt == "" {
				p.CreatedAt = time.Now().UTC().Format(time.RFC3339)
			}
			if p.UpdatedAt == "" {
				p.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
			}
			_, err := d.db.Exec(
				`INSERT OR REPLACE INTO tasks
				(id, user_id, title, description, completed, priority, status, list_id,
				 tags, sub_tasks, reminders, due_date, deleted, deleted_at, order_num,
				 created_at, updated_at)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				c.RecordID, userID, p.Title, p.Description, boolToInt(p.Completed),
				p.Priority, p.Status, p.ListID, p.Tags, p.SubTasks, p.Reminders,
				p.DueDate, boolToInt(p.Deleted), p.DeletedAt, p.OrderNum,
				p.CreatedAt, p.UpdatedAt,
			)
			if err != nil {
				return fmt.Errorf("store task %s: %w", c.RecordID, err)
			}
		}
	}
	return nil
}

func (d *DB) GetChangesSince(userID string, since time.Time) ([]SyncRecord, error) {
	rows, err := d.db.Query(
		"SELECT 'tasks' as table_name, id, updated_at FROM tasks WHERE user_id = ? AND updated_at > ?",
		userID, since.UTC().Format(time.RFC3339),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var changes []SyncRecord
	for rows.Next() {
		var c SyncRecord
		if err := rows.Scan(&c.TableName, &c.RecordID, &c.Timestamp); err != nil {
			return nil, err
		}
		c.Action = "insert"
		changes = append(changes, c)
	}
	if changes == nil {
		changes = []SyncRecord{}
	}
	return changes, rows.Err()
}
