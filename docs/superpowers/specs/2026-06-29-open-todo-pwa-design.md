# Open Todo — Tauri 桌面 + PWA 移动端 架构设计

> 日期: 2026-06-29 | 状态: 已确认

## 概述

将 my-todo 项目重构为 **open-todo**，架构从 "Tauri (Rust) 桌面 + Tauri Android" 改为 **Go 桌面 + PWA 移动端 + Go 服务端**。

### 目标

1. **降低维护成本** — 移除 Rust/Rust → 全栈 Go + TypeScript
2. **覆盖 iOS** — PWA 天然支持 iPhone
3. **安装门槛** — PWA 无需应用商店
4. **多用户同步** — Go 服务端支持登录、多设备同步

---

## 一、总体架构

```
┌──────────────────────────────────────────────────────────┐
│                 React SPA (同一份前端代码)                  │
│              Zustand Store → 业务逻辑不变                  │
└────────────────────────────┬─────────────────────────────┘
                             │
                  ┌──────────┴──────────┐
                  │   DatabaseAdapter   │
                  │   (策略模式)         │
                  └─────────┬───────────┘
                            │
        ┌───────────────────┼───────────────────┐
        ▼                   ▼                   ▼
┌───────────────┐ ┌───────────────┐ ┌───────────────┐
│GoDesktopAdapter│ │ DexieAdapter  │ │ ServerAdapter │
│invoke→Gin API  │ │Dexie/IndexedDB│ │HTTP→Go 服务    │
│本地 SQLite     │ │PWA 本地数据库 │ │远程同步        │
└───────┬───────┘ └───────┬───────┘ └───────┬───────┘
        │                 │                 │
        ▼                 ▼                 ▼
┌───────────────┐ ┌───────────────┐ ┌───────────────┐
│  Go 桌面应用   │ │  PWA 移动端   │ │  Go 服务端     │
│  Gin + systray│ │  SW + Dexie   │ │  多用户同步     │
│  + WebView2   │ │  Offline-First│ │  SQLite/PG     │
└───────────────┘ └───────────────┘ └───────────────┘
```

### 三端职责

| 端 | 技术栈 | 本地存储 | 同步角色 |
|---|--------|---------|---------|
| 桌面 | Go + Gin + WebView2/systray | SQLite | Client, 可离线和在线 |
| PWA | React + Dexie.js + Service Worker | IndexedDB | Client, 可离线和在线 |
| 服务端 | Go + Gin | SQLite/PostgreSQL | Sync relay + 用户认证 |

### 用户数据归属方式

1. **纯本地** — SyncEngine 未启动，数据仅在本机
2. **纯服务端** — 登录后数据在服务端，可跨设备
3. **本地 + 同步** — 本地有副本，自动双向同步

---

## 二、Go 桌面应用

### 目录结构

```
cmd/
  open-todo/
    main.go           # 启动流程
    embed.go          # //go:embed web-dist/*

internal/
  app/                # 任务 CRUD + 备份 + 提醒
    task.go
    list.go
    backup.go
    reminder.go

  store/              # SQLite (modernc.org/sqlite)
    store.go
    task_store.go
    list_store.go
    sync_store.go

  sync/               # 同步引擎
    engine.go
    tracker.go         # sync_log 变更记录
    client.go          # HTTP Client → 服务端

  server/
    server.go          # Gin 引擎
    api/
      task.go          # /api/v1/tasks
      list.go
      sync.go          # /api/v1/sync/push, pull
      backup.go

  ui/                 # 桌面 UI
    ui.go
    ui_windows.go      # WebView2
    ui_darwin.go       # systray + browser
    ui_linux.go        # systray + browser

  config/
    config.go          # ~/.open-todo/config.json

pkg/
  schema/
    schema.go          # 表结构常量 (与服务器共享)
    migrate.go         # 迁移管理

web/                  # React 前端
```

### 启动流程

1. 加载配置 `~/.open-todo/config.json`
2. 打开 SQLite `~/.open-todo/open-todo.db` (WAL)
3. 数据库迁移
4. 初始化 SyncEngine (登录后)
5. 创建 Gin → API 路由 + 静态资源 (go:embed) + SPA fallback
6. 监听端口 (默认 18080, fallback 扫描)
7. 启动提醒检查器 (10s)
8. 启动系统托盘
9. 阻塞等待退出

### 与 cc-go 参考项目的差异

