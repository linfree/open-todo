package store

import (
	"time"
	"github.com/google/uuid"
)

type Tag struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Name      string `json:"name"`
	Color     string `json:"color"`
	CreatedAt string `json:"created_at"`
}

func (s *Store) GetTags() ([]Tag, error) {
	rows, err := s.db.Query("SELECT id, user_id, name, color, created_at FROM tags ORDER BY name ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tags []Tag
	for rows.Next() {
		var t Tag
		if err := rows.Scan(&t.ID, &t.UserID, &t.Name, &t.Color, &t.CreatedAt); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	if tags == nil { tags = []Tag{} }
	return tags, rows.Err()
}

func (s *Store) SaveTag(t *Tag) error {
	if t.ID == "" { t.ID = uuid.New().String() }
	if t.CreatedAt == "" { t.CreatedAt = time.Now().UTC().Format(time.RFC3339) }
	_, err := s.db.Exec("INSERT OR REPLACE INTO tags (id, user_id, name, color, created_at) VALUES (?, ?, ?, ?, ?)",
		t.ID, t.UserID, t.Name, t.Color, t.CreatedAt)
	return err
}

func (s *Store) DeleteTag(id string) error {
	_, err := s.db.Exec("DELETE FROM tags WHERE id = ?", id)
	return err
}
