## T37 — HomeView Assembly

### Changes
- `frontend/src/views/HomeView.vue` — Complete rewrite: composes all 12 components via LayoutShell slots
- `frontend/src/components/GalleryGrid.vue` — Added `card-click` emit propagation from GalleryCard (was missing)

### Architecture
- **Data flow**: API (`getImages`/`searchImages`) → `mapImage()` adapter (`url→lsky_url`, `user_id→uploaded_by`) → component props
- **State reset**: Search/album filter reset page to 1 and clear images before fetching
- **Load-more guard**: `loadingMore` ref prevents rapid double-fetches
- **Search vs empty**: `searchNoResults` flag suppresses GalleryGrid's "暂无图片" and shows "未找到相关图片" instead
- **ImagePopup**: `v-if` + computed `popupImage` ensures non-null image prop; destroyed/recreated on close
- **CommentSection**: Conditionally rendered (`v-if="showComments"`) with `:key="selectedImage?.id"` for proper re-mount on image switch

### API Response Mapping
- API `ImageData.url` → component `GalleryImage.lsky_url`
- API `ImageData.user_id` → component `GalleryImage.uploaded_by`
- API `ImageListParams.page_size` (not `limit` as task described)
- Response: `res.data.data[]` (PaginatedResponse wrapper)

### Placeholder Values (for future wiring)
- `TopNotifyBar`: onlineCount="--", newMembers="--", uptime="--"
- `VisitorCounter`: count=0
- Both components exist but StatsBar doesn't expose its WebSocket data externally

### Verification
- `npm run build` passes clean (vue-tsc -b && vite build)
- 151 modules transformed, 0 errors

### GalleryGrid Card-Click Fix
- Added `'card-click': [id: number]` to emits
- Added `@click="(id: number) => emit('card-click', id)"` on GalleryCard

## T30-T33b — Vue Cyperpunk Components

### Created Files (5 components)
- `frontend/src/components/SearchBar.vue` — Input with neon border, 300ms debounce via setTimeout, @keyup.enter trigger, × clear button, emits @search
- `frontend/src/components/CommentSection.vue` — Comment CRUD with GET/POST/DELETE /api/images/{id}/comments, pagination "加载更多", conditional delete for owner/admin
- `frontend/src/components/StatsBar.vue` — WebSocket (/ws) stats display, reconnect 3s on close, cleanup onUnmounted, initial "--" placeholders
- `frontend/src/components/VisitorCounter.vue` — Digit-by-digit rendering in separate span.monospace with neon-green glow
- `frontend/src/components/FloatingButtons.vue` — Fixed bottom-right (z-index 100), TOP scroll button, login/register or upload based on isLoggedIn

### Design System Used
- All colors reference CSS variables: --neon-green, --neon-cyan, --neon-yellow, --neon-pink, --neon-red
- Fonts: --font-mono for code/stats, --font-display for body text
- Utility classes: neon-box (from main.css) for consistent neon border + glow
- Glow variables: --glow-green, --glow-cyan, --glow-yellow, --glow-pink
- Pattern: all use `<script setup lang="ts">` with `withDefaults(defineProps<>(), {})` and `defineEmits<>()`

### Verification
- `npm run build` passes clean (vue-tsc -b && vite build)
- 29 modules transformed, no TypeScript errors

## T36 — Auth Views (Login / Register / Pending)

### Three views implemented replacing stubs:

**LoginView.vue** — Neon-pink double border card, scanline overlay, ambient glow orbs
- Form: email + password with neon-green inputs (cyan on focus)
- Validation: email regex, password >= 6 chars, Chinese error messages
- Calls `authStore.login()` on submit, `router.push('/')` on success
- Error shake animation + neon-red error box
- Register link: "还没有账号？立即注册"
- Loading spinner on button during submission

**RegisterView.vue** — Neon-green double border card, ambient green+cyan orbs
- Form: email + nickname + password
- Validation: email, nickname required, password >= 6 chars
- Calls `authStore.register()` on submit, shows success message inline (no redirect)
- Success state: "注册成功，请等待管理员审核" with link to login
- Error: "该邮箱已被注册" when backend says email exists
- Form clears on success, hidden behind success box

