# Open Todo Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将 my-todo 从 Tauri (Rust) + Tauri Android 重构为 Go 桌面 + PWA + Go 服务端，改名 open-todo。

**Architecture:** Go 桌面用 Gin + WebView2/systray + modernc.org/sqlite，PWA 用 React + Dexie.js + Service Worker，Go 服务端做 JWT 认证 + 同步中继，所有端通过 DatabaseAdapter 接口和 SyncEngine 连接。

**Tech Stack:** Go 1.25, modernc.org/sqlite, Gin, systray, WebView2, React 19, TypeScript, Dexie.js, vite-plugin-pwa, PostgreSQL (可选)

## Global Constraints

- 项目名称: open-todo (Go module: `github.com/linfree/open-todo`)
- 纯 Go SQLite: `modernc.org/sqlite` (零 CGo)
- 前端端口: 开发 5173 (Vite), 桌面/服务端 18080 (Gin)
- 同步冲突: Last-Write-Wins (比较 `updated_at`)
- 软删除: `deleted` 字段标记, 30天后清理
- JSON 字段: tags/sub_tasks/reminders 以 JSON 字符串存储 (SQLite), 原生对象 (Dexie)
- 配置文件路径: `~/.open-todo/config.json`
- 数据库路径: `~/.open-todo/open-todo.db` (WAL 模式)
- 必须支持三平台桌面 (Windows/macOS/Linux) + PWA 移动 + Go 服务端

---

### Task 1: 初始化 Go 模块与项目结构

**Files:**
- Create: `go.mod`
- Create: `go.sum`
- Create: `Makefile`
- Create: `.gitignore` (update)
- Create: `cmd/open-todo/main.go` (stub)
- Create: `cmd/open-todo/embed.go` (stub)

**Interfaces:**
- Produces: Go module `github.com/linfree/open-todo`, `make build/run/dev` 命令, `cmd/open-todo` 入口

- [ ] **Step 1: 初始化 Go module**

```bash
cd G:/dev/AI/my_todo
go mod init github.com/linfree/open-todo
```

- [ ] **Step 2: 添加基础依赖**

```bash
go get github.com/gin-gonic/gin@latest
go get modernc.org/sqlite@latest
go get github.com/getlantern/systray@latest
go get github.com/google/uuid@latest
go get golang.org/x/sys@latest
go mod tidy
```

- [ ] **Step 3: 创建 Makefile**

```makefile
.PHONY: all web build run dev clean test

all: web build

web:
	cd web && npm install && npm run build

# PWA 构建 (输出到 web/dist-pwa)
web-pwa:
	cd web && npm install && npm run build:pwa

build:
	go build -ldflags="-s -w" -o bin/open-todo ./cmd/open-todo/

build-win:
	GOOS=windows GOARCH=amd64 go build -ldflags="-H windowsgui -s -w" -o bin/open-todo.exe ./cmd/open-todo/

build-linux:
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/open-todo-linux ./cmd/open-todo/

build-mac:
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o bin/open-todo-mac ./cmd/open-todo/

run:
	go run ./cmd/open-todo/

dev:
	cd web && npm run dev & go run ./cmd/open-todo/

clean:
	rm -rf bin/ cmd/open-todo/web-dist/

test:
	go test ./... -v -count=1 -timeout 60s
```

- [ ] **Step 4: 创建入口 stub**

`cmd/open-todo/main.go`:
```go
package main

import "fmt"

func main() {
	fmt.Println("open-todo starting...")
}
```

- [ ] **Step 5: 验证编译**

```bash
go build ./cmd/open-todo/
```
Expected: 成功, 无错误

- [ ] **Step 6: Commit**

```bash
git add go.mod go.sum Makefile cmd/
git commit -m "feat: init Go module and project structure for open-todo"
```

---

### Task 2: Schema 定义层 (Go)

**Files:**
- Create: `internal/pkg/schema/schema.go`
- Create: `internal/pkg/schema/migrate.go`

**Interfaces:**
- Produces: `schema.Tables` (表定义常量), `schema.CreateTablesSQL()` → `[]string`, `schema.Migrations` (迁移历史), `schema.RunMigrations(db *sql.DB) error`

- [ ] **Step 1: 编写表结构常量**

`internal/pkg/schema/schema.go`:
```go
package schema

const TableTasks = "tasks"
const TableLists = "lists"
const TableUsers = "users"
const TableSentReminders = "sent_reminders"
const TableSyncLog = "sync_log"

var Tables = []string{TableUsers, TableLists, TableTasks, TableSentReminders, TableSyncLog}

const CurrentVersion = 1

var CreateTablesSQL = []string{
	`CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		email TEXT UNIQUE NOT NULL,
		name TEXT NOT NULL DEFAULT '',
		password TEXT NOT NULL,
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL
	)`,
	`CREATE TABLE IF NOT EXISTS lists (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL DEFAULT '',
		name TEXT NOT NULL,
		icon TEXT,
		color TEXT,
		order_num INTEGER NOT NULL DEFAULT 0,
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL
	)`,
	`CREATE INDEX IF NOT EXISTS idx_lists_user ON lists(user_id)`,
	`CREATE TABLE IF NOT EXISTS tasks (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL DEFAULT '',
		title TEXT NOT NULL,
		description TEXT,
		completed INTEGER NOT NULL DEFAULT 0,
		priority TEXT NOT NULL DEFAULT 'none',
		status TEXT NOT NULL DEFAULT 'todo',
		list_id TEXT NOT NULL,
		tags TEXT NOT NULL DEFAULT '[]',
		sub_tasks TEXT NOT NULL DEFAULT '[]',
		reminders TEXT NOT NULL DEFAULT '[]',
		due_date TEXT,
		deleted INTEGER NOT NULL DEFAULT 0,
		deleted_at TEXT,
		order_num INTEGER NOT NULL DEFAULT 0,
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL
	)`,
	`CREATE INDEX IF NOT EXISTS idx_tasks_user ON tasks(user_id)`,
	`CREATE INDEX IF NOT EXISTS idx_tasks_list ON tasks(list_id)`,
	`CREATE INDEX IF NOT EXISTS idx_tasks_updated ON tasks(updated_at)`,
	`CREATE TABLE IF NOT EXISTS sent_reminders (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL DEFAULT '',
		task_id TEXT NOT NULL,
		reminder_time INTEGER NOT NULL,
		sent_at INTEGER NOT NULL,
		reminder_data TEXT
	)`,
	`CREATE INDEX IF NOT EXISTS idx_sent_reminders_task ON sent_reminders(task_id)`,
	`CREATE TABLE IF NOT EXISTS sync_log (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		table_name TEXT NOT NULL,
		record_id TEXT NOT NULL,
		action TEXT NOT NULL,
		timestamp TEXT NOT NULL,
		synced INTEGER NOT NULL DEFAULT 0
	)`,
	`CREATE INDEX IF NOT EXISTS idx_sync_log_synced ON sync_log(synced)`,
}
```

- [ ] **Step 2: 编写迁移管理**

`internal/pkg/schema/migrate.go`:
```go
package schema

import (
	"database/sql"
	"fmt"
)

type Migration struct {
	Version int
	SQL     []string
}

var Migrations = []Migration{
	{Version: 1, SQL: CreateTablesSQL},
}

func RunMigrations(db *sql.DB) error {
	_, err := db.Exec("PRAGMA journal_mode=WAL")
	if err != nil {
		return fmt.Errorf("enable WAL: %w", err)
	}

	var currentVersion int
	err = db.QueryRow("PRAGMA user_version").Scan(&currentVersion)
	if err != nil {
		return fmt.Errorf("read version: %w", err)
	}

	for _, m := range Migrations {
		if m.Version <= currentVersion {
			continue
		}
		for _, s := range m.SQL {
			if _, err := db.Exec(s); err != nil {
				return fmt.Errorf("migration v%d: %w", m.Version, err)
			}
		}
		_, err := db.Exec(fmt.Sprintf("PRAGMA user_version = %d", m.Version))
		if err != nil {
			return fmt.Errorf("set version v%d: %w", m.Version, err)
		}
	}
	return nil
}
```

- [ ] **Step 3: 编写测试**

`internal/pkg/schema/schema_test.go`:
```go
package schema

import (
	"database/sql"
	"testing"
)

func TestRunMigrations(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	if err := RunMigrations(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	var version int
	db.QueryRow("PRAGMA user_version").Scan(&version)
	if version != CurrentVersion {
		t.Errorf("version = %d, want %d", version, CurrentVersion)
	}

	for _, table := range Tables {
		var count int
		db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&count)
		if count != 1 {
			t.Errorf("table %s not found", table)
		}
	}
}
```

- [ ] **Step 4: 运行测试**

```bash
go test ./internal/pkg/schema/ -v
```
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/pkg/schema/
git commit -m "feat: add shared schema definitions and migration system"
```

---

### Task 3: 配置管理

**Files:**
- Create: `internal/config/config.go`

**Interfaces:**
- Produces: `config.Load() (*Config, error)`, `config.Save(cfg *Config) error`, `Config` struct

- [ ] **Step 1: 编写配置结构体与加载**

`internal/config/config.go`:
```go
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	Port         int    `json:"port"`
	ServerURL    string `json:"server_url"`
	AuthToken    string `json:"auth_token"`
	AutoSync     bool   `json:"auto_sync"`
	DatabasePath string `json:"database_path"`
}

func Default() *Config {
	home, _ := os.UserHomeDir()
	return &Config{
		Port:         18080,
		AutoSync:     false,
		DatabasePath: filepath.Join(home, ".open-todo", "open-todo.db"),
	}
}

func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".open-todo")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return dir, nil
}

func configPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

func Load() (*Config, error) {
	cfg := Default()
	path, err := configPath()
	if err != nil {
		return cfg, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, err
	}
	json.Unmarshal(data, cfg)
	return cfg, nil
}

