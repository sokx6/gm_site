# T41: 前端-后端联调 + CORS 配置

## Patterns & Conventions
- Go module path: `gm_site`
- Echo framework v4 used for HTTP routing
- Middleware follows Echo's `echo.MiddlewareFunc` pattern
- All handlers in `internal/handler/` use constructor injection with `New*` pattern
- Repository layer in `internal/repository/` wraps `*sql.DB`
- `StatsService.StartBroadcastLoop()` already starts its own goroutine internally — do NOT wrap with `go`
- CORS preflight (OPTIONS) returns `204 No Content`

## Key Method Names (actual vs. task spec)
- Comment handler: `ListByImage` (not `GetComments`), `Create` (not `CreateComment`), `Delete` (not `DeleteComment`)
- Album handler: `DeleteAlbum` (not `Delete`)
- StatsService: `StartBroadcastLoop` runs internal goroutine, call directly

## Files Created
- `backend/internal/middleware/cors.go` — CORS middleware allowing localhost:5173

## Files Modified  
- `backend/cmd/server/main.go` — Full rewrite with all routes wired:
  - CORS middleware applied globally
  - Visitor tracking middleware (skips /api/health)
  - Public routes: health, albums, images, images/search, images/:id, images/:id/comments
  - Auth routes: register, login, refresh
  - Protected routes (AuthRequired): albums CRUD, images CRUD, comments CRUD
  - Admin routes (AuthRequired + AdminRequired): users pending/approve/reject
  - WebSocket: /api/ws
  - All dependencies injected: jwtService, emailSvc, lskyClient, repos, handlers, wsHub, statsService

## Verification
- `go build ./cmd/server` — PASS
- LSP diagnostics on main.go & cors.go — ZERO issues
