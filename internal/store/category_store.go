package store

import (
	"time"
	"github.com/google/uuid"
)

type Category struct {
	ID        string  `json:"id"`
	UserID    string  `json:"user_id"`
	Name      string  `json:"name"`
	Icon      *string `json:"icon"`
	Color     string  `json:"color"`
	OrderNum  int     `json:"order_num"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

func (s *Store) GetCategories() ([]Category, error) {
	rows, err := s.db.Query("SELECT id, user_id, name, icon, color, order_num, created_at, updated_at FROM categories ORDER BY order_num ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cats []Category
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.ID, &c.UserID, &c.Name, &c.Icon, &c.Color, &c.OrderNum, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		cats = append(cats, c)
	}
	if cats == nil { cats = []Category{} }
	return cats, rows.Err()
}

func (s *Store) SaveCategory(c *Category) error {
	if c.ID == "" { c.ID = uuid.New().String() }
	if c.CreatedAt == "" { c.CreatedAt = time.Now().UTC().Format(time.RFC3339) }
	c.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.Exec("INSERT OR REPLACE INTO categories (id, user_id, name, icon, color, order_num, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		c.ID, c.UserID, c.Name, c.Icon, c.Color, c.OrderNum, c.CreatedAt, c.UpdatedAt)
	return err
}

func (s *Store) DeleteCategory(id string) error {
	_, err := s.db.Exec("DELETE FROM categories WHERE id = ?", id)
	return err
}
