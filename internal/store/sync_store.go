package store

import (
	"encoding/json"
	"fmt"
)

type ChangeRecord struct {
	ID        int64  `json:"id"`
	TableName string `json:"table_name"`
	RecordID  string `json:"record_id"`
	Action    string `json:"action"`
	Timestamp string `json:"timestamp"`
	Synced    bool   `json:"synced"`
}

func (s *Store) GetUnsyncedChanges() ([]ChangeRecord, error) {
	rows, err := s.db.Query("SELECT id, table_name, record_id, action, timestamp FROM sync_log WHERE synced = 0 ORDER BY id ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var changes []ChangeRecord
	for rows.Next() {
		var c ChangeRecord
		if err := rows.Scan(&c.ID, &c.TableName, &c.RecordID, &c.Action, &c.Timestamp); err != nil {
			return nil, err
		}
		changes = append(changes, c)
	}
	if changes == nil {
		changes = []ChangeRecord{}
	}
	return changes, rows.Err()
}

func (s *Store) MarkChangesSynced(ids []int64) error {
	for _, id := range ids {
		if _, err := s.db.Exec("UPDATE sync_log SET synced = 1 WHERE id = ?", id); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) GetFullRecord(tableName, recordID string) (json.RawMessage, error) {
	switch tableName {
	case "tasks":
		var t Task
		var completed, deleted int
		row := s.db.QueryRow(
			"SELECT id, user_id, title, description, completed, priority, status, list_id, tags, sub_tasks, reminders, due_date, deleted, deleted_at, order_num, created_at, updated_at FROM tasks WHERE id = ?",
			recordID,
		)
		if err := row.Scan(&t.ID, &t.UserID, &t.Title, &t.Description, &completed, &t.Priority, &t.Status, &t.ListID, &t.Tags, &t.SubTasks, &t.Reminders, &t.DueDate, &deleted, &t.DeletedAt, &t.OrderNum, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("GetFullRecord tasks/%s: %w", recordID, err)
		}
		t.Completed = completed != 0
		t.Deleted = deleted != 0
		return json.Marshal(t)
	case "lists":
		var l TaskList
		row := s.db.QueryRow(
			"SELECT id, user_id, name, icon, color, order_num, created_at, updated_at FROM lists WHERE id = ?",
			recordID,
		)
		if err := row.Scan(&l.ID, &l.UserID, &l.Name, &l.Icon, &l.Color, &l.OrderNum, &l.CreatedAt, &l.UpdatedAt); err != nil {
			return nil, fmt.Errorf("GetFullRecord lists/%s: %w", recordID, err)
		}
		return json.Marshal(l)
	}
	return nil, fmt.Errorf("GetFullRecord: unknown table %s", tableName)
}

func (s *Store) PutFullRecord(tableName, recordID string, data json.RawMessage) error {
	switch tableName {
	case "tasks":
		var t Task
		if err := json.Unmarshal(data, &t); err != nil {
			return fmt.Errorf("PutFullRecord decode: %w", err)
		}
		return s.SaveTask(&t)
	case "lists":
		var l TaskList
		if err := json.Unmarshal(data, &l); err != nil {
			return fmt.Errorf("PutFullRecord decode: %w", err)
		}
		return s.SaveList(&l)
	}
	return fmt.Errorf("PutFullRecord: unknown table %s", tableName)
}
