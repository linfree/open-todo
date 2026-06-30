package database

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

type User struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Password  string `json:"-"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type SyncRecord struct {
	TableName string `json:"table_name"`
	RecordID  string `json:"record_id"`
	Action    string `json:"action"`
	Timestamp string `json:"timestamp"`
}

type DB struct {
	db *sql.DB
}

func Open(driver, dsn string) (*DB, error) {
	db, err := sql.Open(driver, dsn+"?_journal_mode=WAL")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	if err := migrate(db); err != nil {
		return nil, err
	}
	return &DB{db: db}, nil
}

func (d *DB) Close() error { return d.db.Close() }

func migrate(db *sql.DB) error {
	sqls := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY, email TEXT UNIQUE NOT NULL, name TEXT NOT NULL DEFAULT '',
			password TEXT NOT NULL, created_at TEXT NOT NULL, updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS lists (
			id TEXT PRIMARY KEY, user_id TEXT NOT NULL, name TEXT NOT NULL,
			icon TEXT, color TEXT, order_num INTEGER DEFAULT 0,
			created_at TEXT NOT NULL, updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS tasks (
			id TEXT PRIMARY KEY, user_id TEXT NOT NULL, title TEXT NOT NULL,
			description TEXT, completed INTEGER DEFAULT 0, priority TEXT DEFAULT 'none',
			status TEXT DEFAULT 'todo', list_id TEXT NOT NULL, tags TEXT DEFAULT '[]',
			sub_tasks TEXT DEFAULT '[]', reminders TEXT DEFAULT '[]', due_date TEXT,
			deleted INTEGER DEFAULT 0, deleted_at TEXT, order_num INTEGER DEFAULT 0,
			created_at TEXT NOT NULL, updated_at TEXT NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_server_tasks_user ON tasks(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_server_tasks_updated ON tasks(updated_at)`,
	}
	for _, s := range sqls {
		if _, err := db.Exec(s); err != nil {
			return fmt.Errorf("migrate: %w\nSQL: %s", err, s)
		}
	}
	return nil
}

