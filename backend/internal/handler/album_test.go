package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"testing"
	"time"

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
// Test helpers
// ---------------------------------------------------------------------------

// setupAlbumTestDB creates a temporary SQLite database with users, albums, and
// images tables, and returns the db handle and a teardown function.
func setupAlbumTestDB(t *testing.T) (*sql.DB, func()) {
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
			role TEXT NOT NULL DEFAULT 'user',
			status TEXT NOT NULL DEFAULT 'pending',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
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
		CREATE TABLE IF NOT EXISTS friends (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL REFERENCES users(id),
			friend_id INTEGER NOT NULL REFERENCES users(id),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, friend_id)
		);
	`)
	require.NoError(t, err, "failed to create tables")

	teardown := func() {
		db.Close()
	}

	return db, teardown
}

// insertAlbum inserts a test album and returns the populated model.
func insertAlbum(t *testing.T, db *sql.DB, name, description string, createdBy int64) *model.Album {
	t.Helper()

	now := time.Now()
	result, err := db.Exec(
		`INSERT INTO albums (name, description, created_by, privacy, is_friend_album, created_at)
		 VALUES (?, ?, ?, 'public', 0, ?)`,
		name, description, createdBy, now,
	)
	require.NoError(t, err, "failed to insert test album")

	id, err := result.LastInsertId()
	require.NoError(t, err, "failed to get last insert id")

	return &model.Album{
		ID:          id,
		Name:        name,
		Description: description,
		CreatedBy:   createdBy,
		CreatedAt:   now,
	}
}

// insertAlbumImage inserts a test image associated with an album.
func insertAlbumImage(t *testing.T, db *sql.DB, albumID int64, uploadedBy int64) {
	t.Helper()

	now := time.Now()
	_, err := db.Exec(
		`INSERT INTO images (album_id, title, description, tags, lsky_url, thumbnail_url, uploaded_by, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		albumID, "test image", "", "[]", "https://example.com/image.jpg", "", uploadedBy, now, now,
	)
	require.NoError(t, err, "failed to insert test image")
}

// setupAlbumTest creates all dependencies and registers album routes.
func setupAlbumTest(t *testing.T) (*echo.Echo, *AlbumHandler, *sql.DB, *service.JWTService, func()) {
	t.Helper()

	db, teardownDB := setupAlbumTestDB(t)
	albumRepo := repository.NewAlbumRepository(db)
	jwtSvc := newTestJWTService()
	handler := NewAlbumHandler(albumRepo)

	e := echo.New()

	// Public routes with optional auth
	e.GET("/api/albums", handler.ListAlbums, middleware.OptionalAuth(jwtSvc))

	// Authenticated routes
	g := e.Group("/api/albums", middleware.AuthRequired(jwtSvc))
	g.POST("", handler.CreateAlbum)
	g.PUT("/:id", handler.UpdateAlbum)
	g.DELETE("/:id", handler.DeleteAlbum)

	return e, handler, db, jwtSvc, teardownDB
}

// ---------------------------------------------------------------------------
// ListAlbums tests
// ---------------------------------------------------------------------------

func TestListAlbums(t *testing.T) {
	e, _, db, _, teardown := setupAlbumTest(t)
	defer teardown()

	// Insert a user and some albums
	user := insertUser(t, db, "user@test.com", "pass123", "User", model.UserRoleUser, model.UserStatusApproved)
	insertAlbum(t, db, "Album 1", "First album", user.ID)
	insertAlbum(t, db, "Album 2", "Second album", user.ID)

	rec := doRequest(e, http.MethodGet, "/api/albums", "")
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, float64(http.StatusOK), resp["code"])

	data, ok := resp["data"].([]interface{})
	require.True(t, ok, "data should be an array")
	assert.Len(t, data, 2)

	// Verify the most recent album is first (descending order)
	first, ok := data[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Album 2", first["name"])
}

// ---------------------------------------------------------------------------
// CreateAlbum tests
// ---------------------------------------------------------------------------

func TestCreateAlbum_Success(t *testing.T) {
	e, _, db, jwtSvc, teardown := setupAlbumTest(t)
	defer teardown()

	user := insertUser(t, db, "user@test.com", "pass123", "User", model.UserRoleUser, model.UserStatusApproved)
	token := generateAccessToken(t, jwtSvc, user.ID, model.UserRoleUser)

	body := `{"name":"New Album","description":"A test album"}`
	rec := doAuthRequest(e, http.MethodPost, "/api/albums", body, token)
	assert.Equal(t, http.StatusCreated, rec.Code)

	result := parseResponse(t, rec)
	assert.Equal(t, float64(http.StatusCreated), result["code"])
	assert.Equal(t, "created", result["message"])

	data, ok := result["data"].(map[string]interface{})
	require.True(t, ok, "data should be a map")
	assert.Equal(t, "New Album", data["name"])
	assert.Equal(t, "A test album", data["description"])
	assert.Equal(t, float64(user.ID), data["created_by"])
	assert.NotZero(t, data["id"])
}