func Save(cfg *Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
```

- [ ] **Step 2: 运行基础编译验证**

```bash
go build ./internal/config/
```
Expected: 成功

- [ ] **Step 3: Commit**

```bash
git add internal/config/
git commit -m "feat: add config management (~/.open-todo/config.json)"
```

---

### Task 4: SQLite 数据库存储

**Files:**
- Create: `internal/store/store.go`
- Create: `internal/store/task_store.go`
- Create: `internal/store/list_store.go`
- Create: `internal/store/sync_store.go`
- Create: `internal/store/store_test.go`

**Interfaces:**
- Consumes: `schema.RunMigrations`, `config.Config.DatabasePath`
- Produces: `store.Open(path string) (*Store, error)`, `Store.GetTasks()`, `Store.SaveTask()`, `Store.DeleteTask()`, `Store.GetLists()`, `Store.SaveList()`, methods for `sync_log`, `Store.Close()`

- [ ] **Step 1: 编写 Store 结构体**

`internal/store/store.go`:
```go
package store

import (
	"database/sql"
	"fmt"
	"github.com/linfree/open-todo/internal/pkg/schema"
	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	db.SetMaxOpenConns(1)
	if err := schema.RunMigrations(db); err != nil {
		return nil, err
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) DB() *sql.DB {
	return s.db
}
```

- [ ] **Step 2: 编写任务存储**

`internal/store/task_store.go`:
```go
package store

import (
	"time"
	"github.com/google/uuid"
)

type Task struct {
	ID          string  `json:"id"`
	UserID      string  `json:"user_id"`
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

func (s *Store) GetTasks() ([]Task, error) {
	rows, err := s.db.Query("SELECT id, user_id, title, description, completed, priority, status, list_id, tags, sub_tasks, reminders, due_date, deleted, deleted_at, order_num, created_at, updated_at FROM tasks WHERE deleted = 0 ORDER BY order_num ASC, created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		var completed int
		var deleted int
		if err := rows.Scan(&t.ID, &t.UserID, &t.Title, &t.Description, &completed, &t.Priority, &t.Status, &t.ListID, &t.Tags, &t.SubTasks, &t.Reminders, &t.DueDate, &deleted, &t.DeletedAt, &t.OrderNum, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		t.Completed = completed != 0
		t.Deleted = deleted != 0
		tasks = append(tasks, t)
	}
	if tasks == nil {
		tasks = []Task{}
	}
	return tasks, rows.Err()
}

func (s *Store) SaveTask(t *Task) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	if t.CreatedAt == "" {
		t.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}
	t.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	completed := 0
	if t.Completed {
		completed = 1
	}
	deleted := 0
	if t.Deleted {
		deleted = 1
	}

	_, err := s.db.Exec(
		`INSERT OR REPLACE INTO tasks (id, user_id, title, description, completed, priority, status, list_id, tags, sub_tasks, reminders, due_date, deleted, deleted_at, order_num, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		t.ID, t.UserID, t.Title, t.Description, completed, t.Priority, t.Status, t.ListID, t.Tags, t.SubTasks, t.Reminders, t.DueDate, deleted, t.DeletedAt, t.OrderNum, t.CreatedAt, t.UpdatedAt,
	)
	if err != nil {
		return err
	}
	s.logChange("tasks", t.ID, "insert", t.UpdatedAt)
	return nil
}

func (s *Store) SoftDeleteTask(id string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.Exec("UPDATE tasks SET deleted = 1, deleted_at = ?, updated_at = ? WHERE id = ?", now, now, id)
	if err != nil {
		return err
	}
	s.logChange("tasks", id, "delete", now)
	return nil
}

func (s *Store) logChange(table, recordID, action, timestamp string) {
	s.db.Exec("INSERT INTO sync_log (table_name, record_id, action, timestamp, synced) VALUES (?, ?, ?, ?, 0)",
		table, recordID, action, timestamp)
}
```

- [ ] **Step 3: 编写清单存储**

`internal/store/list_store.go`:
```go
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
```

- [ ] **Step 4: 编写同步存储**

`internal/store/sync_store.go`:
```go
package store

type ChangeRecord struct {
	ID        int64  `json:"id"`
	TableName string `json:"table_name"`
	RecordID  string `json:"record_id"`
	Action    string `json:"action"`
	Timestamp string `json:"timestamp"`
	Synced    bool   `json:"synced"`
}

func (s *Store) GetUnsyncedChanges() ([]ChangeRecord, error) {
	rows, err := s.db.Query("SELECT id, table_name, record_id, action, timestamp FROM sync_log WHERE synced = 0 ORDER BY id ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var changes []ChangeRecord
	for rows.Next() {
		var c ChangeRecord
		var synced int
		if err := rows.Scan(&c.ID, &c.TableName, &c.RecordID, &c.Action, &c.Timestamp); err != nil {
			return nil, err
		}
		c.Synced = synced != 0
		changes = append(changes, c)
	}
	if changes == nil {
		changes = []ChangeRecord{}
	}
	return changes, rows.Err()
}

func (s *Store) MarkChangesSynced(ids []int64) error {
	for _, id := range ids {
		if _, err := s.db.Exec("UPDATE sync_log SET synced = 1 WHERE id = ?", id); err != nil {
			return err
		}
	}
	return nil
}
```

- [ ] **Step 5: 编写测试**

`internal/store/store_test.go`:
```go
package store

import (
	"testing"
)

func setupStore(t *testing.T) *Store {
	t.Helper()
	s, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	return s
}

func TestSaveAndGetTask(t *testing.T) {
	s := setupStore(t)
	defer s.Close()

	task := &Task{Title: "测试任务", ListID: "default", Priority: "medium"}
	if err := s.SaveTask(task); err != nil {
		t.Fatalf("save: %v", err)
	}

	tasks, err := s.GetTasks()
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if tasks[0].Title != "测试任务" {
		t.Errorf("title = %s, want 测试任务", tasks[0].Title)
	}
	if tasks[0].ID == "" {
		t.Error("task id is empty")
	}
}

func TestSoftDeleteTask(t *testing.T) {
	s := setupStore(t)
	defer s.Close()

	task := &Task{Title: "待删除", ListID: "default"}
	s.SaveTask(task)
	s.SoftDeleteTask(task.ID)

	tasks, _ := s.GetTasks()
	if len(tasks) != 0 {
		t.Errorf("expected 0 active tasks, got %d", len(tasks))
	}
}

func TestGetUnsyncedChanges(t *testing.T) {
	s := setupStore(t)
	defer s.Close()

	task := &Task{Title: "同步测试", ListID: "default"}
	s.SaveTask(task)

	changes, err := s.GetUnsyncedChanges()
	if err != nil {
		t.Fatalf("get changes: %v", err)
	}
	if len(changes) == 0 {
		t.Error("expected unsynced changes")
	}
	if changes[0].TableName != "tasks" {
		t.Errorf("table = %s, want tasks", changes[0].TableName)
	}
}
```

- [ ] **Step 6: 运行测试**

```bash
go test ./internal/store/ -v
```
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add internal/store/
git commit -m "feat: add SQLite store with tasks, lists, and sync_log support"
```

---

### Task 5: 默认数据初始化 + 提醒记录

**Files:**
- Create: `internal/store/init.go`
- Create: `internal/store/reminder_store.go`
- Modify: `internal/store/store_test.go` (add init test)

**Interfaces:**
- Consumes: `Store.GetLists()`, `Store.SaveList()`
- Produces: `Store.InitDefaults()`, `Store.IsReminderSent(taskID string, time int64) (bool, error)`, `Store.MarkReminderSent(id, taskID string, time int64, data string) error`

- [ ] **Step 1: 编写默认数据初始化**

`internal/store/init.go`:
```go
package store

import "time"

func (s *Store) InitDefaults() error {
	lists, err := s.GetLists()
	if err != nil {
		return err
	}
	if len(lists) > 0 {
		return nil
	}

	now := time.Now().UTC().Format(time.RFC3339)
	defaults := []TaskList{
		{ID: "all", Name: "全部", Icon: strPtr("Inbox"), OrderNum: 0, CreatedAt: now, UpdatedAt: now},
		{ID: "today", Name: "今天", Icon: strPtr("Sun"), OrderNum: 1, CreatedAt: now, UpdatedAt: now},
		{ID: "week", Name: "最近7天", Icon: strPtr("Calendar"), OrderNum: 2, CreatedAt: now, UpdatedAt: now},
	}
	for _, l := range defaults {
		ll := l
		if err := s.SaveList(&ll); err != nil {
			return err
		}
	}
	return nil
}

func strPtr(s string) *string { return &s }
```

- [ ] **Step 2: 编写提醒记录存储**

`internal/store/reminder_store.go`:
```go
package store

import (
	"github.com/google/uuid"
	"time"
)

func (s *Store) IsReminderSent(taskID string, reminderTime int64) (bool, error) {
	var count int
	err := s.db.QueryRow(
		"SELECT COUNT(*) FROM sent_reminders WHERE task_id = ? AND reminder_time = ?",
		taskID, reminderTime,
	).Scan(&count)
	return count > 0, err
}

func (s *Store) MarkReminderSent(id, taskID string, reminderTime int64, data string) error {
	if id == "" {
		id = uuid.New().String()
	}
	sentAt := time.Now().Unix()
	_, err := s.db.Exec(
		"INSERT OR REPLACE INTO sent_reminders (id, task_id, reminder_time, sent_at, reminder_data) VALUES (?, ?, ?, ?, ?)",
		id, taskID, reminderTime, sentAt, data,
	)
	return err
}
```

- [ ] **Step 3: 扩展测试**

在 `internal/store/store_test.go` 末尾追加:
```go
func TestInitDefaults(t *testing.T) {
	s := setupStore(t)
	defer s.Close()

	s.InitDefaults()
	lists, _ := s.GetLists()
	if len(lists) != 3 {
		t.Errorf("expected 3 default lists, got %d", len(lists))
	}
}

func TestReminderSent(t *testing.T) {
	s := setupStore(t)
	defer s.Close()

	s.MarkReminderSent("", "task-1", 100, "{}")
	sent, _ := s.IsReminderSent("task-1", 100)
	if !sent {
		t.Error("expected reminder to be sent")
	}
	notSent, _ := s.IsReminderSent("task-1", 200)
	if notSent {
		t.Error("expected reminder not sent for different time")
	}
}
```

- [ ] **Step 4: 运行测试**

```bash
go test ./internal/store/ -v
```
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/store/
git commit -m "feat: add default list init and reminder tracking"
```

---

### Task 6: HTTP 服务器 + API 路由

**Files:**
- Create: `internal/server/server.go`
- Create: `internal/server/api/task.go`
- Create: `internal/server/api/list.go`
- Create: `internal/server/api/backup.go` (stub)

**Interfaces:**
- Consumes: `store.Store` methods
- Produces: `server.New(store *Store, cfg *Config) *gin.Engine`, API routes under `/api/v1/`

- [ ] **Step 1: 编写 Gin 服务器**

`internal/server/server.go`:
```go
package server

import (
	"github.com/gin-gonic/gin"
	"github.com/linfree/open-todo/internal/config"
	"github.com/linfree/open-todo/internal/server/api"
	"github.com/linfree/open-todo/internal/store"
)

func New(st *store.Store, cfg *config.Config) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(corsMiddleware())

	apiGroup := r.Group("/api/v1")
	api.RegisterTaskRoutes(apiGroup, st)
	api.RegisterListRoutes(apiGroup, st)

	return r
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
```

- [ ] **Step 2: 编写任务 API**

`internal/server/api/task.go`:
```go
package api

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/linfree/open-todo/internal/store"
)

func RegisterTaskRoutes(r *gin.RouterGroup, st *store.Store) {
	r.GET("/tasks", func(c *gin.Context) {
		tasks, err := st.GetTasks()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, tasks)
	})

	r.POST("/tasks", func(c *gin.Context) {
		var t store.Task
		if err := c.ShouldBindJSON(&t); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := st.SaveTask(&t); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, t)
	})

	r.DELETE("/tasks/:id", func(c *gin.Context) {
		id := c.Param("id")
		if err := st.SoftDeleteTask(id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
}
```

- [ ] **Step 3: 编写清单 API**

`internal/server/api/list.go`:
```go
package api

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/linfree/open-todo/internal/store"
)

func RegisterListRoutes(r *gin.RouterGroup, st *store.Store) {
	r.GET("/lists", func(c *gin.Context) {
		lists, err := st.GetLists()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, lists)
	})

	r.POST("/lists", func(c *gin.Context) {
		var l store.TaskList
		if err := c.ShouldBindJSON(&l); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := st.SaveList(&l); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, l)
	})

	r.DELETE("/lists/:id", func(c *gin.Context) {
		id := c.Param("id")
		if err := st.DeleteList(id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
}
```

- [ ] **Step 4: 编写备份 API stub**

`internal/server/api/backup.go`:
```go
package api

import (
	"github.com/gin-gonic/gin"
	"github.com/linfree/open-todo/internal/store"
)

func RegisterBackupRoutes(r *gin.RouterGroup, st *store.Store) {
	r.GET("/backup", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "backup not implemented yet"})
	})
}
```

- [ ] **Step 5: 验证编译 + HTTP 测试**

在 `internal/server/` 下创建 server_test.go:
```go
package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/linfree/open-todo/internal/config"
	"github.com/linfree/open-todo/internal/store"
)

func TestGetTasks(t *testing.T) {
	st, _ := store.Open(":memory:")
	defer st.Close()
	cfg := config.Default()
	router := New(st, cfg)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/tasks", nil)
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var tasks []map[string]interface{}
	json.NewDecoder(w.Body).Decode(&tasks)
}

func TestCreateTask(t *testing.T) {
	st, _ := store.Open(":memory:")
	defer st.Close()
	cfg := config.Default()
	router := New(st, cfg)

	body := strings.NewReader(`{"title":"API测试","list_id":"default"}`)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/tasks", body)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}
```
需要加入 `"strings"` import。

- [ ] **Step 6: 运行测试**

```bash
go test ./internal/server/ -v
```
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add internal/server/
git commit -m "feat: add Gin HTTP server with task/list API routes"
```

---

### Task 7: 前端嵌入 (go:embed) + 桌面入口

**Files:**
- Create: `cmd/open-todo/embed.go`
- Modify: `cmd/open-todo/main.go`
- Create: `cmd/open-todo/icons/` (placeholder)

**Interfaces:**
- Consumes: `server.New()`, `store.Open()`, `config.Load()`
- Produces: 完整桌面启动流程, 从 `web/dist/` 嵌入前端

- [ ] **Step 1: 编写 embed 和静态资源服务**

`cmd/open-todo/embed.go`:
```go
package main

import (
	"embed"
	"io/fs"
	"net/http"
	"github.com/gin-gonic/gin"
)

//go:embed web-dist/*
var webAssets embed.FS

func registerStaticRoutes(r *gin.Engine) {
	sub, err := fs.Sub(webAssets, "web-dist")
	if err != nil {
		// 开发模式: 没有 web-dist 目录, 跳过
		return
	}

	fileServer := http.FileServer(http.FS(sub))

	r.GET("/assets/*filepath", func(c *gin.Context) {
		c.Writer.Header().Set("Cache-Control", "public, max-age=31536000")
		fileServer.ServeHTTP(c.Writer, c.Request)
	})

	r.NoRoute(func(c *gin.Context) {
		data, err := webAssets.ReadFile("web-dist/index.html")
		if err != nil {
			c.String(http.StatusNotFound, "Frontend not built. Run: cd web && npm run build")
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	})
}
```

- [ ] **Step 2: 编写 main.go**

`cmd/open-todo/main.go`:
```go
package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"github.com/linfree/open-todo/internal/config"
	"github.com/linfree/open-todo/internal/server"
	"github.com/linfree/open-todo/internal/store"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	st, err := store.Open(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("open store: %v", err)
	}
	defer st.Close()

	if err := st.InitDefaults(); err != nil {
		log.Printf("init defaults: %v", err)
	}

	router := server.New(st, cfg)
	registerStaticRoutes(router)

	addr := fmt.Sprintf(":%d", cfg.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("listen %s: %v", addr, err)
	}

	log.Printf("open-todo server starting on http://localhost%s", addr)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		log.Println("shutting down...")
		listener.Close()
	}()

	if err := router.RunListener(listener); err != nil {
		log.Printf("server: %v", err)
	}
}
```

- [ ] **Step 3: 创建 web/dist 占位**

```bash
mkdir -p cmd/open-todo/web-dist
touch cmd/open-todo/web-dist/index.html
```

在 `cmd/open-todo/web-dist/index.html` 写最小占位:
```html
<!DOCTYPE html><html><body>open-todo frontend placeholder</body></html>
```

- [ ] **Step 4: 构建并验证**

```bash
cd web && npm run build
# 如果前端还没改, 跳过

go build ./cmd/open-todo/
./open-todo.exe &
sleep 2
curl http://localhost:18080/api/v1/tasks
# 期望返回 []
```

- [ ] **Step 5: 修改 vite 输出目录**

修改 `web/vite.config.ts` 的 build.outDir 为 `../cmd/open-todo/web-dist`:
```typescript
export default defineConfig({
  plugins: [react()],
  build: {
    outDir: '../cmd/open-todo/web-dist',
  },
  server: {
    port: 5173,
    proxy: {
      '/api': 'http://localhost:18080',
    },
  },
})
```

- [ ] **Step 6: Commit**

```bash
git add cmd/open-todo/ go.mod go.sum web/vite.config.ts
git commit -m "feat: add Go desktop entry point with embedded frontend support"
```

---

### Task 8: 桌面 UI (Systray + WebView2)

**Files:**
- Create: `internal/ui/ui.go`
- Create: `internal/ui/ui_windows.go`
- Create: `internal/ui/ui_darwin.go`
- Create: `internal/ui/ui_linux.go`
- Modify: `cmd/open-todo/main.go` (集成 UI)

**Interfaces:**
- Consumes: 端口号
- Produces: `ui.New(port int) UI`, `UI.Run(onReady func())`

- [ ] **Step 1: 编写 UI 接口 + 通用方法**

`internal/ui/ui.go`:
```go
package ui

import (
	"fmt"
	"os/exec"
	"runtime"
)

type UI interface {
	Run(onReady func())
	Quit()
}

func New(port int) UI {
	return newUI(port)
}

func openBrowser(url string) {
	switch runtime.GOOS {
	case "darwin":
		exec.Command("open", url).Start()
	case "linux":
		exec.Command("xdg-open", url).Start()
	case "windows":
		exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	}
}

func localURL(port int) string {
	return fmt.Sprintf("http://localhost:%d", port)
}
```

- [ ] **Step 2: Windows UI (WebView2 + Systray)**

`internal/ui/ui_windows.go`:
```go
//go:build windows

package ui

import (
	"fmt"
	"log"
	"github.com/getlantern/systray"
	"github.com/jchv/go-webview2"
)

type windowsUI struct {
	port   int
	webview *webview2.WebView
}

func newUI(port int) UI {
	return &windowsUI{port: port}
}

func (u *windowsUI) Run(onReady func()) {
	systray.Run(func() {
		systray.SetTitle("Open Todo")
		systray.SetTooltip("Open Todo - 开源待办清单")
		onReady()

		mShow := systray.AddMenuItem("显示窗口", "显示主窗口")
		mQuit := systray.AddMenuItem("退出", "退出应用")

		go func() {
			u.webview = webview2.New(false)
			defer u.webview.Destroy()
			u.webview.SetTitle("Open Todo")
			u.webview.SetSize(800, 600, webview2.HintNone)
			u.webview.Navigate(localURL(u.port))
			u.webview.Run()
		}()

		go func() {
			for {
				select {
				case <-mShow.ClickedCh:
					if u.webview != nil {
						u.webview.Show()
					}
				case <-mQuit.ClickedCh:
					systray.Quit()
					return
				}
			}
		}()
	}, func() {
		if u.webview != nil {
			u.webview.Destroy()
		}
	})
}

func (u *windowsUI) Quit() {
	systray.Quit()
}
```

- [ ] **Step 3: macOS UI (Systray + Browser)**

`internal/ui/ui_darwin.go`:
```go
//go:build darwin

package ui

import (
	"github.com/getlantern/systray"
)

type darwinUI struct {
	port int
}

func newUI(port int) UI {
	return &darwinUI{port: port}
}

func (u *darwinUI) Run(onReady func()) {
	systray.Run(func() {
		systray.SetTitle("Open Todo")
		onReady()

		mOpen := systray.AddMenuItem("在浏览器中打开", "打开 Web 界面")
		mQuit := systray.AddMenuItem("退出", "退出应用")

		go func() {
			openBrowser(localURL(u.port))
		}()

		go func() {
			for {
				select {
				case <-mOpen.ClickedCh:
					openBrowser(localURL(u.port))
				case <-mQuit.ClickedCh:
					systray.Quit()
					return
				}
			}
		}()
	}, func() {})
}

func (u *darwinUI) Quit() {
	systray.Quit()
}
```

- [ ] **Step 4: Linux UI (Systray + Browser)**

`internal/ui/ui_linux.go`:
```go
//go:build linux

package ui

import (
	"github.com/getlantern/systray"
)

type linuxUI struct {
	port int
}

func newUI(port int) UI {
	return &linuxUI{port: port}
}

func (u *linuxUI) Run(onReady func()) {
	systray.Run(func() {
		systray.SetTitle("Open Todo")
		onReady()

		mOpen := systray.AddMenuItem("在浏览器中打开", "打开 Web 界面")
		mQuit := systray.AddMenuItem("退出", "退出应用")

		go func() {
			openBrowser(localURL(u.port))
		}()

		go func() {
			for {
				select {
				case <-mOpen.ClickedCh:
					openBrowser(localURL(u.port))
				case <-mQuit.ClickedCh:
					systray.Quit()
					return
				}
			}
		}()
	}, func() {})
}

func (u *linuxUI) Quit() {
	systray.Quit()
}
```

- [ ] **Step 5: 更新 main.go 集成 UI**

修改 `cmd/open-todo/main.go`, 在启动服务器后添加:

```go
	// 在 signal.Notify 之后，router.RunListener 之前:
	go func() {
		<-sigCh
		log.Println("shutting down...")
		listener.Close()
	}()

	// 启动 UI
	uiApp := ui.New(cfg.Port)
	uiApp.Run(func() {
		log.Println("UI ready")
	})

	// router.RunListener 移到 goroutine
	go func() {
		if err := router.RunListener(listener); err != nil {
			log.Printf("server: %v", err)
		}
	}()

	select {} // 阻塞直到 UI 退出
```

- [ ] **Step 6: 验证编译(仅 Windows)**

```bash
go build ./cmd/open-todo/
```
Expected: 成功 (Windows), 其他平台需要交叉编译验证

- [ ] **Step 7: Commit**

```bash
git add internal/ui/ cmd/open-todo/main.go
git commit -m "feat: add cross-platform desktop UI (systray + WebView2/browser)"
```

---

### Task 9: 前端 DatabaseAdapter 接口 + GoDesktopAdapter

**Files:**
- Create: `web/src/lib/adapter/types.ts`
- Create: `web/src/lib/adapter/DatabaseAdapter.ts`
- Create: `web/src/lib/adapter/GoDesktopAdapter.ts`
- Create: `web/src/lib/adapter/factory.ts`
- Modify: `web/src/lib/api.ts` (转为 re-export adapter)

**Interfaces:**
- Produces: `DatabaseAdapter` 接口, `GoDesktopAdapter`, `createAdapter()` 工厂函数

- [ ] **Step 1: 定义接口和类型**

`web/src/lib/adapter/types.ts`:
```typescript
export interface Task {
  id: string;
  user_id?: string;
  title: string;
  description?: string;
  completed: boolean;
  priority: string;
  status: string;
  list_id: string;
  tags: string;
  sub_tasks: string;
  reminders: string;
  due_date?: string;
  deleted: boolean;
  deleted_at?: string;
  order_num: number;
  created_at: string;
  updated_at: string;
}

export interface TaskList {
  id: string;
  user_id?: string;
  name: string;
  icon?: string;
  color?: string;
  order_num: number;
  created_at: string;
  updated_at: string;
}

export interface ChangeRecord {
  id: number;
  table_name: string;
  record_id: string;
  action: string;
  timestamp: string;
  synced: boolean;
}

export interface DatabaseAdapter {
  init(): Promise<void>;
  getTasks(): Promise<Task[]>;
  saveTask(task: Task): Promise<Task>;
  deleteTask(id: string): Promise<void>;
  getLists(): Promise<TaskList[]>;
  saveList(list: TaskList): Promise<TaskList>;
  deleteList(id: string): Promise<void>;
}
```

- [ ] **Step 2: 定义接口**

`web/src/lib/adapter/DatabaseAdapter.ts`:
```typescript
export { type DatabaseAdapter, type Task, type TaskList, type ChangeRecord } from './types';
```

- [ ] **Step 3: GoDesktopAdapter 实现**

`web/src/lib/adapter/GoDesktopAdapter.ts`:
```typescript
import type { DatabaseAdapter, Task, TaskList, ChangeRecord } from './types';

const BASE = ''; // 同源, 用相对路径

export class GoDesktopAdapter implements DatabaseAdapter {
  async init(): Promise<void> {
    // Go 服务器在启动时已初始化数据库
  }

  async getTasks(): Promise<Task[]> {
    const res = await fetch(`${BASE}/api/v1/tasks`);
    if (!res.ok) throw new Error(`getTasks: ${res.status}`);
    return res.json();
  }

  async saveTask(task: Task): Promise<Task> {
    const res = await fetch(`${BASE}/api/v1/tasks`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(task),
    });
    if (!res.ok) throw new Error(`saveTask: ${res.status}`);
    return res.json();
  }

  async deleteTask(id: string): Promise<void> {
    const res = await fetch(`${BASE}/api/v1/tasks/${id}`, { method: 'DELETE' });
    if (!res.ok) throw new Error(`deleteTask: ${res.status}`);
  }

  async getLists(): Promise<TaskList[]> {
    const res = await fetch(`${BASE}/api/v1/lists`);
    if (!res.ok) throw new Error(`getLists: ${res.status}`);
    return res.json();
  }

  async saveList(list: TaskList): Promise<TaskList> {
    const res = await fetch(`${BASE}/api/v1/lists`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(list),
    });
    if (!res.ok) throw new Error(`saveList: ${res.status}`);
    return res.json();
  }

  async deleteList(id: string): Promise<void> {
    const res = await fetch(`${BASE}/api/v1/lists/${id}`, { method: 'DELETE' });
    if (!res.ok) throw new Error(`deleteList: ${res.status}`);
  }
}
```

- [ ] **Step 4: 工厂函数**

`web/src/lib/adapter/factory.ts`:
```typescript
import type { DatabaseAdapter } from './types';
import { GoDesktopAdapter } from './GoDesktopAdapter';

let cached: DatabaseAdapter | null = null;

export function createAdapter(): DatabaseAdapter {
  if (cached) return cached;

  // 检测是否在 Go 桌面环境 (访问同一来源的 Gin 服务器)
  // PWA 模式下不走这个 adapter
  cached = new GoDesktopAdapter();
  return cached;
}

export function resetAdapter(): void {
  cached = null;
}
```

- [ ] **Step 5: 更新 api.ts**

修改 `web/src/lib/api.ts`, 删除 `isTauri()` 相关代码, 改为:

```typescript
import { createAdapter } from './adapter/factory';
import type { DatabaseAdapter, Task, TaskList } from './adapter/types';

const adapter: DatabaseAdapter = createAdapter();

export const databaseApi = {
  async initDatabase(): Promise<void> { return adapter.init(); },
  async getTasks(): Promise<Task[]> { return adapter.getTasks(); },
  async saveTask(task: Task): Promise<Task> { return adapter.saveTask(task); },
  async deleteTask(id: string): Promise<void> { return adapter.deleteTask(id); },
  async getLists(): Promise<TaskList[]> { return adapter.getLists(); },
};
```

- [ ] **Step 6: 验证编译**

```bash
cd web && npx tsc --noEmit
```
Expected: 无类型错误 (可能需要调整现有代码中的导入)

- [ ] **Step 7: Commit**

```bash
git add web/src/lib/adapter/ web/src/lib/api.ts
git commit -m "feat: add DatabaseAdapter interface and GoDesktopAdapter"
```

---

### Task 10: 更新 Zustand Store 适配新 Adapter

**Files:**
- Modify: `web/src/store/todoStore.ts`
- Modify: `web/src/types/index.ts` (对齐新的字段命名)

**Interfaces:**
- Consumes: `databaseApi`, new Task type with `list_id`/`order_num` 等 snake_case 字段

- [ ] **Step 1: 更新 types/index.ts**

task 类型的 `listId` → `list_id`, `order` → `order_num`, `dueDate` → `due_date` 等。为保持向后兼容，保留 TypeScript 类型但映射 API 返回的 snake_case。

`web/src/types/index.ts` 中 Task 接口的字段名保持不变 (listId, order, dueDate), 在 adapter 层做 camelCase↔snake_case 转换。

修改 `web/src/lib/adapter/GoDesktopAdapter.ts`，在 `getTasks()` 中做映射:

```typescript
async getTasks(): Promise<Task[]> {
  const res = await fetch(`${BASE}/api/v1/tasks`);
  if (!res.ok) throw new Error(`getTasks: ${res.status}`);
  const data = await res.json();
  return data.map((t: any) => ({
    ...t,
    listId: t.list_id,
    orderNum: t.order_num,
    dueDate: t.due_date,
    createdAt: t.created_at,
    updatedAt: t.updated_at,
    subTasks: JSON.parse(t.sub_tasks || '[]'),
    tags: JSON.parse(t.tags || '[]'),
    reminders: JSON.parse(t.reminders || '[]'),
  }));
}
```

`saveTask()` 做反向映射:

```typescript
async saveTask(task: Task): Promise<Task> {
  const body = {
    ...task,
    list_id: task.listId,
    order_num: task.orderNum,
    due_date: task.dueDate,
    created_at: task.createdAt,
    updated_at: task.updatedAt,
    sub_tasks: JSON.stringify(task.subTasks || []),
    tags: JSON.stringify(task.tags || []),
    reminders: JSON.stringify(task.reminders || []),
  };
  // ...
}
```

- [ ] **Step 2: 修改 todoStore.ts 异步保存部分**

`todoStore.ts` 中 `addTask` 和 `updateTask` 的异步后端保存改为使用新的 `databaseApi.saveTask()`:

当前 `databaseApi.saveTask()` 已经是 async 的, 保持调用方式不变即可。只需确保类型匹配。

- [ ] **Step 3: 验证**

```bash
cd web && npx tsc --noEmit
```
Expected: 无类型错误

- [ ] **Step 4: Commit**

```bash
git add web/src/
git commit -m "feat: update store and types for adapter camelCase/snake_case mapping"
```

---

### Task 11: 整理项目，移除 Tauri/Rust/Android 代码

**Files:**
- Delete: `src-tauri/` (整个目录)
- Delete: `build_windows.bat`, `build_windows.ps1`, `build_rust.ps1`
- Delete: `check_cargo.ps1`, `check_cargo_postgres.ps1`, `list_packages.ps1`
- Delete: `package.json` 中的 `"tauri"` 依赖和 script
- Modify: `vite.config.ts` (移除 Tauri 相关配置)
- Delete: `src/lib/api.ts` 中旧的 `invoke` 导入
- Modify: 组件中 `isTauri()` 引用 → 移除

**注意:** 此任务为清理性任务, 没有新功能产出。

- [ ] **Step 1: 移除 Tauri 后端**

```bash
rm -rf src-tauri/
```

- [ ] **Step 2: 移除构建脚本**

```bash
rm -f build_windows.bat build_windows.ps1 build_rust.ps1
rm -f check_cargo.ps1 check_cargo_postgres.ps1 list_packages.ps1
rm -f *.bat
```

- [ ] **Step 3: 清理 package.json**

```bash
cd web
npm uninstall @tauri-apps/cli @tauri-apps/api @tauri-apps/plugin-opener
```

- [ ] **Step 4: 清理 vite.config.ts**

移除 Tauri server 配置:

```typescript
import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  build: {
    outDir: '../cmd/open-todo/web-dist',
  },
  server: {
    port: 5173,
    proxy: {
      '/api': 'http://localhost:18080',
    },
  },
});
```

- [ ] **Step 5: 清理前端中的 isTauri 引用**

搜索 `isTauri`:
```bash
cd web && grep -r "isTauri" src/ --include="*.ts" --include="*.tsx"
```

对于 NotificationSettings.tsx、DataSettings.tsx 中的 `isTauri()`, 改为检查 adapter 能力或直接移除条件判断 (Go 桌面始终有完整能力)。

- [ ] **Step 6: 验证**

```bash
cd web && npx tsc --noEmit
go build ./cmd/open-todo/
```
Expected: 编译通过

- [ ] **Step 7: Commit**

```bash
git add -A .
git commit -m "feat: remove Tauri/Rust/Android, clean up old build scripts"
```

---

### Task 12: WebDAV 备份 (Go 迁移)

**Files:**
- Create: `internal/app/backup.go`
- Create: `internal/app/reminder.go` (stub)
- Modify: `internal/server/api/backup.go`
- Modify: `internal/store/store.go` (新增备份相关方法)

**Interfaces:**
- Consumes: `store.Store`
- Produces: 备份到 WebDAV, 从 WebDAV 恢复

- [ ] **Step 1: 编写 WebDAV 备份逻辑**

`internal/app/backup.go`:
```go
package app

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
	"github.com/linfree/open-todo/internal/store"
)