**PendingView.vue** — Neon-orange double border card, ambient orange+yellow orbs
- Pulsing dot with expanding ring animation (CSS `pulse-expand` + `pulse-dot` keyframes)
- Info box showing status "待审核" (blinking) + user email
- "重新检查状态" button calls `authStore.tryRestoreSession()`, redirects if approved
- "返回首页" link
- Scanline overlay for CRT effect

### Design Tokens Used Across All Three
- `--bg-primary`, `--neon-green`, `--neon-pink`, `--neon-cyan`, `--neon-yellow`, `--neon-orange`, `--neon-red`, `--neon-blue`
- `--glow-green`, `--glow-yellow`, `--glow-cyan`, `--glow-pink`
- `--font-display`, `--font-mono`
- Utility classes: `.glow-text` for neon text, no custom `.neon-box` variants (used inline border styles to differentiate per-view color)

### Verification
- `npm run build` passes clean — 151 modules transformed, no TypeScript errors
- Fixed unused `router` import in RegisterView after initial build failure

## ImagePopup.vue — Draggable floating image viewer

### Created File
- `frontend/src/components/ImagePopup.vue` — Demo-style draggable floating popup for image viewing

### Features Implemented
- Props: `image` (Image interface: id, title, lsky_url, tags, uploaded_by, created_at), `visible` boolean
- Internal `showPopup` ref manages visibility; v-show preserves DOM/animation state
- Draggable: mousedown on header → mousemove updates left/top → mouseup stops, with viewport clamping
- Close animation: 0.3s ease-in scale(0) opacity(0) transition → auto-reappear after 5-13s random delay with bounce-in
- Open animation: custom `popup-scale-in` keyframe (scale 0→1.05→1)
- First open: random position within padded viewport, re-clamped on drag
- Image error fallback state with `imgLoadError` ref
- Uploader displayed as `User#${uploaded_by}` since backend model has int64 uploaded_by (user ID, not username)

### Design System Used
- Colors: `--neon-yellow` (border), `--neon-red` (close btn, glow), `--neon-cyan` (uploader text), `--neon-pink/green/orange/purple` (tags)
- Background: `linear-gradient(180deg, #1a0000, #0a0a0a)` dark gradient
- Border: `3px ridge var(--neon-yellow)` matching demo style
- Box-shadow: `0 0 30px var(--neon-red), 0 0 60px var(--neon-yellow)`
- Header: `linear-gradient(90deg, var(--neon-red), var(--neon-yellow), var(--neon-red))` matching demo .popup-header
- Fonts: `--font-display` for body, `--font-mono` for tags/info

### Code Conventions Followed
- `<script setup lang="ts">` pattern with Composition API
- `defineProps<T>()` and `defineEmits<T>()` TypeScript generics
- `scoped` styles with CSS variables
- `formatDate()` mirrors GalleryCard.vue pattern
- `tagColors` / `tagColor()` mirrors GalleryCard.vue pattern
- Image interface matches backend `model.Image` struct JSON tags

### Verification
- `npm run build` passes: vue-tsc + vite build clean
- No TypeScript errors after removing unused `onMounted` import

## T7 - Go Data Model Definitions

### Created files
- `backend/internal/model/user.go` — User struct + UserRole/UserStatus constants
- `backend/internal/model/album.go` — Album struct
- `backend/internal/model/image.go` — Image struct with Tags `[]string` JSON field
- `backend/internal/model/comment.go` — Comment struct
- `backend/internal/model/request.go` — Register, Login, CreateAlbum, CreateImage, UpdateImage, CreateComment request structs with validate tags
- `backend/internal/model/response.go` — APIResponse, TokenPair, ImageDetail, Claims, RefreshRequest