func TestCreateAlbum_Unauthorized(t *testing.T) {
	e, _, _, _, teardown := setupAlbumTest(t)
	defer teardown()

	body := `{"name":"New Album","description":"A test album"}`
	rec := doAuthRequest(e, http.MethodPost, "/api/albums", body, "")
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// ---------------------------------------------------------------------------
// UpdateAlbum tests
// ---------------------------------------------------------------------------

func TestUpdateAlbum_Owner(t *testing.T) {
	e, _, db, jwtSvc, teardown := setupAlbumTest(t)
	defer teardown()

	user := insertUser(t, db, "user@test.com", "pass123", "User", model.UserRoleUser, model.UserStatusApproved)
	album := insertAlbum(t, db, "Original", "Original description", user.ID)
	token := generateAccessToken(t, jwtSvc, user.ID, model.UserRoleUser)

	body := `{"name":"Updated","description":"Updated description"}`
	rec := doAuthRequest(e, http.MethodPut, "/api/albums/"+fmt.Sprintf("%d", album.ID), body, token)
	assert.Equal(t, http.StatusOK, rec.Code)

	result := parseResponse(t, rec)
	assert.Equal(t, float64(http.StatusOK), result["code"])

	data, ok := result["data"].(map[string]interface{})
	require.True(t, ok, "data should be a map")
	assert.Equal(t, "Updated", data["name"])
	assert.Equal(t, "Updated description", data["description"])
}

func TestUpdateAlbum_NotOwner(t *testing.T) {
	e, _, db, jwtSvc, teardown := setupAlbumTest(t)
	defer teardown()

	owner := insertUser(t, db, "owner@test.com", "pass123", "Owner", model.UserRoleUser, model.UserStatusApproved)
	other := insertUser(t, db, "other@test.com", "pass123", "Other", model.UserRoleUser, model.UserStatusApproved)
	album := insertAlbum(t, db, "Original", "Original description", owner.ID)
	token := generateAccessToken(t, jwtSvc, other.ID, model.UserRoleUser)

	body := `{"name":"Hacked","description":"Should not work"}`
	rec := doAuthRequest(e, http.MethodPut, "/api/albums/"+fmt.Sprintf("%d", album.ID), body, token)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestUpdateAlbum_Admin(t *testing.T) {
	e, _, db, jwtSvc, teardown := setupAlbumTest(t)
	defer teardown()

	user := insertUser(t, db, "user@test.com", "pass123", "User", model.UserRoleUser, model.UserStatusApproved)
	admin := insertUser(t, db, "admin@test.com", "admin123", "Admin", model.UserRoleAdmin, model.UserStatusApproved)
	album := insertAlbum(t, db, "Original", "Original description", user.ID)
	token := generateAccessToken(t, jwtSvc, admin.ID, model.UserRoleAdmin)

	body := `{"name":"Admin Updated","description":"Admin can edit any album"}`
	rec := doAuthRequest(e, http.MethodPut, "/api/albums/"+fmt.Sprintf("%d", album.ID), body, token)
	assert.Equal(t, http.StatusOK, rec.Code)

	result := parseResponse(t, rec)
	assert.Equal(t, float64(http.StatusOK), result["code"])

	data, ok := result["data"].(map[string]interface{})
	require.True(t, ok, "data should be a map")
	assert.Equal(t, "Admin Updated", data["name"])
	assert.Equal(t, "Admin can edit any album", data["description"])
}

// ---------------------------------------------------------------------------
// DeleteAlbum tests
// ---------------------------------------------------------------------------

func TestDeleteAlbum_WithImages(t *testing.T) {
	e, _, db, jwtSvc, teardown := setupAlbumTest(t)
	defer teardown()

	user := insertUser(t, db, "user@test.com", "pass123", "User", model.UserRoleUser, model.UserStatusApproved)
	album := insertAlbum(t, db, "Album with images", "Has images", user.ID)
	insertAlbumImage(t, db, album.ID, user.ID)

	token := generateAccessToken(t, jwtSvc, user.ID, model.UserRoleUser)
	rec := doAuthRequest(e, http.MethodDelete, "/api/albums/"+fmt.Sprintf("%d", album.ID), "", token)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	result := parseResponse(t, rec)
	assert.Equal(t, "相册下有图片，无法删除", result["message"])
}

func TestDeleteAlbum_Empty(t *testing.T) {
	e, _, db, jwtSvc, teardown := setupAlbumTest(t)
	defer teardown()

	user := insertUser(t, db, "user@test.com", "pass123", "User", model.UserRoleUser, model.UserStatusApproved)
	album := insertAlbum(t, db, "Empty album", "No images", user.ID)

	token := generateAccessToken(t, jwtSvc, user.ID, model.UserRoleUser)
	rec := doAuthRequest(e, http.MethodDelete, "/api/albums/"+fmt.Sprintf("%d", album.ID), "", token)
	assert.Equal(t, http.StatusOK, rec.Code)

	result := parseResponse(t, rec)
	assert.Equal(t, float64(http.StatusOK), result["code"])
}