type WebDAVConfig struct {
	URL      string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func BackupToWebDAV(st *store.Store, cfg WebDAVConfig) (string, error) {
	tasks, _ := st.GetTasks()
	lists, _ := st.GetLists()

	data := map[string]interface{}{
		"tasks":     tasks,
		"lists":     lists,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	jsonData, _ := json.MarshalIndent(data, "", "  ")

	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	f, _ := w.Create("open-todo-backup.json")
	f.Write(jsonData)
	w.Close()

	filename := fmt.Sprintf("open-todo-%s.zip", time.Now().UTC().Format("20060102-150405"))
	url := fmt.Sprintf("%s/%s", cfg.URL, filename)

	req, err := http.NewRequest("PUT", url, &buf)
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(cfg.Username, cfg.Password)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("webdav error: %s", resp.Status)
	}

	return filename, nil
}

func RestoreFromWebDAV(st *store.Store, cfg WebDAVConfig, filename string) error {
	url := fmt.Sprintf("%s/%s", cfg.URL, filename)
	req, _ := http.NewRequest("GET", url, nil)
	req.SetBasicAuth(cfg.Username, cfg.Password)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	zr, _ := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))

	for _, f := range zr.File {
		if f.Name == "open-todo-backup.json" {
			rc, _ := f.Open()
			defer rc.Close()
			data, _ := io.ReadAll(rc)

			var backup struct {
				Tasks []store.Task     `json:"tasks"`
				Lists []store.TaskList `json:"lists"`
			}
			json.Unmarshal(data, &backup)

			for _, t := range backup.Tasks {
				tt := t
				st.SaveTask(&tt)
			}
			for _, l := range backup.Lists {
				ll := l
				st.SaveList(&ll)
			}
		}
	}
	return nil
}
```

- [ ] **Step 2: 编写提醒 stub**

`internal/app/reminder.go`:
```go
package app

