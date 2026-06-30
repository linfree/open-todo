package database

import (
	"fmt"
	"time"
)

func (d *DB) StoreChanges(userID string, changes []SyncRecord) error {
	for _, c := range changes {
		switch c.TableName {
		case "tasks":
			_, err := d.db.Exec(
				`INSERT OR REPLACE INTO tasks
				(id, user_id, title, description, completed, priority, status, list_id,
				 tags, sub_tasks, reminders, due_date, deleted, deleted_at, order_num,
				 created_at, updated_at)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				c.RecordID, userID, "", "", 0, "none", "todo", "",
				"[]", "[]", "[]", nil, 0, nil, 0,
				time.Now().UTC().Format(time.RFC3339),
				time.Now().UTC().Format(time.RFC3339),
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
