package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"

	"gm_site/internal/middleware"
	"gm_site/internal/model"
	"gm_site/internal/repository"
	"gm_site/internal/service"
)

// ---------------------------------------------------------------------------
// Integration test setup
// ---------------------------------------------------------------------------

// setupIntegrationDB creates all tables needed for the full integration test.
func setupIntegrationDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	dsn := "file:" + dbPath + "?cache=shared&_journal_mode=WAL"

	db, err := sql.Open("sqlite", dsn)
	require.NoError(t, err, "failed to open test database")

	_, err = db.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			nickname TEXT NOT NULL DEFAULT '',
			role TEXT NOT NULL DEFAULT 'user' CHECK(role IN ('admin','user')),
			status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending','approved','rejected')),
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX idx_users_email ON users(email);
		CREATE INDEX idx_users_status ON users(status);

		CREATE TABLE albums (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			created_by INTEGER NOT NULL REFERENCES users(id),
			privacy TEXT NOT NULL DEFAULT 'public',
			is_friend_album INTEGER DEFAULT 0,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE images (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			album_id INTEGER REFERENCES albums(id) ON DELETE SET NULL,
			title TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			tags TEXT NOT NULL DEFAULT '[]',
			lsky_url TEXT NOT NULL,
			thumbnail_url TEXT NOT NULL DEFAULT '',
			uploaded_by INTEGER NOT NULL REFERENCES users(id),
			privacy TEXT NOT NULL DEFAULT 'public',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX idx_images_album_id ON images(album_id);
		CREATE INDEX idx_images_uploaded_by ON images(uploaded_by);
		CREATE INDEX idx_images_created_at ON images(created_at);

		CREATE TABLE comments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			image_id INTEGER NOT NULL REFERENCES images(id) ON DELETE CASCADE,
			user_id INTEGER NOT NULL REFERENCES users(id),
			content TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE notifications (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL REFERENCES users(id),
			type TEXT NOT NULL,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			related_id INTEGER,
			image_id INTEGER,
			is_read INTEGER DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)
	require.NoError(t, err, "failed to create tables")

	teardown := func() {
		db.Close()
	}

	return db, teardown
}

// integrationTestFixture holds all the dependencies for integration tests.
type integrationTestFixture struct {
	e       *echo.Echo
	db      *sql.DB
	jwtSvc  *service.JWTService
	lskyURL string
	lskySrv *httptest.Server
	cleanup func()
}

// setupIntegrationTest creates the full Echo server with all routes registered.
// Returns the fixture with Echo instance, DB, JWT service, Lsky mock URL, and cleanup.
func setupIntegrationTest(t *testing.T) *integrationTestFixture {
	t.Helper()

	db, teardownDB := setupIntegrationDB(t)

	// Repos
	userRepo := repository.NewUserRepository(db)
	albumRepo := repository.NewAlbumRepository(db)
	imageRepo := repository.NewImageRepository(db)
	commentRepo := repository.NewCommentRepository(db)
	notificationRepo := repository.NewNotificationRepository(db)

	// Services (mocked)
	jwtSvc := newTestJWTService()
	emailSvc := service.NewMockEmailService()

	// Lsky mock server
	lskySrv, lskyURL := setupLskyMock(t)
	lskyClient := service.NewLskyClient(lskyURL, "test@test.com", "password")

	// Handlers
	authHandler := NewAuthHandler(userRepo, jwtSvc, emailSvc, "")
	albumHandler := NewAlbumHandler(albumRepo)
	imageHandler := NewImageHandler(imageRepo, lskyClient, 10)
	commentHandler := NewCommentHandler(commentRepo, imageRepo, userRepo, notificationRepo, emailSvc, db)
	adminHandler := NewAdminHandler(userRepo, emailSvc)

	// Echo server
	e := echo.New()

	// ── Public routes (with optional auth) ──
	e.GET("/api/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})
	e.GET("/api/albums", albumHandler.ListAlbums, middleware.OptionalAuth(jwtSvc))
	e.GET("/api/images", imageHandler.ListImages, middleware.OptionalAuth(jwtSvc))
	e.GET("/api/images/search", imageHandler.SearchImages, middleware.OptionalAuth(jwtSvc))
	e.GET("/api/images/:id", imageHandler.GetImage)
	e.GET("/api/images/:id/comments", commentHandler.ListByImage)

	// ── Auth routes (no auth required) ──
	e.POST("/api/auth/register", authHandler.Register)
	e.POST("/api/auth/login", authHandler.Login)
	e.POST("/api/auth/refresh", authHandler.Refresh)

	// ── Protected routes (auth required) ──
	auth := e.Group("/api")
	auth.Use(middleware.AuthRequired(jwtSvc))
	auth.POST("/albums", albumHandler.CreateAlbum)
	auth.PUT("/albums/:id", albumHandler.UpdateAlbum)
	auth.DELETE("/albums/:id", albumHandler.DeleteAlbum)
	auth.POST("/images/upload", imageHandler.UploadImage)
	auth.PUT("/images/:id", imageHandler.UpdateImage)
	auth.DELETE("/images/:id", imageHandler.DeleteImage)
	auth.POST("/images/:id/comments", commentHandler.Create)
	auth.DELETE("/comments/:id", commentHandler.Delete)

	// ── Admin routes (auth + admin role) ──
	admin := e.Group("/api/admin")
	admin.Use(middleware.AuthRequired(jwtSvc), middleware.AdminRequired())
	admin.GET("/users/pending", adminHandler.ListPending)
	admin.PUT("/users/:id/approve", adminHandler.ApproveUser)
	admin.PUT("/users/:id/reject", adminHandler.RejectUser)

	cleanup := func() {
		lskySrv.Close()
		teardownDB()
	}

	return &integrationTestFixture{
		e:       e,
		db:      db,
		jwtSvc:  jwtSvc,
		lskyURL: lskyURL,
		lskySrv: lskySrv,
		cleanup: cleanup,
	}
}

// insertApprovedUser inserts a user with "approved" status and returns the model.
func insertApprovedUser(t *testing.T, db *sql.DB, email, password, nickname, role string) *model.User {
	t.Helper()
	return insertUser(t, db, email, password, nickname, role, model.UserStatusApproved)
}

// doJSONRequest performs a JSON request and returns the response recorder.
func doJSONRequest(e *echo.Echo, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

// doJSONRequestWithToken performs a JSON request with a Bearer token.
func doJSONRequestWithToken(e *echo.Echo, method, path, body, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

// parseData extracts the "data" field from a JSON API response.
func parseData(t *testing.T, rec *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	var result map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&result)
	require.NoError(t, err, "failed to decode response body")
	data, ok := result["data"].(map[string]interface{})
	if !ok {
		// data might be nil for error responses - return result as-is
		return result
	}
	return data
}

// ---------------------------------------------------------------------------
// 1. TestFullAuthFlow: register → (admin approve) → login → refresh token
// ---------------------------------------------------------------------------

func TestIntegration_FullAuthFlow(t *testing.T) {
	f := setupIntegrationTest(t)
	defer f.cleanup()

	// Step 1: Register a new user
	registerBody := `{"email":"alice@test.com","password":"pass123","nickname":"Alice"}`
	rec := doJSONRequest(f.e, http.MethodPost, "/api/auth/register", registerBody)
	assert.Equal(t, http.StatusCreated, rec.Code)

	// Step 2: Admin approves the user
	// First create an admin user in the DB (approved)
	admin := insertApprovedUser(t, f.db, "admin@test.com", "adminpass", "Admin", model.UserRoleAdmin)
	adminToken := generateAccessToken(t, f.jwtSvc, admin.ID, model.UserRoleAdmin)

	// List pending users
	rec = doJSONRequestWithToken(f.e, http.MethodGet, "/api/admin/users/pending", "", adminToken)
	assert.Equal(t, http.StatusOK, rec.Code)
	pendingData := parseData(t, rec)
	pendingList, ok := pendingData["list"].([]interface{})
	if !ok {
		// data might be the array directly
		pendingList, ok = pendingData["data"].([]interface{})
		if !ok {
			// Check if data itself is the list of users
			// parseData already extracts "data", so pendingData should be the list or have it
			t.Logf("pendingData keys: %v", pendingData)
		}
	}
	if pendingList == nil {
		// The API returns data directly as the list
		// Let's just approve the first pending user by finding them
		users, ok := pendingData["data"].([]interface{})
		if !ok {
			// data might be directly a list
			t.Log("pending response:", rec.Body.String())
		} else {
			t.Logf("found %d pending users", len(users))
		}
	}

	// Find alice's user ID from DB
	var aliceID int64
	err := f.db.QueryRow("SELECT id FROM users WHERE email = ?", "alice@test.com").Scan(&aliceID)
	require.NoError(t, err, "alice should exist in DB")

	// Approve alice
	rec = doJSONRequestWithToken(f.e, http.MethodPut, fmt.Sprintf("/api/admin/users/%d/approve", aliceID), "", adminToken)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Step 3: Login as alice
	loginBody := `{"email":"alice@test.com","password":"pass123"}`
	rec = doJSONRequest(f.e, http.MethodPost, "/api/auth/login", loginBody)
	assert.Equal(t, http.StatusOK, rec.Code)
	loginData := parseData(t, rec)
	assert.NotEmpty(t, loginData["access_token"], "should return access token")
	assert.NotEmpty(t, loginData["refresh_token"], "should return refresh token")

	accessToken := loginData["access_token"].(string)
	refreshToken := loginData["refresh_token"].(string)

	// Map keys are lowercase in the API response
	if accessToken == "" {
		// Try lowercase
		if at, ok := loginData["access_token"].(string); ok {
			accessToken = at
		}
	}
	if refreshToken == "" {
		if rt, ok := loginData["refresh_token"].(string); ok {
			refreshToken = rt
		}
	}
	require.NotEmpty(t, accessToken)
	require.NotEmpty(t, refreshToken)

	// Step 4: Refresh token
	refreshBody := fmt.Sprintf(`{"refresh_token":"%s"}`, refreshToken)
	rec = doJSONRequest(f.e, http.MethodPost, "/api/auth/refresh", refreshBody)
	assert.Equal(t, http.StatusOK, rec.Code)
	refreshData := parseData(t, rec)
	assert.NotEmpty(t, refreshData["access_token"], "should return new access token")
	assert.NotEmpty(t, refreshData["refresh_token"], "should return new refresh token")
}

// ---------------------------------------------------------------------------
// 2. TestImageCRUDFlow: login → upload image → list images → update → delete
// ---------------------------------------------------------------------------

func TestIntegration_ImageCRUDFlow(t *testing.T) {
	f := setupIntegrationTest(t)
	defer f.cleanup()

	// Create an approved user and log in
	user := insertApprovedUser(t, f.db, "uploader@test.com", "pass123", "Uploader", model.UserRoleUser)
	token := generateAccessToken(t, f.jwtSvc, user.ID, model.UserRoleUser)

	// Upload image
	body, ct := createMultipartBody("photo.png", testPNGData, map[string]string{
		"title": "My Photo",
		"tags":  "nature,travel",
	})
	rec := doMultipartRequest(f.e, http.MethodPost, "/api/images/upload", body, ct, token)
	assert.Equal(t, http.StatusCreated, rec.Code)
	uploadData := parseData(t, rec)
	assert.Equal(t, "My Photo", uploadData["title"])
	assert.Equal(t, "https://images.example.com/mock.png", uploadData["lsky_url"])
	assert.Equal(t, float64(user.ID), uploadData["uploaded_by"])

	imageID := uploadData["id"].(float64)
	require.NotZero(t, imageID)

	// List images (public endpoint - no auth needed)
	rec = doJSONRequest(f.e, http.MethodGet, "/api/images", "")
	assert.Equal(t, http.StatusOK, rec.Code)
	listData := parseData(t, rec)
	list, ok := listData["list"].([]interface{})
	require.True(t, ok, "response should contain a list")
	assert.GreaterOrEqual(t, len(list), 1)

	// Update image (own image)
	updateBody := `{"title":"Updated Photo","tags":["nature","travel","updated"]}`
	rec = doJSONRequestWithToken(f.e, http.MethodPut, fmt.Sprintf("/api/images/%.0f", imageID), updateBody, token)
	assert.Equal(t, http.StatusOK, rec.Code)
	updateData := parseData(t, rec)
	assert.Equal(t, "Updated Photo", updateData["title"])

	// Delete image
	rec = doJSONRequestWithToken(f.e, http.MethodDelete, fmt.Sprintf("/api/images/%.0f", imageID), "", token)
	assert.Equal(t, http.StatusOK, rec.Code)
	deleteData := parseData(t, rec)
	assert.Equal(t, true, deleteData["deleted"])
}

// ---------------------------------------------------------------------------
// 3. TestGuestBrowse: GET images without auth → 200
// ---------------------------------------------------------------------------

func TestIntegration_GuestBrowse(t *testing.T) {
	f := setupIntegrationTest(t)
	defer f.cleanup()

	// Insert a user and an image directly in the DB
	user := insertApprovedUser(t, f.db, "user@test.com", "pass", "User", model.UserRoleUser)

	// Insert an image directly
	imageRepo := repository.NewImageRepository(f.db)
	img := &model.Image{
		Title:      "Guest Test",
		Tags:       []string{"test"},
		LskyURL:    "https://example.com/img.png",
		UploadedBy: user.ID,
	}
	err := imageRepo.Create(img)
	require.NoError(t, err)

	// Public endpoints should work without auth
	t.Run("GET /api/images", func(t *testing.T) {
		rec := doJSONRequest(f.e, http.MethodGet, "/api/images", "")
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("GET /api/images/:id", func(t *testing.T) {
		rec := doJSONRequest(f.e, http.MethodGet, fmt.Sprintf("/api/images/%d", img.ID), "")
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("GET /api/health", func(t *testing.T) {
		rec := doJSONRequest(f.e, http.MethodGet, "/api/health", "")
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("GET /api/albums", func(t *testing.T) {
		rec := doJSONRequest(f.e, http.MethodGet, "/api/albums", "")
		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

// ---------------------------------------------------------------------------
// 4. TestPermissionBoundary: user A cannot delete user B's image → 403
// ---------------------------------------------------------------------------

func TestIntegration_PermissionBoundary(t *testing.T) {
	f := setupIntegrationTest(t)
	defer f.cleanup()

	// Create user A (the uploader)
	userA := insertApprovedUser(t, f.db, "alice@test.com", "pass", "Alice", model.UserRoleUser)

	// Create user B (the would-be attacker)
	userB := insertApprovedUser(t, f.db, "bob@test.com", "pass", "Bob", model.UserRoleUser)
	tokenB := generateAccessToken(t, f.jwtSvc, userB.ID, model.UserRoleUser)

	// User A uploads an image
	tokenA := generateAccessToken(t, f.jwtSvc, userA.ID, model.UserRoleUser)
	body, ct := createMultipartBody("alice.png", testPNGData, map[string]string{
		"title": "Alice Photo",
	})
	rec := doMultipartRequest(f.e, http.MethodPost, "/api/images/upload", body, ct, tokenA)
	assert.Equal(t, http.StatusCreated, rec.Code)
	uploadData := parseData(t, rec)
	imageID := uploadData["id"].(float64)

	// User B tries to delete user A's image → 403
	rec = doJSONRequestWithToken(f.e, http.MethodDelete, fmt.Sprintf("/api/images/%.0f", imageID), "", tokenB)
	assert.Equal(t, http.StatusForbidden, rec.Code)

	// User B tries to update user A's image → 403
	updateBody := `{"title":"Hacked!"}`
	rec = doJSONRequestWithToken(f.e, http.MethodPut, fmt.Sprintf("/api/images/%.0f", imageID), updateBody, tokenB)
	assert.Equal(t, http.StatusForbidden, rec.Code)

	// User A updates own image → 200 (verify still works)
	updateBody = `{"title":"Still Mine"}`
	rec = doJSONRequestWithToken(f.e, http.MethodPut, fmt.Sprintf("/api/images/%.0f", imageID), updateBody, tokenA)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Admin can delete user A's image
	admin := insertApprovedUser(t, f.db, "admin@test.com", "pass", "Admin", model.UserRoleAdmin)
	adminToken := generateAccessToken(t, f.jwtSvc, admin.ID, model.UserRoleAdmin)
	rec = doJSONRequestWithToken(f.e, http.MethodDelete, fmt.Sprintf("/api/images/%.0f", imageID), "", adminToken)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// ---------------------------------------------------------------------------
// 5. TestAdminFlow: admin login → list pending → approve → user can login
// ---------------------------------------------------------------------------

func TestIntegration_AdminFlow(t *testing.T) {
	f := setupIntegrationTest(t)
	defer f.cleanup()

	// Create an admin user (pre-approved)
	admin := insertApprovedUser(t, f.db, "admin@test.com", "adminpass", "Admin", model.UserRoleAdmin)
	adminToken := generateAccessToken(t, f.jwtSvc, admin.ID, model.UserRoleAdmin)

	// Register a pending user via API
	rec := doJSONRequest(f.e, http.MethodPost, "/api/auth/register", `{"email":"newuser@test.com","password":"pass123","nickname":"NewUser"}`)
	assert.Equal(t, http.StatusCreated, rec.Code)

	// New user cannot login yet (pending)
	rec = doJSONRequest(f.e, http.MethodPost, "/api/auth/login", `{"email":"newuser@test.com","password":"pass123"}`)
	assert.Equal(t, http.StatusForbidden, rec.Code)

	// Admin lists pending users
	rec = doJSONRequestWithToken(f.e, http.MethodGet, "/api/admin/users/pending", "", adminToken)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Find the pending user's ID
	var pendingUserID int64
	err := f.db.QueryRow("SELECT id FROM users WHERE email = ? AND status = ?", "newuser@test.com", model.UserStatusPending).Scan(&pendingUserID)
	require.NoError(t, err)

	// Admin approves the user
	rec = doJSONRequestWithToken(f.e, http.MethodPut, fmt.Sprintf("/api/admin/users/%d/approve", pendingUserID), "", adminToken)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Now the user can login
	rec = doJSONRequest(f.e, http.MethodPost, "/api/auth/login", `{"email":"newuser@test.com","password":"pass123"}`)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// ---------------------------------------------------------------------------
// 6. TestSearchFlow: upload image with tags → search by tag → finds image
// ---------------------------------------------------------------------------

func TestIntegration_SearchFlow(t *testing.T) {
	f := setupIntegrationTest(t)
	defer f.cleanup()

	// Create user and upload an image with tags
	user := insertApprovedUser(t, f.db, "searcher@test.com", "pass", "Searcher", model.UserRoleUser)
	token := generateAccessToken(t, f.jwtSvc, user.ID, model.UserRoleUser)

	// Upload image with specific tags
	body, ct := createMultipartBody("beach.png", testPNGData, map[string]string{
		"title": "Sunset Beach",
		"tags":  "sunset,beach,vacation",
	})
	rec := doMultipartRequest(f.e, http.MethodPost, "/api/images/upload", body, ct, token)
	assert.Equal(t, http.StatusCreated, rec.Code)

	// Upload another image without those tags
	body2, ct2 := createMultipartBody("city.png", testPNGData, map[string]string{
		"title": "City Night",
		"tags":  "city,night,urban",
	})
	rec = doMultipartRequest(f.e, http.MethodPost, "/api/images/upload", body2, ct2, token)
	assert.Equal(t, http.StatusCreated, rec.Code)

	// Search by tag "beach" — should find the first image
	rec = doJSONRequest(f.e, http.MethodGet, "/api/images/search?q=beach", "")
	assert.Equal(t, http.StatusOK, rec.Code)

	searchData := parseData(t, rec)
	images, ok := searchData["images"].([]interface{})
	require.True(t, ok, "search response should contain images array")
	assert.GreaterOrEqual(t, len(images), 1, "should find at least one image matching 'beach'")

	// Search by "sunset" — should also find the first image
	rec = doJSONRequest(f.e, http.MethodGet, "/api/images/search?q=sunset", "")
	assert.Equal(t, http.StatusOK, rec.Code)
	searchData = parseData(t, rec)
	images, ok = searchData["images"].([]interface{})
	require.True(t, ok)
	assert.GreaterOrEqual(t, len(images), 1)

	// Search by "nonexistent" — should return empty results
	rec = doJSONRequest(f.e, http.MethodGet, "/api/images/search?q=nonexistent", "")
	assert.Equal(t, http.StatusOK, rec.Code)
	searchData = parseData(t, rec)
	images, ok = searchData["images"].([]interface{})
	require.True(t, ok)
	assert.Equal(t, 0, len(images), "should find no images matching 'nonexistent'")
}