import (
	"log"
	"time"
	"github.com/linfree/open-todo/internal/store"
)

func StartReminderChecker(st *store.Store, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			// TODO: Phase 4 完整实现提醒逻辑
			log.Println("[reminder] check running...")
			_ = st
		}
	}()
}
```

- [ ] **Step 3: 更新 backup API handler**

`internal/server/api/backup.go`:
```go
package api

import (
	"github.com/gin-gonic/gin"
	"github.com/linfree/open-todo/internal/app"
	"github.com/linfree/open-todo/internal/store"
)

func RegisterBackupRoutes(r *gin.RouterGroup, st *store.Store) {
	r.POST("/backup", func(c *gin.Context) {
		var cfg app.WebDAVConfig
		if err := c.ShouldBindJSON(&cfg); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		filename, err := app.BackupToWebDAV(st, cfg)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"filename": filename})
	})

	r.POST("/restore", func(c *gin.Context) {
		var req struct {
			Config   app.WebDAVConfig `json:"config"`
			Filename string           `json:"filename"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		if err := app.RestoreFromWebDAV(st, req.Config, req.Filename); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"ok": true})
	})
}
```

- [ ] **Step 4: 验证编译**

```bash
go build ./internal/app/
go build ./internal/server/api/
```
Expected: 成功

- [ ] **Step 5: Commit**

```bash
git add internal/app/ internal/server/api/backup.go
git commit -m "feat: migrate WebDAV backup/restore from Rust to Go"
```

---

### Task 13: PWA 配置 (vite-plugin-pwa + Manifest)

**Files:**
- Create: `web/public/manifest.json`
- Create: `web/public/icon-192.png` (logo 复制)
- Create: `web/public/icon-512.png` (logo 复制)
- Modify: `web/vite.config.ts`
- Modify: `web/package.json`

**Interfaces:**
- Produces: PWA manifest, Service Worker 自动生成, 可安装 PWA

- [ ] **Step 1: 安装 PWA 插件**

```bash
cd web
npm install -D vite-plugin-pwa
```

- [ ] **Step 2: 创建 manifest.json**

`web/public/manifest.json`:
```json
{
  "name": "Open Todo",
  "short_name": "OpenTodo",
  "description": "开源待办清单，数据由你掌控",
  "start_url": "/",
  "display": "standalone",
  "background_color": "#ffffff",
  "theme_color": "#3b82f6",
  "icons": [
    { "src": "/icon-192.png", "sizes": "192x192", "type": "image/png" },
    { "src": "/icon-512.png", "sizes": "512x512", "type": "image/png" }
  ]
}
```

- [ ] **Step 3: 复制图标**

```bash
cp public/logo.png public/icon-192.png
cp public/logo.png public/icon-512.png
```

- [ ] **Step 4: 更新 vite.config.ts**

```typescript
import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import { VitePWA } from "vite-plugin-pwa";

export default defineConfig({
  plugins: [
    react(),
    VitePWA({
      registerType: 'autoUpdate',
      includeAssets: ['icon-192.png', 'icon-512.png'],
      manifest: {
        name: 'Open Todo',
        short_name: 'OpenTodo',
        description: '开源待办清单，数据由你掌控',
        theme_color: '#3b82f6',
        background_color: '#ffffff',
        display: 'standalone',
        icons: [
          { src: '/icon-192.png', sizes: '192x192', type: 'image/png' },
          { src: '/icon-512.png', sizes: '512x512', type: 'image/png' },
        ],
      },
      workbox: {
        runtimeCaching: [
          {
            urlPattern: /\/api\/v1\/.*/,
            handler: 'NetworkFirst',
            options: {
              cacheName: 'api-cache',
              expiration: { maxEntries: 100, maxAgeSeconds: 60 * 60 },
            },
          },
          {
            urlPattern: /\.(js|css|png|jpg|svg|ico|woff2?)$/,
            handler: 'CacheFirst',
            options: {
              cacheName: 'static-cache',
              expiration: { maxEntries: 200, maxAgeSeconds: 30 * 24 * 3600 },
            },
          },
        ],
      },
    }),
  ],
  build: {
    outDir: '../cmd/open-todo/web-dist',
  },
  server: {
    port: 5173,
    proxy: {
      '/api': 'http://localhost:18080',
    },
  },
});
```

- [ ] **Step 5: 添加 PWA 构建脚本**

修改 `web/package.json`:
```json
"scripts": {
  "dev": "vite",
  "build": "tsc && vite build",
  "build:pwa": "tsc && vite build --mode pwa",
  "preview": "vite preview"
}
```

- [ ] **Step 6: 验证构建**

```bash
cd web && npm run build
ls dist/  # PWA 模式下在 dist/ 目录
```
Expected: 生成 PWA 文件, 包括 sw.js, manifest.json

- [ ] **Step 7: Commit**

```bash
git add web/
git commit -m "feat: add PWA configuration with vite-plugin-pwa"
```

---

### Task 14: DexieAdapter (PWA 本地存储)

**Files:**
- Create: `web/src/lib/adapter/DexieAdapter.ts`
- Modify: `web/src/lib/adapter/factory.ts` (添加环境检测)

**Interfaces:**
- Consumes: `DatabaseAdapter` 接口
- Produces: IndexedDB 本地存储实现

- [ ] **Step 1: 安装 Dexie**

```bash
cd web
npm install dexie
```

- [ ] **Step 2: 编写 DexieAdapter**

`web/src/lib/adapter/DexieAdapter.ts`:
```typescript
import Dexie, { type Table } from 'dexie';
import type { DatabaseAdapter, Task, TaskList, ChangeRecord } from './types';
import { v4 as uuid } from './uuid';

class TodoDB extends Dexie {
  tasks!: Table<Task>;
  lists!: Table<TaskList>;
  syncLog!: Table<ChangeRecord>;

  constructor() {
    super('open-todo');
    this.version(1).stores({
      tasks: 'id, user_id, list_id, updated_at, [user_id+list_id]',
      lists: 'id, user_id, updated_at',
      syncLog: '++id, table_name, record_id, synced, timestamp',
    });
  }
}

const db = new TodoDB();

function now(): string {
  return new Date().toISOString();
}

export class DexieAdapter implements DatabaseAdapter {
  async init(): Promise<void> {
    const lists = await db.lists.count();
    if (lists === 0) {
      const ts = now();
      await db.lists.bulkPut([
        { id: 'all', name: '全部', icon: 'Inbox', order_num: 0, created_at: ts, updated_at: ts },
        { id: 'today', name: '今天', icon: 'Sun', order_num: 1, created_at: ts, updated_at: ts },
        { id: 'week', name: '最近7天', icon: 'Calendar', order_num: 2, created_at: ts, updated_at: ts },
      ]);
    }
  }

  async getTasks(): Promise<Task[]> {
    return db.tasks.filter(t => !t.deleted).toArray();
  }

  async saveTask(task: Task): Promise<Task> {
    const t = {
      ...task,
      id: task.id || uuid(),
      created_at: task.created_at || now(),
      updated_at: now(),
    };
    await db.tasks.put(t);
    await db.syncLog.put({
      table_name: 'tasks',
      record_id: t.id,
      action: 'insert',
      timestamp: t.updated_at,
      synced: false,
    } as ChangeRecord);
    return t;
  }

  async deleteTask(id: string): Promise<void> {
    const ts = now();
    await db.tasks.update(id, { deleted: true, deleted_at: ts, updated_at: ts });
    await db.syncLog.put({
      table_name: 'tasks',
      record_id: id,
      action: 'delete',
      timestamp: ts,
      synced: false,
    } as ChangeRecord);
  }

  async getLists(): Promise<TaskList[]> {
    return db.lists.toArray();
  }

  async saveList(list: TaskList): Promise<TaskList> {
    const l = {
      ...list,
      id: list.id || uuid(),
      created_at: list.created_at || now(),
      updated_at: now(),
    };
    await db.lists.put(l);
    return l;
  }

  async deleteList(id: string): Promise<void> {
    await db.lists.delete(id);
  }
}
```

- [ ] **Step 3: UUID 辅助**

`web/src/lib/adapter/uuid.ts`:
```typescript
export function v4(): string {
  return crypto.randomUUID();
}
```

- [ ] **Step 4: 更新工厂函数**

修改 `web/src/lib/adapter/factory.ts`:

```typescript
import type { DatabaseAdapter } from './types';
import { GoDesktopAdapter } from './GoDesktopAdapter';
import { DexieAdapter } from './DexieAdapter';

let cached: DatabaseAdapter | null = null;

function isGoDesktop(): boolean {
  // 检测是否连接了 Go Gin 后端
  // 在桌面环境下, Go Gin 服务运行在 localhost
  const url = new URL(window.location.href);
  return url.hostname === 'localhost' || url.hostname === '127.0.0.1';
}

export function createAdapter(): DatabaseAdapter {
  if (cached) return cached;

  if (isGoDesktop()) {
    cached = new GoDesktopAdapter();
  } else {
    cached = new DexieAdapter();
  }
  return cached;
}

export function resetAdapter(): void {
  cached = null;
}
```

- [ ] **Step 5: 验证编译**

```bash
cd web && npx tsc --noEmit
```
Expected: 无错误

- [ ] **Step 6: Commit**

```bash
git add web/src/lib/adapter/
git commit -m "feat: add DexieAdapter for PWA offline local storage"
```

---

### Task 15: Go 服务端基础

**Files:**
- Create: `server/go.mod`
- Create: `server/cmd/server/main.go`
- Create: `server/internal/config/config.go`
- Create: `server/internal/database/database.go`
- Create: `server/internal/database/sqlite.go`
- Create: `server/internal/model/user.go`
- Create: `server/internal/handler/auth.go`
- Create: `server/internal/handler/sync.go`
- Create: `server/internal/handler/health.go`
- Create: `server/internal/middleware/auth.go`
- Create: `server/internal/router/router.go`

**Interfaces:**
- Produces: 独立 Go 服务端, JWT 认证 + 同步 API

- [ ] **Step 1: 初始化 server module**

```bash
mkdir -p server/cmd/server
cd server
go mod init github.com/linfree/open-todo/server
go get github.com/gin-gonic/gin
go get golang.org/x/crypto
go get github.com/golang-jwt/jwt/v5
go get modernc.org/sqlite
go get github.com/google/uuid
```

- [ ] **Step 2: 编写配置**

`server/internal/config/config.go`:
```go
package config

import "os"

type Config struct {
	Port       string
	Driver     string
	DSN        string
	JWTSecret  string
}

func Load() *Config {
	return &Config{
		Port:      getEnv("PORT", "8080"),
		Driver:    getEnv("DB_DRIVER", "sqlite"),
		DSN:       getEnv("DB_DSN", "./open-todo-server.db"),
		JWTSecret: getEnv("JWT_SECRET", "change-me-in-production"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
```

- [ ] **Step 3: 编写 Database 接口 + SQLite 实现**

`server/internal/database/database.go`:
```go
package database

import (
	"database/sql"
	"fmt"
	"time"
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
		// 同步数据的表 (与客户端结构一致)
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
```

- [ ] **Step 4: 编写 Auth handler**

`server/internal/handler/auth.go`:
```go
package handler

import (
	"net/http"
	"time"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/linfree/open-todo/server/internal/database"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	DB        *database.DB
	JWTSecret string
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required"`
		Name     string `json:"name"`
		Password string `json:"password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	now := time.Now().UTC().Format(time.RFC3339)
	user := &database.User{
		ID: uuid.New().String(), Email: req.Email, Name: req.Name,
		Password: string(hash), CreatedAt: now, UpdatedAt: now,
	}

	_, err := h.DB.CreateUser(user)
	if err != nil {
		c.JSON(409, gin.H{"error": "user already exists"})
		return
	}

	token, _ := h.generateToken(user.ID)
	c.JSON(201, gin.H{"user": user, "token": token})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	user, err := h.DB.GetUserByEmail(req.Email)
	if err != nil {
		c.JSON(401, gin.H{"error": "invalid credentials"})
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)) != nil {
		c.JSON(401, gin.H{"error": "invalid credentials"})
		return
	}

	token, _ := h.generateToken(user.ID)
	c.JSON(200, gin.H{"user": user, "token": token})
}

func (h *AuthHandler) generateToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(72 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.JWTSecret))
}
```

需要在 `server/internal/database/` 添加 User CRUD 方法:

`server/internal/database/user.go`:
```go
package database

func (d *DB) CreateUser(u *User) (*User, error) {
	_, err := d.db.Exec(
		"INSERT INTO users (id, email, name, password, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
		u.ID, u.Email, u.Name, u.Password, u.CreatedAt, u.UpdatedAt,
	)
	return u, err
}

func (d *DB) GetUserByEmail(email string) (*User, error) {
	var u User
	err := d.db.QueryRow(
		"SELECT id, email, name, password, created_at, updated_at FROM users WHERE email = ?", email,
	).Scan(&u.ID, &u.Email, &u.Name, &u.Password, &u.CreatedAt, &u.UpdatedAt)
	return &u, err
}

func (d *DB) GetUserByID(id string) (*User, error) {
	var u User
	err := d.db.QueryRow(
		"SELECT id, email, name, password, created_at, updated_at FROM users WHERE id = ?", id,
	).Scan(&u.ID, &u.Email, &u.Name, &u.Password, &u.CreatedAt, &u.UpdatedAt)
	return &u, err
}
```

- [ ] **Step 5: 编写 Sync handler**

`server/internal/handler/sync.go`:
```go
package handler

import (
	"net/http"
	"time"
	"github.com/gin-gonic/gin"
	"github.com/linfree/open-todo/server/internal/database"
)

type SyncHandler struct {
	DB *database.DB
}

func (h *SyncHandler) Push(c *gin.Context) {
	userID := c.GetString("user_id")
	var req struct {
		Changes []database.SyncRecord `json:"changes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if err := h.DB.StoreChanges(userID, req.Changes); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"server_time": time.Now().UTC().Format(time.RFC3339)})
}

func (h *SyncHandler) Pull(c *gin.Context) {
	userID := c.GetString("user_id")
	var req struct {
		Since string `json:"since"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	since, _ := time.Parse(time.RFC3339, req.Since)
	if req.Since == "" {
		since = time.Unix(0, 0)
	}

	changes, err := h.DB.GetChangesSince(userID, since)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"changes": changes, "server_time": time.Now().UTC().Format(time.RFC3339)})
}
```

- [ ] **Step 6: 编写 Sync 存储逻辑**

`server/internal/database/sync.go`:
```go
package database

import (
	"fmt"
	"time"
)

func (d *DB) StoreChanges(userID string, changes []SyncRecord) error {
	for _, c := range changes {
		switch c.TableName {
		case "tasks":
			// 直接存储到 tasks 表
			_, err := d.db.Exec(
				`INSERT OR REPLACE INTO tasks 
				(id, user_id, title, description, completed, priority, status, list_id,
				 tags, sub_tasks, reminders, due_date, deleted, deleted_at, order_num,
				 created_at, updated_at)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				c.RecordID, userID, "", "", 0, "none", "todo", "",
				"[]", "[]", "[]", nil, 0, nil, 0,
				time.Now().UTC().Format(time.RFC3339),
				time.Now().UTC().Format(time.RFC3339),
			)
			if err != nil {
				return fmt.Errorf("store task %s: %w", c.RecordID, err)
			}
		}
	}
	return nil
}

