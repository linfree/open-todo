package schema

import "fmt"

// Table name constants.
const (
	TableUsers      = "users"
	TableLists      = "lists"
	TableTasks      = "tasks"
	TableCategories = "categories"
	TableTags       = "tags"
)

// Tables is the ordered list of all tables for migrations.
var Tables = []string{
	TableUsers,
	TableLists,
	TableTasks,
	TableCategories,
	TableTags,
}

// CurrentVersion is the latest schema version.
const CurrentVersion = 2

// Migration holds a set of SQL statements for a schema version upgrade.
type Migration struct {
	Version int
	SQL     []string
}

// Migrations contains the schema upgrade path from start to CurrentVersion.
var Migrations = []Migration{
	{
		Version: 1,
		SQL: []string{
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
		},
	},
	{
		Version: 2,
		SQL: []string{
			`CREATE TABLE IF NOT EXISTS categories (
				id TEXT PRIMARY KEY, user_id TEXT NOT NULL DEFAULT '',
				name TEXT NOT NULL, icon TEXT, color TEXT NOT NULL DEFAULT '#6b7280',
				order_num INTEGER NOT NULL DEFAULT 0,
				created_at TEXT NOT NULL, updated_at TEXT NOT NULL
			)`,
			`CREATE INDEX IF NOT EXISTS idx_categories_user ON categories(user_id)`,
			`CREATE TABLE IF NOT EXISTS tags (
				id TEXT PRIMARY KEY, user_id TEXT NOT NULL DEFAULT '',
				name TEXT NOT NULL, color TEXT NOT NULL DEFAULT '#3b82f6',
				created_at TEXT NOT NULL
			)`,
			`CREATE INDEX IF NOT EXISTS idx_tags_user ON tags(user_id)`,
		},
	},
}

// CreateTablesSQL returns the full SQL to create all tables from scratch.
// This is a convenience for initializing a fresh database.
func CreateTablesSQL() []string {
	var sql []string
	for _, m := range Migrations {
		sql = append(sql, m.SQL...)
	}
	return sql
}

// ApplyMigrations applies all migrations from the current version to the target.
func ApplyMigrations(exec func(string) error, currentVersion, targetVersion int) error {
	for _, m := range Migrations {
		if m.Version > currentVersion && m.Version <= targetVersion {
			for _, s := range m.SQL {
				if err := exec(s); err != nil {
					return fmt.Errorf("migration v%d: %w\nSQL: %s", m.Version, err, s)
				}
			}
		}
	}
	return nil
}