| 组件 | cc-go | open-todo |
|------|-------|-----------|
| 核心 | Claude 子进程管理 | 任务 CRUD + 提醒 |
| 数据库 | sessions 单表 | 5表 + sync_log |
| 实时通信 | WebSocket (Claude 输出流) | WebSocket (同步通知) |
| WebDAV | 无 | 有 (Rust→Go 迁移) |
| 同步引擎 | 无 | SyncEngine (新增) |

---

## 三、数据库设计

### SQLite Schema

```sql
-- 用户
CREATE TABLE users (
    id          TEXT PRIMARY KEY,
    email       TEXT UNIQUE NOT NULL,
    name        TEXT NOT NULL DEFAULT '',
    password    TEXT NOT NULL,
    created_at  TEXT NOT NULL,
    updated_at  TEXT NOT NULL
);

-- 清单
CREATE TABLE lists (
    id          TEXT PRIMARY KEY,
    user_id     TEXT NOT NULL DEFAULT '',
    name        TEXT NOT NULL,
    icon        TEXT,
    color       TEXT,
    order_num   INTEGER NOT NULL DEFAULT 0,
    created_at  TEXT NOT NULL,
    updated_at  TEXT NOT NULL
);

-- 任务
CREATE TABLE tasks (
    id          TEXT PRIMARY KEY,
    user_id     TEXT NOT NULL DEFAULT '',
    title       TEXT NOT NULL,
    description TEXT,
    completed   INTEGER NOT NULL DEFAULT 0,
    priority    TEXT NOT NULL DEFAULT 'none',
    status      TEXT NOT NULL DEFAULT 'todo',
    list_id     TEXT NOT NULL,
    tags        TEXT NOT NULL DEFAULT '[]',
    sub_tasks   TEXT NOT NULL DEFAULT '[]',
    reminders   TEXT NOT NULL DEFAULT '[]',
    due_date    TEXT,
    deleted     INTEGER NOT NULL DEFAULT 0,
    deleted_at  TEXT,
    order_num   INTEGER NOT NULL DEFAULT 0,
    created_at  TEXT NOT NULL,
    updated_at  TEXT NOT NULL
);

-- 提醒记录
CREATE TABLE sent_reminders (
    id              TEXT PRIMARY KEY,
    user_id         TEXT NOT NULL DEFAULT '',
    task_id         TEXT NOT NULL,
    reminder_time   INTEGER NOT NULL,
    sent_at         INTEGER NOT NULL,
    reminder_data   TEXT
);

-- 同步变更日志
CREATE TABLE sync_log (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    table_name  TEXT NOT NULL,
    record_id   TEXT NOT NULL,
    action      TEXT NOT NULL,
    timestamp   TEXT NOT NULL,
    synced      INTEGER NOT NULL DEFAULT 0
);
```

### Dexie Schema (PWA)

```typescript
db.version(1).stores({
  lists:  'id, user_id, updated_at',
  tasks:  'id, user_id, title, list_id, updated_at, [user_id+list_id]',
  sent_reminders: 'id, user_id, task_id, reminder_time',
  sync_log: '++id, table_name, record_id, synced, timestamp',
  session: 'key',  // JWT token
});
```

### 关键字段

- `updated_at` — LWW 冲突解决唯一依据
- `deleted` — 软删除，同步必须标记
- `user_id` — 本地模式为空，登录后填充
- `sync_log` — 每次变更记录，推送完成后标记

---

## 四、SyncEngine 同步引擎

### 推/拉流程

```
本地变更 → sync_log (synced=0)
           → PUSH POST /api/v1/sync/push
           → 标记 synced=1

定时轮询 → PULL POST /api/v1/sync/pull (since=last_sync)
          → LWW 合并 → 写入本地
```

### 触发时机

- 本地变更后: debounce 3s 推送
- 定时拉取: 30s
- 网络恢复: 立即全量
- 应用启动: 立即拉取

### LWW 冲突解决

```
比较 remote.updated_at vs local.updated_at
  谁新用谁
  相等用 remote
```

### 软删除协议

1. 本地 delete → `deleted=1, updated_at=now` → push
2. 其他设备 pull 到 → 标记 `deleted=1`
3. 30天后清理 `deleted_at > 30d` 的记录

### 桌面 vs PWA

| | 桌面 (Go) | PWA (TS + SW) |
|---|---|---|
| SyncEngine | goroutine | Service Worker |
| 变更追踪 | DB trigger/埋点 | Dexie hooks |
| 后台同步 | 常驻 goroutine | Background Sync API |
| 离线检测 | Go net | navigator.onLine |

---

## 五、Go 服务端

### 目录结构

```
server/
  cmd/server/main.go

  internal/
    config/       # 环境变量 + 配置
    database/     # Database 接口 + sqlite/postgres 实现
    model/        # User, Task, SyncRecord
    handler/      # auth, sync, health
    middleware/    # JWT auth, CORS
    router/       # Gin 路由
```

