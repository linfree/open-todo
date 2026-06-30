# Task 9 Report: Frontend DatabaseAdapter + GoDesktopAdapter

**Status:** Complete
**Commit:** 091db9e

## Files Created

| File | Purpose |
|------|---------|
| `src/lib/adapter/types.ts` | `Task`, `TaskList`, `ChangeRecord`, `DatabaseAdapter` interfaces matching Go API schema |
| `src/lib/adapter/DatabaseAdapter.ts` | Re-export barrel for adapter types |
| `src/lib/adapter/GoDesktopAdapter.ts` | `fetch()`-based implementation calling Go Gin API at same-origin |
| `src/lib/adapter/factory.ts` | `createAdapter()` with singleton caching, defaults to `GoDesktopAdapter` |

## Files Modified

| File | Changes |
|------|---------|
| `src/lib/api.ts` | Replaced `isTauri()`/`invoke()` with adapter-based calls; added `isDesktop()` environment detection; added adapterApp type mapping layer; kept localStorage fallback for non-desktop usage; kept `notificationApi`/`webdavApi` with fetch-based Go API calls |
| `src/components/settings/NotificationSettings.tsx` | Changed `isTauri` -> `isDesktop as checkIsDesktop`; removed unused `Save` import |
| `src/components/settings/DataSettings.tsx` | Changed `isTauri` -> `isDesktop as checkIsDesktop` |
| `src/components/TaskFilterSidebar.tsx` | Removed unused `Bell`, `Database`, `Cloud` imports (pre-existing) |

## Architecture

```
src/lib/adapter/
  types.ts          -- Interfaces (Task, TaskList, ChangeRecord, DatabaseAdapter)
  DatabaseAdapter.ts -- Re-export barrel
  GoDesktopAdapter.ts -- fetch() to Go Gin API on same-origin
  factory.ts         -- createAdapter() singleton factory

src/lib/api.ts
  - Imports adapter via factory
  - Maps between app types (camelCase, Date, Tag[] etc.) and adapter types (snake_case, JSON strings)
  - databaseApi delegates to adapter with fallback to localStorage
  - notificationApi / webdavApi use fetch() with localStorage fallback
```

## Type Mapping

App types (`src/types/index.ts`) use camelCase with complex nested types (Date, Tag[], SubTask[], Reminder[]). Adapter types use snake_case with JSON strings for complex fields, matching the Go backend wire format. The `api.ts` layer handles bidirectional conversion:

- `adapterTaskToApp()` - snake_case JSON strings -> camelCase with parsed arrays and Date objects
- `appTaskToAdapter()` - camelCase with Date/arrays -> snake_case with JSON strings
- `adapterListToApp()` / `_appListToAdapter()` - similar for TaskList

## Verification

- `npx tsc --noEmit` passes with zero errors
- No remaining `isTauri` or `@tauri-apps/api/core` references in `src/`
- Environment detection via `isDesktop()` checks `__GO_DESKTOP__` / legacy `__TAURI__` window properties