func (d *DB) GetChangesSince(userID string, since time.Time) ([]SyncRecord, error) {
	rows, err := d.db.Query(
		"SELECT 'tasks' as table_name, id, updated_at FROM tasks WHERE user_id = ? AND updated_at > ?",
		userID, since.UTC().Format(time.RFC3339),
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
```

- [ ] **Step 7: 编写中间件 + 路由 + 入口**

`server/internal/middleware/auth.go`:
```go
package middleware

import (
	"net/http"
	"strings"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthRequired(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(401, gin.H{"error": "missing token"})
			return
		}
		tokenStr := strings.TrimPrefix(header, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid token"})
			return
		}
		claims := token.Claims.(jwt.MapClaims)
		c.Set("user_id", claims["user_id"])
		c.Next()
	}
}
```

`server/internal/handler/health.go`:
```go
package handler

import "github.com/gin-gonic/gin"

func Health(c *gin.Context) {
	c.JSON(200, gin.H{"ok": true, "version": "0.1.0"})
}
```

`server/internal/router/router.go`:
```go
package router

import (
	"github.com/gin-gonic/gin"
	"github.com/linfree/open-todo/server/internal/database"
	"github.com/linfree/open-todo/server/internal/handler"
	"github.com/linfree/open-todo/server/internal/middleware"
)

func New(db *database.DB, jwtSecret string) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	authH := &handler.AuthHandler{DB: db, JWTSecret: jwtSecret}
	syncH := &handler.SyncHandler{DB: db}

	r.GET("/api/v1/health", handler.Health)

	auth := r.Group("/api/v1/auth")
	{
		auth.POST("/register", authH.Register)
		auth.POST("/login", authH.Login)
	}

	sync := r.Group("/api/v1/sync")
	sync.Use(middleware.AuthRequired(jwtSecret))
	{
		sync.POST("/push", syncH.Push)
		sync.POST("/pull", syncH.Pull)
	}

	return r
}
```

`server/cmd/server/main.go`:
```go
package main

