package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/google/uuid"
)

// postgresDB is the PostgreSQL-backed Database implementation.
type postgresDB struct {
	db *sql.DB
}

func newPostgres(dsn string) (*postgresDB, error) {
	db, err := openAndMigrate("pgx", dsn, migratePostgres)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(25)
	return &postgresDB{db: db}, nil
}

func migratePostgres(db *sql.DB) error {
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
	}
	for _, s := range sqls {
		if _, err := db.Exec(s); err != nil {
			return fmt.Errorf("migrate postgres: %w\nSQL: %s", err, s)
		}
	}
	return nil
}

func (d *postgresDB) Migrate() error { return migratePostgres(d.db) }
func (d *postgresDB) Close() error    { return d.db.Close() }

// ---------- users ----------

func (d *postgresDB) CreateUser(u *User) (*User, error) {
	_, err := d.db.Exec(
		`INSERT INTO users (id, email, name, password, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (id) DO NOTHING`,
		u.ID, u.Email, u.Name, u.Password, u.CreatedAt, u.UpdatedAt,
	)
	return u, err
}

func (d *postgresDB) GetUserByEmail(email string) (*User, error) {
	var u User
	err := d.db.QueryRow(
		`SELECT id, email, name, password, created_at, updated_at
		 FROM users WHERE email = $1`, email,
	).Scan(&u.ID, &u.Email, &u.Name, &u.Password, &u.CreatedAt, &u.UpdatedAt)
	return &u, err
}

func (d *postgresDB) GetUserByID(id string) (*User, error) {
	var u User
	err := d.db.QueryRow(
		`SELECT id, email, name, password, created_at, updated_at
		 FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Email, &u.Name, &u.Password, &u.CreatedAt, &u.UpdatedAt)
	return &u, err
}

// ---------- sync ----------

func (d *postgresDB) StoreChanges(userID string, changes []SyncRecord) error {
	for _, c := range changes {
		switch c.TableName {
		case "tasks":
			var p taskPayload
			if c.Data != nil && len(c.Data) > 0 {
				if err := json.Unmarshal(c.Data, &p); err != nil {
					return fmt.Errorf("store task %s: parse data: %w", c.RecordID, err)
				}
			}
			taskDefaults(&p)
			if p.CreatedAt == "" {
				p.CreatedAt = time.Now().UTC().Format(time.RFC3339)
			}
			if p.UpdatedAt == "" {
				p.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
			}
			_, err := d.db.Exec(
				`INSERT INTO tasks
				(id, user_id, title, description, completed, priority, status, list_id,
				 tags, sub_tasks, reminders, due_date, deleted, deleted_at, order_num,
				 created_at, updated_at)
				VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)
				ON CONFLICT (id) DO UPDATE SET
					user_id = EXCLUDED.user_id,
					title = EXCLUDED.title,
					description = EXCLUDED.description,
					completed = EXCLUDED.completed,
					priority = EXCLUDED.priority,
					status = EXCLUDED.status,
					list_id = EXCLUDED.list_id,
					tags = EXCLUDED.tags,
					sub_tasks = EXCLUDED.sub_tasks,
					reminders = EXCLUDED.reminders,
					due_date = EXCLUDED.due_date,
					deleted = EXCLUDED.deleted,
					deleted_at = EXCLUDED.deleted_at,
					order_num = EXCLUDED.order_num,
					updated_at = EXCLUDED.updated_at`,
				c.RecordID, userID, p.Title, p.Description, boolToInt(p.Completed),
				p.Priority, p.Status, p.ListID, p.Tags, p.SubTasks, p.Reminders,
				p.DueDate, boolToInt(p.Deleted), p.DeletedAt, p.OrderNum,
				p.CreatedAt, p.UpdatedAt,
			)
			if err != nil {
				return fmt.Errorf("store task %s: %w", c.RecordID, err)
			}
		}
	}
	return nil
}

func (d *postgresDB) GetChangesSince(userID string, since any) ([]SyncRecord, error) {
	s, ok := since.(time.Time)
	if !ok {
		return nil, fmt.Errorf("postgres: GetChangesSince expects time.Time, got %T", since)
	}
	rows, err := d.db.Query(
		`SELECT 'tasks' as table_name, id, updated_at
		 FROM tasks WHERE user_id = $1 AND updated_at > $2`,
		userID, s.UTC().Format(time.RFC3339),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var changes []SyncRecord
	for rows.Next() {
		var c SyncRecord
		if err := rows.Scan(&c.TableName, &c.RecordID, &c.Timestamp); err != nil {
			return nil, err
		}
		c.Action = "insert"
		changes = append(changes, c)
	}
	if changes == nil {
		changes = []SyncRecord{}
	}
	return changes, rows.Err()
}

// ---------- categories ----------

func (d *postgresDB) GetCategories(userID string) ([]Category, error) {
	rows, err := d.db.Query(
		`SELECT id, user_id, name, icon, color, order_num, created_at, updated_at
		 FROM categories WHERE user_id = $1 ORDER BY order_num`,
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

func (d *postgresDB) SaveCategory(c *Category) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	now := time.Now().UTC().Format(time.RFC3339)
	if c.CreatedAt == "" {
		c.CreatedAt = now
	}
	c.UpdatedAt = now

	_, err := d.db.Exec(
		`INSERT INTO categories
		(id, user_id, name, icon, color, order_num, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (id) DO UPDATE SET
			user_id = EXCLUDED.user_id,
			name = EXCLUDED.name,
			icon = EXCLUDED.icon,
			color = EXCLUDED.color,
			order_num = EXCLUDED.order_num,
			updated_at = EXCLUDED.updated_at`,
		c.ID, c.UserID, c.Name, c.Icon, c.Color, c.OrderNum, c.CreatedAt, c.UpdatedAt,
	)
	return err
}

func (d *postgresDB) DeleteCategory(id string) error {
	_, err := d.db.Exec("DELETE FROM categories WHERE id = $1", id)
	return err
}

// ---------- tags ----------

func (d *postgresDB) GetTags(userID string) ([]Tag, error) {
	rows, err := d.db.Query(
		`SELECT id, user_id, name, color, created_at
		 FROM tags WHERE user_id = $1 ORDER BY name`,
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

func (d *postgresDB) SaveTag(t *Tag) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	now := time.Now().UTC().Format(time.RFC3339)
	if t.CreatedAt == "" {
		t.CreatedAt = now
	}

	_, err := d.db.Exec(
		`INSERT INTO tags
		(id, user_id, name, color, created_at)
		VALUES ($1,$2,$3,$4,$5)
		ON CONFLICT (id) DO UPDATE SET
			user_id = EXCLUDED.user_id,
			name = EXCLUDED.name,
			color = EXCLUDED.color`,
		t.ID, t.UserID, t.Name, t.Color, t.CreatedAt,
	)
	return err
}

func (d *postgresDB) DeleteTag(id string) error {
	_, err := d.db.Exec("DELETE FROM tags WHERE id = $1", id)
	return err
}

// ---------- maintenance ----------

func (d *postgresDB) CleanupDeleted() error {
	_, err := d.db.Exec("DELETE FROM tasks WHERE deleted = 1")
	return err
}