### Conventions used
- All structs use double tags: `json:"field" db:"column"` for both JSON serialization and database/sql scanning
- PasswordHash tagged with `json:"-"` to prevent leaking in API responses
- Tags field in Image uses `db:"tags"` for JSON text storage in SQLite
- AlbumID in Image uses `*int64` (pointer) for nullable FK
- ImageDetail embeds Image + CommentCount for API response enrichment
- `go build ./internal/model/` and `go vet ./internal/model/` both pass clean

## T2 - Go Project Scaffolding

### Created structure
- `backend/cmd/server/main.go` — Echo server with /health route, graceful shutdown (signal.NotifyContext)
- `backend/internal/config/config.go` — Config stub (empty struct, detailed impl in T4)
- `backend/internal/deps/deps.go` — Blank imports to pin all direct dependencies in go.mod
- `backend/migrations/` — empty directory for future SQL migrations

### Dependencies installed (10 direct deps)
- github.com/labstack/echo/v4 v4.15.2
- github.com/golang-jwt/jwt/v5 v5.3.1
- modernc.org/sqlite v1.50.0
- github.com/golang-migrate/migrate/v4 v4.19.1
- gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df
- github.com/gorilla/websocket v1.5.3
- github.com/stretchr/testify v1.11.1
- github.com/spf13/viper v1.21.0
- github.com/go-playground/validator/v10 v10.30.2
- golang.org/x/crypto v0.50.0

### Gotchas
- `go mod tidy` removes unused deps. Pin deps with blank imports in a deps.go file to keep them in go.mod before they're actually imported by business logic.
- `golang.org/x/crypto` has no root package — use `golang.org/x/crypto/bcrypt` (or another subpackage) for blank imports.
- `github.com/stretchr/testify` root package exists but `testify/assert` is more idiomatic for blank imports.
- `go build ./cmd/server` passes clean with zero diagnostics.

## T14 - Admin Handler (User Review)

### Created Files
- `backend/internal/handler/admin.go` — AdminHandler with ListPending, ApproveUser, RejectUser
- `backend/internal/handler/admin_test.go` — 8 tests covering all scenarios

### Key Patterns
- AdminHandler dependencies: `*repository.UserRepository` + `service.EmailService`
- Routes registered with `middleware.AuthRequired(jwtSvc)` + `middleware.AdminRequired()` middleware chain
- Approve/Reject share `updateUserStatus` private method to avoid duplication
- Status update blocked if `user.Status != model.UserStatusPending` (handles already-approved/rejected → 400)
- Email sent asynchronously via `go func()` — ignored in tests via `mockEmailService`
- `password_hash` excluded automatically via `json:"-"` tag on `model.User.PasswordHash`

### Test Patterns
- `setupAdminTest` returns `(*echo.Echo, *AdminHandler, *sql.DB, *service.JWTService, func())`
- `generateAccessToken` helper creates valid JWT tokens for auth
- `doAuthRequest` helper adds `Authorization: Bearer <token>` header
- DB verification via `db.QueryRow("SELECT status FROM users WHERE id = ?", id)` after status updates

### Pitfalls
- `auth.go` was modified by T12/T13 to require `EmailService` + `adminEmail` in `NewAuthHandler` — needed to update `setupAuthTest` in `auth_test.go` to compile
- Test regex `-run TestAdmin` doesn't match tests named `TestListPending_Admin` etc. — use `-run "Test(ListPending|ApproveUser|RejectUser)"` instead

## T4 - Config Module

### Created files
- `backend/internal/config/config.go` — Full Config struct with LoadConfig(path string) using viper
- `backend/internal/config/config_test.go` — TestLoadConfig, TestDefaults, TestEnvOverride, TestLoadConfigNoFile
- `backend/config.yaml` — Full config template with Chinese comments
- `backend/.env.example` — Environment variable override examples

### Implementation details
- `viper.SetConfigFile(path)` with `os.Stat` pre-check to handle non-existent path gracefully
- Default search paths: `.` and `./backend` for `config.yaml` when path is empty
- Environment vars use `GM_` prefix with dot-to-underscore replacer (GM_SMTP_PASSWORD → smtp.password)
- Go struct mapstructure tags for viper unmarshal compatibility
- time.Duration fields (AccessExpire, RefreshExpire) auto-parsed from strings like "15m" via viper's default StringToTimeDurationHookFunc
- viper v1.21.0 uses `github.com/go-viper/mapstructure/v2` (not mitchellh/mapstructure)

