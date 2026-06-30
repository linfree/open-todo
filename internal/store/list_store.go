package store

import (
	"time"
	"github.com/google/uuid"
)

type TaskList struct {
	ID        string  `json:"id"`
	UserID    string  `json:"user_id"`
	Name      string  `json:"name"`
	Icon      *string `json:"icon"`
	Color     *string `json:"color"`
	OrderNum  int     `json:"order_num"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

func (s *Store) GetLists() ([]TaskList, error) {
	rows, err := s.db.Query("SELECT id, user_id, name, icon, color, order_num, created_at, updated_at FROM lists ORDER BY order_num ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lists []TaskList
	for rows.Next() {
		var l TaskList
		if err := rows.Scan(&l.ID, &l.UserID, &l.Name, &l.Icon, &l.Color, &l.OrderNum, &l.CreatedAt, &l.UpdatedAt); err != nil {
			return nil, err
		}
		lists = append(lists, l)
	}
	if lists == nil {
		lists = []TaskList{}
	}
	return lists, rows.Err()
}

func (s *Store) SaveList(l *TaskList) error {
	if l.ID == "" {
		l.ID = uuid.New().String()
	}
	if l.CreatedAt == "" {
		l.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}
	l.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	_, err := s.db.Exec(
		`INSERT OR REPLACE INTO lists (id, user_id, name, icon, color, order_num, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		l.ID, l.UserID, l.Name, l.Icon, l.Color, l.OrderNum, l.CreatedAt, l.UpdatedAt,
	)
	if err != nil {
		return err
	}
	s.logChange("lists", l.ID, "insert", l.UpdatedAt)
	return nil
}

func (s *Store) DeleteList(id string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	s.db.Exec("DELETE FROM lists WHERE id = ?", id)
	s.logChange("lists", id, "delete", now)
	return nil
}
