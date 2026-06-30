package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

// User represents a registered user account.
type User struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Password  string `json:"-"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// SyncRecord is a single change from a client to be applied or replicated.
type SyncRecord struct {
	TableName string          `json:"table_name"`
	RecordID  string          `json:"record_id"`
	Action    string          `json:"action"`
	Timestamp string          `json:"timestamp"`
	Data      json.RawMessage `json:"data,omitempty"`
}

// Database is the shared interface implemented by every backend.
type Database interface {
	CreateUser(u *User) (*User, error)
	GetUserByEmail(email string) (*User, error)
	GetUserByID(id string) (*User, error)

	StoreChanges(userID string, changes []SyncRecord) error
	GetChangesSince(userID string, since any /* time.Time */) ([]SyncRecord, error)

	GetCategories(userID string) ([]Category, error)
	SaveCategory(c *Category) error
	DeleteCategory(id string) error

	GetTags(userID string) ([]Tag, error)
	SaveTag(t *Tag) error
	DeleteTag(id string) error

	Migrate() error
	CleanupDeleted() error
	Close() error
}

// New opens the database described by driver/DSN and returns its
// Database handle.  Supported drivers: "sqlite", "postgres".
func New(driver, dsn string) (Database, error) {
	switch driver {
	case "sqlite":
		return newSQLite(dsn)
	case "postgres":
		return newPostgres(dsn)
	default:
		return nil, fmt.Errorf("database: unsupported driver %q", driver)
	}
}

// openAndMigrate is a small helper used by backend constructors so
// they can keep the same error-handling shape.
func openAndMigrate(driverName, dsn string, migrate func(*sql.DB) error) (*sql.DB, error) {
	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, err
	}
	if err := migrate(db); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

// ---------- helpers shared by all backends ----------

type taskPayload struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description *string `json:"description"`
	Completed   bool    `json:"completed"`
	Priority    string  `json:"priority"`
	Status      string  `json:"status"`
	ListID      string  `json:"list_id"`
	Tags        string  `json:"tags"`
	SubTasks    string  `json:"sub_tasks"`
	Reminders   string  `json:"reminders"`
	DueDate     *string `json:"due_date"`
	Deleted     bool    `json:"deleted"`
	DeletedAt   *string `json:"deleted_at"`
	OrderNum    int     `json:"order_num"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

func taskDefaults(p *taskPayload) {
	if p.Title == "" {
		p.Title = ""
	}
	if p.Priority == "" {
		p.Priority = "none"
	}
	if p.Status == "" {
		p.Status = "todo"
	}
	if p.Tags == "" {
		p.Tags = "[]"
	}
	if p.SubTasks == "" {
		p.SubTasks = "[]"
	}
	if p.Reminders == "" {
		p.Reminders = "[]"
	}
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