### Default values configured
- server.port: 1323, server.host: "0.0.0.0"
- database.path: "gm_site.db"
- jwt.access_expire: "15m", jwt.refresh_expire: "168h"
- smtp.port: 587, smtp.use_tls: true
- site.name: "顾夏"
- upload.max_size_mb: 10

### Important: TestLoadConfigNoFile
- When `SetConfigFile(path)` is used and file doesn't exist, viper returns `*os.PathError`, NOT `viper.ConfigFileNotFoundError`
- Fix: pre-check with `os.Stat(path)` before calling `SetConfigFile(path)`
- When file is not found, silently fall through to defaults + env vars

## T11 - Email Service (SMTP TLS)

### Created files
- `backend/internal/service/email.go` — EmailService interface, SmtpEmailService, MockEmailService
- `backend/internal/service/email_test.go` — Tests for Mock + constructor + real SMTP (SkipIf)

### Implementation details
- SMTP STARTTLS flow for port 587: `net.Dial("tcp", addr)` → `smtp.NewClient` → `StartTLS` → `PlainAuth` → `Mail` → `Rcpt` → `Data` → `Quit`
- `net.JoinHostPort` used instead of `fmt.Sprintf("%s:%d")` for IPv6 compatibility (go vet requires this)
- Chinese Subject lines encoded with `mime.BEncoding.Encode("utf-8", subject)` for RFC 5322 compliance
- `InsecureSkipVerify: false` in TLS config for production security
- MockEmailService implements EmailService interface, logs to `log.Printf`
- Compile-time interface check: `var _ EmailService = (*MockEmailService)(nil)`

### Real SMTP test
- `TestSmtpEmailService_Send` reads SMTP credentials from env vars (`GM_SMTP_*`)
- Skips automatically if credentials not configured
- Covers: SendAdminNotification, SendApprovalNotification (approved + rejected)

### Gotchas
- `go vet` requires `net.JoinHostPort` for address formatting to support IPv6
- `mime.BEncoding.Encode` needed for Chinese subject lines in email headers
- PlainAuth requires the hostname used in auth to match the server certificate

## T6 - SQLite Migration Scripts

### Created files (8 migration files)
- `backend/migrations/001_create_users.up.sql` / `.down.sql`
- `backend/migrations/002_create_albums.up.sql` / `.down.sql`
- `backend/migrations/003_create_images.up.sql` / `.down.sql`
- `backend/migrations/004_create_comments.up.sql` / `.down.sql`

### Conventions
- Up/down pair per version: `{version}_{description}.up.sql` / `.down.sql` (golang-migrate convention)
- Foreign keys use `REFERENCES table(column)` syntax (compatible with modernc.org/sqlite when `PRAGMA foreign_keys=ON`)
- `images.tags` → `TEXT NOT NULL DEFAULT '[]'` for JSON array string storage
- `images.album_id` → nullable FK with `ON DELETE SET NULL`
- `comments.image_id` → `ON DELETE CASCADE`
- Indexes on all queried columns for performance

## T5 - Database connection + migrations (2026-05-09)

### Implementation
- database.NewDatabase(path) creates SQLite connection with modernc.org/sqlite
- Connection pool: MaxOpenConns=1, MaxIdleConns=1, ConnMaxLifetime=1h
- DSN: ile:<path>?cache=shared&_journal_mode=WAL&_busy_timeout=5000
- database.RunMigrations(db, path) uses golang-migrate/v4/database/sqlite (NOT sqlite3!)

