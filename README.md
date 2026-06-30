# Open Todo

一个基于 Go + React + TypeScript 构建的跨平台待办事项应用。桌面端使用 Go + WebView，移动端使用 PWA，支持自建服务端同步。

## 功能特性

- **多视图支持**
  - 看板视图：拖拽式任务管理
  - 日历视图：按日期查看和管理任务
  - 列表视图：传统任务列表

- **任务管理**
  - 创建、编辑、软删除任务
  - 任务优先级（紧急/高/中/低）
  - 任务标签和分类
  - 子任务和提醒
  - 拖拽排序

- **筛选与搜索**
  - 按优先级、标签、状态筛选
  - 关键词搜索

- **数据存储**
  - 本地 SQLite 数据库（桌面端）
  - IndexedDB 本地存储（PWA 移动端）
  - 可选服务端同步（多设备）

- **备份**
  - WebDAV 自动/手动备份

## 架构

```
Go 桌面 ←─ SyncEngine ──→ Go 服务端 ←─ Sync ──→ PWA 移动端
   │                         │                     │
 SQLite                    SQLite              IndexedDB
```

## 技术栈

### 前端
- React 19 + TypeScript
- Vite + Tailwind CSS
- Zustand（状态管理）
- Dexie.js（PWA 本地数据库）
- PWA（vite-plugin-pwa + Service Worker）

### 后端
- Go + Gin（HTTP 服务）
- modernc.org/sqlite（纯 Go SQLite，零 CGo）
- systray（跨平台系统托盘）

### 服务端（可选）
- Go + Gin
- JWT 认证
- SQLite / PostgreSQL

## 环境要求

- Go >= 1.24
- Node.js >= 18
- pnpm

## 快速开始

```bash
# 安装前端依赖
pnpm install

# 构建前端
pnpm run build

# 启动桌面应用
make run
```

浏览器自动打开 `http://localhost:25080`。

## 开发

```bash
# 终端1：启动 Go 后端
go run ./cmd/open-todo/

# 终端2：启动前端开发服务器（热更新）
pnpm dev
```

前端开发服务器运行在 `http://localhost:5173`，API 请求自动代理到 Go 后端。

## 构建

```bash
# 当前平台
make build

# 交叉编译
make build-win     # Windows
make build-linux   # Linux
make build-mac     # macOS
```

构建产物在 `bin/` 目录。

## 服务端部署

```bash
cd server
go build ./cmd/server/

# SQLite 模式（默认）
./server --driver sqlite --dsn /data/open-todo.db

# PostgreSQL 模式
./server --driver postgres --dsn "postgres://user:pass@host/db"
```

## 项目结构

```
open-todo/
├── cmd/open-todo/          # Go 桌面入口
├── internal/
│   ├── app/                # 业务逻辑（备份、提醒）
│   ├── config/             # 配置管理
│   ├── pkg/schema/         # 数据库 schema
│   ├── server/api/         # HTTP API 路由
│   ├── store/              # SQLite 数据库层
│   ├── sync/               # 同步引擎
│   └── ui/                 # 桌面 UI（systray）
├── server/                 # Go 服务端（独立模块）
├── src/                    # React 前端
└── public/                 # 静态资源 + PWA manifest
```

## 许可证

MIT