import (
	"log"
	"github.com/linfree/open-todo/server/internal/config"
	"github.com/linfree/open-todo/server/internal/database"
	"github.com/linfree/open-todo/server/internal/router"
)

func main() {
	cfg := config.Load()
	db, err := database.Open(cfg.Driver, cfg.DSN)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer db.Close()

	r := router.New(db, cfg.JWTSecret)
	log.Printf("open-todo server starting on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server: %v", err)
	}
}
```

- [ ] **Step 8: 验证编译**

```bash
cd server && go build ./cmd/server/
```
Expected: 成功

- [ ] **Step 9: Commit**

```bash
git add server/
git commit -m "feat: add Go sync server with JWT auth and sync API"
```

---

### Task 16: SyncEngine (Go 桌面端)

**Files:**
- Create: `internal/sync/engine.go`
- Create: `internal/sync/client.go`
- Modify: `cmd/open-todo/main.go` (集成 SyncEngine)

**Interfaces:**
- Consumes: `store.Store`, `config.Config.ServerURL`
- Produces: 后台 goroutine 自动推拉同步

- [ ] **Step 1: 编写同步客户端**

`internal/sync/client.go`:
```go
package sync

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	ServerURL string
	Token     string
	http      *http.Client
}

func NewClient(serverURL, token string) *Client {
	return &Client{
		ServerURL: serverURL,
		Token:     token,
		http:      &http.Client{Timeout: 15 * time.Second},
	}
}

