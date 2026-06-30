package store

import "github.com/linfree/open-todo/server/internal/database"

// Category is re-exported from database package.
type Category = database.Category

// Tag is re-exported from database package.
type Tag = database.Tag

// Store wraps the database.DB for category/tag operations.
type Store struct {
	DB *database.DB
}

func New(db *database.DB) *Store {
	return &Store{DB: db}
}