### Database 接口

```go
type Database interface {
    CreateUser(email, name, password string) (*User, error)
    GetUserByEmail(email string) (*User, error)
    GetUserByID(id string) (*User, error)
    StoreChanges(userID string, changes []SyncRecord) error
    GetChangesSince(userID string, since time.Time) ([]SyncRecord, error)
    CleanupDeleted(olderThan time.Time) error
    Migrate() error
    Close() error
}
```

### API

```
POST /api/v1/auth/register   → { user, token }
POST /api/v1/auth/login      → { user, token }
POST /api/v1/auth/refresh    → { token }
POST /api/v1/sync/push       → { server_time }
POST /api/v1/sync/pull       → { changes[], server_time }
GET  /api/v1/health          → { ok, version }
```

### 驱动选择

```bash
./open-todo-server --driver sqlite   --dsn /data/open-todo.db
./open-todo-server --driver postgres --dsn "postgres://user:pass@host/db"
```

---

## 六、PWA 配置

### 技术

- `vite-plugin-pwa` (Workbox)
- `manifest.json` (display: standalone, shortcuts)
- Service Worker (离线缓存 + Background Sync + Push)

### 缓存策略

| 资源类型 | 策略 | 说明 |
|---------|------|------|
| JS/CSS/图片/字体 | CacheFirst | 静态资源, 版本化文件名 |
| API 响应 | NetworkFirst | 先网络后缓存, 离线回退 |
| HTML | NetworkFirst | 确保最新版本 |

### Background Sync

```
离线操作 → sync_log → sw.sync.register('sync-changes')
  → 网络恢复 → 'sync' event → push 到服务端
```

### PWA 安装

监听 `beforeinstallprompt`，显示自定义安装提示。

---

## 七、前端改动

### 现有代码调整

| 当前 | 目标 |
|------|------|
| `isTauri()` 遍地分支 | `adapter.xxx()` 统一调用 |
| `api.ts` ~287行 | 拆分为 adapter 各实现 |
| `todoStore.ts` 直接调 api | 不变，adapter 透明 |
| `window.__TAURI__` 检测 | 检测 Go Gin API 端口 |

### Adapter 接口

```typescript
interface DatabaseAdapter {
  getTasks(): Promise<Task[]>;
  saveTask(task: Task): Promise<void>;
  deleteTask(id: string): Promise<void>;
  getLists(): Promise<TaskList[]>;
  saveList(list: TaskList): Promise<void>;
  deleteList(id: string): Promise<void>;
  getUnsyncedChanges(): Promise<ChangeRecord[]>;
  markChangesSynced(changes: ChangeRecord[]): Promise<void>;
  init(): Promise<void>;
}
```

三个实现: `GoDesktopAdapter`, `DexieAdapter`, `ServerAdapter`

### 共享 Schema 定义

`shared/schema.ts` — 表结构定义，桌面生成 SQL，PWA 生成 Dexie stores。

---

## 八、迁移路径

### 阶段 1: Go 桌面基础

- 搭建 Go 桌面框架 (Gin + systray + WebView2)
- 实现 SQLite 存储 (modernc.org/sqlite)
- 从 Rust 迁移 WebDAV 备份、提醒逻辑
- 前端 Adapter 抽象 (GoDesktopAdapter)
- 移除 Tauri、Rust、Android 代码

### 阶段 2: PWA 移动端

- vite-plugin-pwa 配置
- Dexie.js 本地存储
- Service Worker 离线缓存
- 响应式 UI 优化 (触控目标 44px, 安全区域)

### 阶段 3: Go 服务端 + 同步

- Auth (JWT)
- Sync API (push/pull)
- SyncEngine (桌面 Go + PWA TS)
- 多后端部署支持 (SQLite/PostgreSQL)

### 阶段 4: 打磨

- Push 通知提醒
- 共享清单 (多用户协作)
- Docker 镜像
- 文档 + 自建指南

---

## 九、需要移除的代码

| 文件/目录 | 原因 |
|-----------|------|
| `src-tauri/` | 不再使用 Tauri |
| `src/lib/api.ts` 中的 `isTauri()` 分支 | 改为 Adapter 模式 |
| `build_windows.*` `build_rust.ps1` `list_packages.ps1` | 旧构建脚本 |
| `src-tauri/gen/android/` | 移除 Android 原生 |
| `*.bat` 构建脚本 | Go 用 Makefile |
| Rust 条件编译逻辑 | 不再需要 |