type SyncRecord struct {
	TableName string `json:"table_name"`
	RecordID  string `json:"record_id"`
	Action    string `json:"action"`
	Timestamp string `json:"timestamp"`
}

type PushRequest struct {
	Changes []SyncRecord `json:"changes"`
}

type PullRequest struct {
	Since string `json:"since"`
}

type PullResponse struct {
	Changes    []SyncRecord `json:"changes"`
	ServerTime string       `json:"server_time"`
}

func (c *Client) Push(changes []SyncRecord) error {
	body, _ := json.Marshal(PushRequest{Changes: changes})
	req, _ := http.NewRequest("POST", c.ServerURL+"/api/v1/sync/push", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("push failed: %s", resp.Status)
	}
	return nil
}

func (c *Client) Pull(since string) (*PullResponse, error) {
	body, _ := json.Marshal(PullRequest{Since: since})
	req, _ := http.NewRequest("POST", c.ServerURL+"/api/v1/sync/pull", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("pull failed: %s", resp.Status)
	}

	var result PullResponse
	json.NewDecoder(resp.Body).Decode(&result)
	return &result, nil
}
```

- [ ] **Step 2: 编写 SyncEngine**

`internal/sync/engine.go`:
```go
package sync

import (
	"log"
	"time"
	"github.com/linfree/open-todo/internal/store"
)

type Engine struct {
	store     *store.Store
	client    *Client
	lastSync  time.Time
	stopCh    chan struct{}
}

func NewEngine(st *store.Store, serverURL, token string) *Engine {
	return &Engine{
		store:   st,
		client:  NewClient(serverURL, token),
		lastSync: time.Now().UTC(),
		stopCh:  make(chan struct{}),
	}
}

func (e *Engine) Start() {
	log.Println("[sync] engine started")

	go e.pushLoop()
	go e.pullLoop()
}

func (e *Engine) Stop() {
	close(e.stopCh)
}

func (e *Engine) pushLoop() {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-e.stopCh:
			return
		case <-ticker.C:
			e.doPush()
		}
	}
}

func (e *Engine) pullLoop() {
	// 启动时立即拉取
	e.doPull()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-e.stopCh:
			return
		case <-ticker.C:
			e.doPull()
		}
	}
}

func (e *Engine) doPush() {
	changes, err := e.store.GetUnsyncedChanges()
	if err != nil || len(changes) == 0 {
		return
	}

	var records []SyncRecord
	var ids []int64
	for _, c := range changes {
		records = append(records, SyncRecord{
			TableName: c.TableName,
			RecordID:  c.RecordID,
			Action:    c.Action,
			Timestamp: c.Timestamp,
		})
		ids = append(ids, c.ID)
	}

	if err := e.client.Push(records); err != nil {
		log.Printf("[sync] push error: %v", err)
		return
	}

	e.store.MarkChangesSynced(ids)
	log.Printf("[sync] pushed %d changes", len(records))
}

func (e *Engine) doPull() {
	resp, err := e.client.Pull(e.lastSync.Format(time.RFC3339))
	if err != nil {
		log.Printf("[sync] pull error: %v", err)
		return
	}

	if len(resp.Changes) == 0 {
		return
	}

	for _, c := range resp.Changes {
		// TODO Phase 4: LWW merge into local DB
		log.Printf("[sync] pulled %s/%s", c.TableName, c.RecordID)
	}

	if t, err := time.Parse(time.RFC3339, resp.ServerTime); err == nil {
		e.lastSync = t
	}

	log.Printf("[sync] pulled %d changes", len(resp.Changes))
}
```

- [ ] **Step 3: 集成到 main.go (条件启动)**

在 `cmd/open-todo/main.go` 中, Store 和 config 初始化之后:

```go
	// 启动 SyncEngine (如果配置了服务端)
	if cfg.ServerURL != "" && cfg.AuthToken != "" {
		engine := sync.NewEngine(st, cfg.ServerURL, cfg.AuthToken)
		engine.Start()
		defer engine.Stop()
	}
