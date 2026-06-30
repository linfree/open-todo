# Final Code Review Fix Report

## Summary

Fixed 3 critical issues identified during the final code review. All fixes compile and pass type checking.

## Commit 1: fix: register backup API routes in server (c39669c)

**File:** `internal/server/server.go`
**Change:** Added `api.RegisterBackupRoutes(apiGroup, st)` after `api.RegisterListRoutes(apiGroup, st)`.

The backup routes (`POST /api/v1/backup`, `POST /api/v1/restore`) were defined in `internal/server/api/backup.go` but never registered in the server initialization, making them unreachable at runtime.

## Commit 2: fix: add getUnsyncedChanges/markChangesSynced to DatabaseAdapter (8f30682)

**Files:**
- `src/lib/adapter/types.ts` — Added `getUnsyncedChanges()` and `markChangesSynced(changes)` to the `DatabaseAdapter` interface
- `src/lib/adapter/DexieAdapter.ts` — Implemented both methods: `getUnsyncedChanges()` filters `syncLog` for `synced === false`; `markChangesSynced()` bulk-updates by IDs to set `synced: true`
- `src/lib/adapter/GoDesktopAdapter.ts` — Added stub implementations that throw "not implemented" (desktop sync is handled by the Go SyncEngine)

## Commit 3: fix: include full record data in sync protocol payloads (0edf59c)

**Files:**
- `internal/sync/client.go` — Added `Data json.RawMessage` field to `SyncRecord`
- `internal/sync/engine.go` — Updated `doPush()` to call `GetFullRecord()` for each change and embed the full record data in the push payload
- `internal/store/sync_store.go` — Added `GetFullRecord(tableName, recordID)` method that queries the actual table row and returns it as JSON; handles both "tasks" and "lists" tables
- `server/internal/database/database.go` — Added `Data json.RawMessage` field to server-side `SyncRecord`
- `server/internal/database/sync.go` — Updated `StoreChanges()` to parse the `Data` field via a `taskPayload` struct and INSERT OR REPLACE with actual values instead of empty defaults

## Build Verification

All builds pass:
- `go build ./cmd/open-todo/` — OK
- `go build ./internal/sync/` — OK
- `cd server && go build ./cmd/server/` — OK
- `npx tsc --noEmit` — OK
