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