### Key findings
- golang-migrate/v4/database/sqlite uses modernc.org/sqlite (pure Go, no CGO)
- golang-migrate/v4/database/sqlite3 uses mattn/go-sqlite3 (requires CGO) — DO NOT USE
- Source URL format: ile:<absolute_path> (uses URL.Opaque field, compatible with both Windows and Unix)
- ile:///<path> prefix caused issues on Windows (path became /C:/... instead of C:/...)
- Empty migrations dir returns s.PathError{Err: fs.ErrNotExist} from source driver's First() — must handle this alongside migrate.ErrNoChange
- Server startup sequence: config → database → migrations → echo server

## T8 - User Repository

### Created files
- \ackend/internal/repository/user.go\ — UserRepository with 7 CRUD methods (Create, FindByEmail, FindByID, UpdateStatus, ListPending, CountByStatus, UpdateUser)
- \ackend/internal/repository/user_repository_test.go\ — 11 tests (Create, FindByEmail, FindByEmail_NotFound, FindByID, FindByID_NotFound, UpdateStatus, UpdateStatus_NotFound, ListPending, ListPending_Empty, CountByStatus, UpdateUser)

### Conventions used
- Repository constructor takes \*sql.DB\ and returns \*UserRepository\
- All queries use \?\ parameter placeholder (SQLite syntax via modernc.org/sqlite)
- Time fields set in Go before insert (not relying on SQLite DEFAULT CURRENT_TIMESTAMP) for deterministic tests
- \sql.ErrNoRows\ propagated directly from QueryRow for FindByEmail/FindByID
- Update methods return \sql.ErrNoRows\ when no rows affected (0 rows matched)
- ListPending returns empty slice (not nil) for consistency when no results
- Model constants used for status/role values (untyped string constants)

### Gotchas
- Module path is \gm_site\ (not \gm_site/backend\) — import model as \gm_site/internal/model\
- model.UserStatus constants are untyped string constants, not a defined type — use plain \string\ in function parameter types
- Test database created via \	.TempDir()\ + \cache=shared&_journal_mode=WAL\ DSN to match production settings
- \setupTestDB\ helper creates table+indexes matching migrations/001_create_users.up.sql exactly
- All 11 tests pass: \go test ./internal/repository -v -run TestUser\ → PASS in ~1.5s

## T9 - JWT Service (2026-05-09)

