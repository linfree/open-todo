package store

import (
	"github.com/google/uuid"
	"time"
)

func (s *Store) IsReminderSent(taskID string, reminderTime int64) (bool, error) {
	var count int
	err := s.db.QueryRow(
		"SELECT COUNT(*) FROM sent_reminders WHERE task_id = ? AND reminder_time = ?",
		taskID, reminderTime,
	).Scan(&count)
	return count > 0, err
}

func (s *Store) MarkReminderSent(id, taskID string, reminderTime int64, data string) error {
	if id == "" {
		id = uuid.New().String()
	}
	sentAt := time.Now().Unix()
	_, err := s.db.Exec(
		"INSERT OR REPLACE INTO sent_reminders (id, task_id, reminder_time, sent_at, reminder_data) VALUES (?, ?, ?, ?, ?)",
		id, taskID, reminderTime, sentAt, data,
	)
	return err
}
