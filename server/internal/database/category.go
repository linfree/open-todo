package database

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

func (d *DB) GetCategories(userID string) ([]Category, error) {
	rows, err := d.db.Query(
		"SELECT id, user_id, name, icon, color, order_num, created_at, updated_at FROM categories WHERE user_id = ? ORDER BY order_num",
		userID,
	)
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
	if cats == nil {
		cats = []Category{}
	}
	return cats, rows.Err()
}

func (d *DB) SaveCategory(c *Category) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	now := time.Now().UTC().Format(time.RFC3339)
	if c.CreatedAt == "" {
		c.CreatedAt = now
	}
	c.UpdatedAt = now

	_, err := d.db.Exec(
		`INSERT OR REPLACE INTO categories
		(id, user_id, name, icon, color, order_num, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		c.ID, c.UserID, c.Name, c.Icon, c.Color, c.OrderNum, c.CreatedAt, c.UpdatedAt,
	)
	return err
}

func (d *DB) DeleteCategory(id string) error {
	_, err := d.db.Exec("DELETE FROM categories WHERE id = ?", id)
	return err
}
