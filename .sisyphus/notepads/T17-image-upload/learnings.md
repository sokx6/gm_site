## T17 Image Upload - Learnings

### Patterns Used
- Repository pattern: `ImageRepository` follows same convention as `AlbumRepository` — `NewXxxRepository(db *sql.DB)`, `Create` method with auto-populated ID/timestamps
- Handler pattern: `ImageHandler` follows `AlbumHandler` — struct with dependencies, constructor, method as Echo handler
- Test helpers: Shared across `handler` package (same package tests) — `insertUser` from auth_test.go, `generateAccessToken`/`doAuthRequest` from admin_test.go, `doRequest`/`parseResponse` from auth_test.go
- Lsky mock: httptest server handling `/api/v1/tokens` and `/api/v1/upload` endpoints

### Key Decisions
1. **Tags as JSON array**: `[]string` marshalled to JSON string for SQLite TEXT column; empty slice initialized as `make([]string, 0)` for `[]` (not `null`) in JSON
2. **MIME type validation**: Read first 512 bytes via `http.DetectContentType` then `file.Seek(0, io.SeekStart)` to reset before passing to Lsky uploader
3. **File size check**: Before MIME validation and Lsky upload, using `fileHeader.Size > int64(maxSizeMB)*1024*1024`
4. **Auth double-check**: Handler checks `c.Get(middleware.UserIDKey)` even though route is behind `AuthRequired` middleware (defensive)
5. **maxSizeMB=0 in tests**: For FileTooLarge test, setting maxSizeMB=0 makes any non-empty file trigger the limit, avoiding large in-memory data
6. **Test PNG data**: Minimal 8-byte PNG signature suffices for `http.DetectContentType` → `image/png`

### Files Created
- `backend/internal/repository/image.go` — ImageRepository with Create
- `backend/internal/repository/image_repository_test.go` — TestImageRepo_Create
- `backend/internal/handler/image.go` — ImageHandler with UploadImage
- `backend/internal/handler/image_test.go` — 5 test cases (Success, Unauthorized, NoTitle, FileTooLarge, InvalidType)
