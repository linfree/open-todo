package database

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

func (d *DB) GetTags(userID string) ([]Tag, error) {
	rows, err := d.db.Query(
		"SELECT id, user_id, name, color, created_at FROM tags WHERE user_id = ? ORDER BY name",
		userID,
	)
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
	if tags == nil {
		tags = []Tag{}
	}
	return tags, rows.Err()
}

func (d *DB) SaveTag(t *Tag) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	now := time.Now().UTC().Format(time.RFC3339)
	if t.CreatedAt == "" {
		t.CreatedAt = now
	}

	_, err := d.db.Exec(
		`INSERT OR REPLACE INTO tags
		(id, user_id, name, color, created_at)
		VALUES (?, ?, ?, ?, ?)`,
		t.ID, t.UserID, t.Name, t.Color, t.CreatedAt,
	)
	return err
}

func (d *DB) DeleteTag(id string) error {
	_, err := d.db.Exec("DELETE FROM tags WHERE id = ?", id)
	return err
}