```

需要添加 import: `"github.com/linfree/open-todo/internal/sync"`

- [ ] **Step 4: 验证编译**

```bash
go build ./internal/sync/
go build ./cmd/open-todo/
```
Expected: 成功

- [ ] **Step 5: Commit**

```bash
git add internal/sync/ cmd/open-todo/main.go
git commit -m "feat: add Go SyncEngine for desktop push/pull sync"
```

---

### Task 17: SyncEngine (PWA TypeScript)

**Files:**
- Create: `web/src/lib/sync/engine.ts`
- Create: `web/src/lib/sync/client.ts`
- Modify: `web/src/lib/adapter/factory.ts` (集成 SyncEngine 启动)

**Interfaces:**
- Consumes: `DatabaseAdapter`, 服务端 URL + token
- Produces: 后台同步

- [ ] **Step 1: 编写 TypeScript 同步客户端**

`web/src/lib/sync/client.ts`:
```typescript
export interface SyncRecord {
  table_name: string;
  record_id: string;
  action: string;
  timestamp: string;
}

interface PullResponse {
  changes: SyncRecord[];
  server_time: string;
}

export class SyncClient {
  constructor(private serverUrl: string, private token: string) {}

  async push(changes: SyncRecord[]): Promise<void> {
    const res = await fetch(`${this.serverUrl}/api/v1/sync/push`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${this.token}`,
      },
      body: JSON.stringify({ changes }),
    });
    if (!res.ok) throw new Error(`push: ${res.status}`);
  }

  async pull(since: string): Promise<PullResponse> {
    const res = await fetch(`${this.serverUrl}/api/v1/sync/pull`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${this.token}`,
      },
      body: JSON.stringify({ since }),
    });
    if (!res.ok) throw new Error(`pull: ${res.status}`);
    return res.json();
  }
}
```

- [ ] **Step 2: 编写 SyncEngine**

`web/src/lib/sync/engine.ts`:
```typescript
import { SyncClient, type SyncRecord } from './client';
import type { DatabaseAdapter } from '../adapter/types';

export class SyncEngine {
  private client: SyncClient;
  private lastSync: string;
  private pushTimer: number | null = null;
  private pullTimer: number | null = null;

  constructor(
    private adapter: DatabaseAdapter,
    serverUrl: string,
    token: string,
  ) {
    this.client = new SyncClient(serverUrl, token);
    this.lastSync = new Date().toISOString();
  }

  start(): void {
    // 启动时立即拉取
    this.doPull();

    // 每 3 秒推送
    this.pushTimer = window.setInterval(() => this.doPush(), 3000);
    // 每 30 秒拉取
    this.pullTimer = window.setInterval(() => this.doPull(), 30000);
  }

  stop(): void {
    if (this.pushTimer) clearInterval(this.pushTimer);
    if (this.pullTimer) clearInterval(this.pullTimer);
  }

  private async doPush(): Promise<void> {
    try {
      const changes = await (this.adapter as any).getUnsyncedChanges?.();
      if (!changes || changes.length === 0) return;

      const records: SyncRecord[] = changes.map((c: any) => ({
        table_name: c.table_name,
        record_id: c.record_id,
        action: c.action,
        timestamp: c.timestamp,
      }));

      await this.client.push(records);
      await (this.adapter as any).markChangesSynced?.(changes);
    } catch (err) {
      console.warn('[sync] push failed:', err);
    }
  }

  private async doPull(): Promise<void> {
    try {
      const resp = await this.client.pull(this.lastSync);
      this.lastSync = resp.server_time;
      if (resp.changes.length > 0) {
        // TODO Phase 4: LWW merge
        console.log('[sync] pulled', resp.changes.length, 'changes');
      }
    } catch (err) {
      console.warn('[sync] pull failed:', err);
    }
  }
}
```

- [ ] **Step 3: 验证编译**

```bash
cd web && npx tsc --noEmit
```
Expected: 无错误

- [ ] **Step 4: Commit**

```bash
git add web/src/lib/sync/
git commit -m "feat: add TypeScript SyncEngine for PWA background sync"
```

---

### Task 18: Service Worker Background Sync

**Files:**
- Create: `web/src/sw.ts` (Service Worker 入口)
- Modify: `web/vite.config.ts` (Workbox 配置)
- Modify: `web/src/main.tsx` (注册 SW)

**Interfaces:**
- Produces: 后台同步, 离线操作排队

- [ ] **Step 1: 编写 Service Worker**

`web/src/sw.ts`:
```typescript
/// <reference lib="webworker" />

import { clientsClaim } from 'workbox-core';
import { precacheAndRoute } from 'workbox-precaching';
import { registerRoute } from 'workbox-routing';
import { NetworkFirst, CacheFirst } from 'workbox-strategies';
import { ExpirationPlugin } from 'workbox-expiration';

declare let self: ServiceWorkerGlobalScope;

clientsClaim();
precacheAndRoute(self.__WB_MANIFEST);

// 同步事件
self.addEventListener('sync', (event: SyncEvent) => {
  if (event.tag === 'sync-changes') {
    event.waitUntil(performSync());
  }
});

async function performSync(): Promise<void> {
  const db = await openDB();
  const changes = await db.getAll('syncLog');
  const unsynced = changes.filter((c: any) => !c.synced);

  if (unsynced.length === 0) return;

  const token = await getToken();
  if (!token) return;

  try {
    const serverUrl = await getServerUrl();
    await fetch(`${serverUrl}/api/v1/sync/push`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
      body: JSON.stringify({ changes: unsynced }),
    });
    // 标记已同步
    const tx = db.transaction('syncLog', 'readwrite');
    for (const c of unsynced) {
      tx.store.put({ ...c, synced: true });
    }
    await tx.done;
  } catch (e) {
    console.warn('[sw] sync failed:', e);
  }
}

async function openDB(): Promise<IDBDatabase> {
  return new Promise((resolve, reject) => {
    const req = indexedDB.open('open-todo');
    req.onsuccess = () => resolve(req.result);
    req.onerror = () => reject(req.error);
  });
}

async function getToken(): Promise<string | null> {
  const db = await openDB();
  return new Promise((resolve) => {
    const tx = db.transaction('session', 'readonly');
    const req = tx.objectStore('session').get('jwt');
    req.onsuccess = () => resolve(req.result?.value ?? null);
  });
}

async function getServerUrl(): Promise<string> {
  const db = await openDB();
  return new Promise((resolve) => {
    const tx = db.transaction('session', 'readonly');
    const req = tx.objectStore('session').get('server_url');
    req.onsuccess = () => resolve(req.result?.value ?? '');
  });
}
```

- [ ] **Step 2: 在主入口注册 SW**

修改 `web/src/main.tsx`:
```typescript
// 注册 Service Worker
if ('serviceWorker' in navigator) {
  window.addEventListener('load', () => {
    navigator.serviceWorker.register('/sw.js').catch((err) => {
      console.warn('SW registration failed:', err);
    });
  });
}
```

- [ ] **Step 3: 验证构建**

```bash
cd web && npm run build
```
Expected: 生成 PWA 文件 + sw.js

- [ ] **Step 4: Commit**

```bash
git add web/src/sw.ts web/src/main.tsx web/vite.config.ts
git commit -m "feat: add Service Worker with background sync support"
```

---

### Task 19: 响应式 UI + 触控优化

**Files:**
- Modify: `web/src/App.tsx` (已部分实现)
- Modify: `web/src/index.css` (已部分实现)

**目标:** 确保 PWA 在手机上有 44px 最小触控目标, safe-area 适配, 底部导航优化。当前代码已有基础, 此任务为审查和完善。

- [ ] **Step 1: 审查现有触控样式**

检查 `web/src/index.css` 中已有的 `touch-target`、`safe-area-inset-*`、`-webkit-tap-highlight-color` 等规则是否完整。

- [ ] **Step 2: 确保关键按钮有最小触控区域**

在所有可点击元素的 class 中添加 `.touch-target` (min-height/min-width 44px)。

- [ ] **Step 3: 验证构建**

```bash
cd web && npm run build
```
Expected: 构建成功

- [ ] **Step 4: Commit**

```bash
git add web/src/
git commit -m "feat: review and enhance PWA touch optimization"
```

---

## 计划总结

| Task | 内容 | 产出 |
|------|------|------|
| 1 | Go module + 项目结构 | go.mod, Makefile |
| 2 | Schema 定义 + 迁移 | internal/pkg/schema/ |
| 3 | 配置管理 | internal/config/ |
| 4 | SQLite 存储 | internal/store/ |
| 5 | 默认数据 + 提醒 | store init + reminder |
| 6 | HTTP 服务器 + API | internal/server/ |
| 7 | 前端嵌入 + 桌面入口 | cmd/open-todo/ |
| 8 | 桌面 UI (Systray + WebView) | internal/ui/ |
| 9 | DatabaseAdapter + GoDesktopAdapter | web/src/lib/adapter/ |
| 10 | 更新 Zustand Store | web/src/store/ |
| 11 | 移除 Tauri/Rust/Android | 清理旧代码 |
| 12 | WebDAV 备份 | internal/app/ |
| 13 | PWA 配置 | vite-plugin-pwa, manifest |
| 14 | DexieAdapter | PWA 本地存储 |
| 15 | Go 服务端 | server/ |
| 16 | SyncEngine (Go) | internal/sync/ |
| 17 | SyncEngine (TS) | web/src/lib/sync/ |
| 18 | Service Worker Background Sync | web/src/sw.ts |
| 19 | 响应式 UI 审查 | 触控优化 |