### Created files
- \`backend/internal/service/jwt.go\` — JWTService with 6 methods
- \`backend/internal/service/jwt_test.go\` — 7+ test functions

### Implementation details
- JWTService constructor takes secrets (strings) and durations — no direct config coupling
- Access tokens use HS256 signing with custom \`Claims\` struct (embeds \`jwt.RegisteredClaims\`)
- Refresh tokens use HS256 signing with plain \`jwt.RegisteredClaims\` (subject = userID string)
- \`ValidateAccessToken\` returns \`*Claims\` with UserID and Role
- \`ValidateRefreshToken\` returns \`int64\` (parsed from Subject field)
- \`GenerateTokenPair\` returns \`*model.TokenPair\` with both tokens + ExpiresIn unix timestamp
- Custom Claims struct in service package (distinct from \`model.Claims\` DTO)
- Key func validates signing method (\`*jwt.SigningMethodHMAC\`) before returning secret

### Test coverage
- 7 test functions: GenerateAndValidateAccessToken, GenerateAndValidateRefreshToken, ExpiredAccessToken, ExpiredRefreshToken, InvalidToken (5 subtests), TokenPair, ClaimsValues (3 subtests)
- Short-lived tokens (1s expiry) with 1500ms sleep verify expiration rejection
- Cross-secret validation tested: refresh token fails access validation, vice versa
- \`errors.Is(err, jwt.ErrTokenExpired)\` used for expiration checks
- All pass: \`go test ./internal/service -v -run TestJWT -count=1\` → PASS in ~4.4s

## T13 - Auth handler (Login + Refresh)
- **Date**: 2026-05-09
- Auth handler uses constructor injection for UserRepository and JWTService
- Login flow: bind JSON → FindByEmail → bcrypt compare → status check → GenerateTokenPair → response
- Status checks: pending→403("账户正在审核中"), rejected→403("账户已被拒绝")
- User not found and wrong password both return 401("邮箱或密码错误") to avoid email enumeration
- Refresh flow: bind JSON → ValidateRefreshToken → FindByID(to get role) → GenerateTokenPair → response
- Refresh errors: expired→401("刷新令牌已过期"), invalid→401("无效的刷新令牌")
- Response helpers: JSON(c, code, message, data), Success(c, data), Created(c, data), Error(c, code, message)
- All responses use model.APIResponse{Code, Message, Data} envelope
- Test pattern: setupAuthTest returns (echo, handler, db, teardown) for full integration testing with real SQLite
- 8 tests all passing
- jwtSvc field on AuthHandler is accessed directly in tests (same package)
- generateExpiredRefreshToken helper manually constructs JWT with past expiry for expired token tests

## T12 - Register Handler (2026-05-09)

### Changes Made
- ackend/internal/handler/auth.go — Updated AuthHandler (added emailSvc + adminEmail fields), updated NewAuthHandler, added Register handler
- ackend/internal/handler/auth_test.go — Added setupAuthRegisterTest helper, 6 Register tests
- ackend/internal/service/email.go — Updated MockEmailService with Mutex-protected Messages slice for recording calls

### Register Handler Implementation
- Validates request with go-playground/validator/v10 struct tags (required, email, min=6)
- Checks duplicate email via userRepo.FindByEmail → 409 Conflict
- Bcrypt hash with cost=10
- Role determined by comparing email to adminEmail param ("admin" vs "user")
- User created with status="pending" (never activated directly)
- Async email notification via go emailSvc.SendAdminNotification(...)
- Returns 201 Created with message "注册成功，请等待管理员审核"
- Validation errors return 400 with field-level error details (e.g., {"Email": "email", "Password": "min"})

### Test Patterns
- setupAuthRegisterTest(adminEmail) creates dedicated Echo instance with only Register route
- DB verification via db.Query(...) after registration to check stored values
- ssert.Eventually used to wait for goroutine-based email notification
- MockEmailService.Messages[] records calls for assertion

### Key Findings
- ssert.Contains with Chinese substrings may fail due to Unicode normalization differences — use ssert.Equal with the full expected string instead
- Using go func() for async email requires thread-safe mock (sync.Mutex on Messages slice)
- Existing tests didn't need changes — setupAuthTest internally updated to pass new constructor args

## T21 - WebSocket Hub + Client Management

### Created files
- `backend/internal/websocket/hub.go` — Hub struct with clients map, register/unregister/broadcast channels, sync.RWMutex
- `backend/internal/websocket/client.go` — Client struct with ReadPump/WritePump goroutines, ping/pong heartbeat
- `backend/internal/handler/ws.go` — ServeWS handler for WebSocket upgrade, extracts optional userID from middleware context
- `backend/internal/websocket/hub_test.go` — 5 tests: RegisterClient, UnregisterClient, Broadcast, ClientCount, BroadcastGracefulIgnoreFullSendBuffer

### Key Patterns
- Hub uses channel-based event loop (Run goroutine) for register/unregister/broadcast — avoids mutex contention on hot path
- sync.RWMutex used for ClientCount() and the clients map iteration in Broadcast
- Client send channel buffer size: 256 messages
- Heartbeat: pingPeriod = pongWait * 9/10 (54s ping, 60s pong timeout)
- WritePump batches queued messages into a single WebSocket frame via NextWriter
- ReadPump echoes received messages to all clients (chat broadcast pattern)
- When send buffer is full, client is dropped gracefully (no deadlock)

### Gotchas
- Circular import: `internal/websocket` <-> `internal/handler`. Solved by inlining the WS upgrade handler in hub_test.go instead of importing handler.ServeWS
- gorilla/websocket is already in go.mod (v1.5.3) via deps.go blank import
- Test uses httptest.Server + gorilla/websocket.Dialer for end-to-end WebSocket testing
- Need time.Sleep(50ms) after Register to let hub goroutine process the registration before asserting ClientCount

### Test results
- All 5 tests pass, 0 lsp errors (1 cosmetic hint on for loop modernization)
