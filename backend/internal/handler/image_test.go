package handler

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
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
// Test data & helpers
// ---------------------------------------------------------------------------

// testPNGData is a minimal valid PNG byte sequence (8-byte PNG signature + padding)
// that passes http.DetectContentType as "image/png".
var testPNGData = []byte{
	0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
}

// createMultipartBody builds a multipart/form-data request body with a file and fields.
func createMultipartBody(fileName string, fileData []byte, fields map[string]string) (*bytes.Buffer, string) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	for key, val := range fields {
		w.WriteField(key, val)
	}

	part, _ := w.CreateFormFile("file", fileName)
	part.Write(fileData)

	w.Close()
	return &buf, w.FormDataContentType()
}

// doMultipartRequest performs an HTTP request with multipart body and optional auth.
func doMultipartRequest(e *echo.Echo, method, path string, body *bytes.Buffer, contentType, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, body)
	req.Header.Set("Content-Type", contentType)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

// ---------------------------------------------------------------------------
// Database & dependency setup
// ---------------------------------------------------------------------------

// setupImageHandlerDB creates tables for users and images.
func setupImageHandlerDB(t *testing.T) (*sql.DB, func()) {
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
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE comments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			image_id INTEGER NOT NULL REFERENCES images(id) ON DELETE CASCADE,
			user_id INTEGER NOT NULL REFERENCES users(id),
			content TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`)
	require.NoError(t, err, "failed to create tables")

	teardown := func() {
		db.Close()
	}

	return db, teardown
}

// setupLskyMock creates an httptest server that mimics the Lsky Pro API
// (token issuance and file upload). Returns the server and its URL.
func setupLskyMock(t *testing.T) (*httptest.Server, string) {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/tokens":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status": true,
				"data": map[string]interface{}{
					"token": "mock-upload-token",
				},
			})
		case "/api/v1/upload":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status": true,
				"data": map[string]interface{}{
					"links": map[string]interface{}{
						"url": "https://images.example.com/mock.png",
					},
				},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	return server, server.URL
}

// setupImageTest sets up Echo, ImageHandler, DB, JWT service and returns a teardown.
func setupImageTest(t *testing.T, maxSizeMB int) (*echo.Echo, *ImageHandler, *sql.DB, *service.JWTService, func()) {
	t.Helper()

	db, teardownDB := setupImageHandlerDB(t)
	lskyServer, lskyURL := setupLskyMock(t)

	lskyClient := service.NewLskyClient(lskyURL, "test@test.com", "password")
	imageRepo := repository.NewImageRepository(db)
	jwtSvc := newTestJWTService()
	handler := NewImageHandler(imageRepo, lskyClient, maxSizeMB)

	e := echo.New()

	// Public routes
	e.GET("/api/images/search", handler.SearchImages)
	e.GET("/api/images", handler.ListImages)
	e.GET("/api/images/:id", handler.GetImage)

	// Authenticated routes
	g := e.Group("/api/images", middleware.AuthRequired(jwtSvc))
	g.POST("/upload", handler.UploadImage)
	g.PUT("/:id", handler.UpdateImage)
	g.DELETE("/:id", handler.DeleteImage)

	teardown := func() {
		lskyServer.Close()
		teardownDB()
	}

	return e, handler, db, jwtSvc, teardown
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestUploadImage_Success(t *testing.T) {
	e, _, db, jwtSvc, teardown := setupImageTest(t, 10)
	defer teardown()

	user := insertUser(t, db, "uploader@test.com", "pass123", "Uploader", model.UserRoleUser, model.UserStatusApproved)
	token := generateAccessToken(t, jwtSvc, user.ID, model.UserRoleUser)

	body, ct := createMultipartBody("photo.png", testPNGData, map[string]string{
		"title": "My Photo",
		"tags":  "nature,travel",
	})

	rec := doMultipartRequest(e, http.MethodPost, "/api/images/upload", body, ct, token)
	assert.Equal(t, http.StatusCreated, rec.Code)

	result := parseResponse(t, rec)
	assert.Equal(t, float64(http.StatusCreated), result["code"])

	data, ok := result["data"].(map[string]interface{})
	require.True(t, ok, "data should be a map")
	assert.Equal(t, "My Photo", data["title"])
	assert.Equal(t, "https://images.example.com/mock.png", data["lsky_url"])
	assert.Equal(t, float64(user.ID), data["uploaded_by"])

	// Verify tags
	tags, ok := data["tags"].([]interface{})
	require.True(t, ok, "tags should be an array")
	assert.Len(t, tags, 2)
}

func TestUploadImage_Unauthorized(t *testing.T) {
	e, _, _, _, teardown := setupImageTest(t, 10)
	defer teardown()

	body, ct := createMultipartBody("photo.png", testPNGData, map[string]string{
		"title": "My Photo",
	})

	rec := doMultipartRequest(e, http.MethodPost, "/api/images/upload", body, ct, "")
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestUploadImage_NoTitle(t *testing.T) {
	e, _, db, jwtSvc, teardown := setupImageTest(t, 10)
	defer teardown()

	user := insertUser(t, db, "uploader2@test.com", "pass123", "Uploader2", model.UserRoleUser, model.UserStatusApproved)
	token := generateAccessToken(t, jwtSvc, user.ID, model.UserRoleUser)

	// No title field in the form
	body, ct := createMultipartBody("photo.png", testPNGData, map[string]string{})

	rec := doMultipartRequest(e, http.MethodPost, "/api/images/upload", body, ct, token)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	result := parseResponse(t, rec)
	assert.Equal(t, float64(http.StatusBadRequest), result["code"])
}

func TestUploadImage_FileTooLarge(t *testing.T) {
	// Use maxSizeMB=0 so any non-empty file triggers the limit
	e, _, db, jwtSvc, teardown := setupImageTest(t, 0)
	defer teardown()

	user := insertUser(t, db, "uploader3@test.com", "pass123", "Uploader3", model.UserRoleUser, model.UserStatusApproved)
	token := generateAccessToken(t, jwtSvc, user.ID, model.UserRoleUser)

	body, ct := createMultipartBody("big.jpg", []byte("small"), map[string]string{
		"title": "Big File",
	})

	rec := doMultipartRequest(e, http.MethodPost, "/api/images/upload", body, ct, token)
	assert.Equal(t, http.StatusRequestEntityTooLarge, rec.Code)
}

func TestUploadImage_InvalidType(t *testing.T) {
	e, _, db, jwtSvc, teardown := setupImageTest(t, 10)
	defer teardown()

	user := insertUser(t, db, "uploader4@test.com", "pass123", "Uploader4", model.UserRoleUser, model.UserStatusApproved)
	token := generateAccessToken(t, jwtSvc, user.ID, model.UserRoleUser)

	// File content that starts with text, not image magic bytes
	textData := []byte("This is not an image file at all")
	body, ct := createMultipartBody("readme.txt", textData, map[string]string{
		"title": "Fake Image",
	})

	rec := doMultipartRequest(e, http.MethodPost, "/api/images/upload", body, ct, token)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// ---------------------------------------------------------------------------
// Search tests
// ---------------------------------------------------------------------------

// insertImage inserts a test image directly into the database and returns the image model.
func insertImage(t *testing.T, db *sql.DB, uploadedBy int64, title, description string, tags []string, lskyURL string) *model.Image {
	t.Helper()

	tagsJSON, err := json.Marshal(tags)
	require.NoError(t, err, "failed to marshal tags")

	now := time.Now()
	result, err := db.Exec(
		`INSERT INTO images (title, description, tags, lsky_url, uploaded_by, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		title, description, string(tagsJSON), lskyURL, uploadedBy, now, now,
	)
	require.NoError(t, err, "failed to insert test image")

	id, err := result.LastInsertId()
	require.NoError(t, err, "failed to get last insert id")

	return &model.Image{
		ID:         id,
		Title:      title,
		Description: description,
		Tags:       tags,
		LskyURL:    lskyURL,
		UploadedBy: uploadedBy,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// insertImageWithAlbum inserts a test image with an album_id and returns the image model.
func insertImageWithAlbum(t *testing.T, db *sql.DB, uploadedBy, albumID int64, title string, tags []string, lskyURL string) *model.Image {
	t.Helper()

	tagsJSON, err := json.Marshal(tags)
	require.NoError(t, err, "failed to marshal tags")

	now := time.Now()
	result, err := db.Exec(
		`INSERT INTO images (album_id, title, tags, lsky_url, uploaded_by, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		albumID, title, string(tagsJSON), lskyURL, uploadedBy, now, now,
	)
	require.NoError(t, err, "failed to insert test image")

	id, err := result.LastInsertId()
	require.NoError(t, err, "failed to get last insert id")

	return &model.Image{
		ID:         id,
		AlbumID:    &albumID,
		Title:      title,
		Tags:       tags,
		LskyURL:    lskyURL,
		UploadedBy: uploadedBy,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

func TestSearchImages_ByTitle(t *testing.T) {
	e, _, db, _, teardown := setupImageTest(t, 10)
	defer teardown()

	user := insertUser(t, db, "search-title@test.com", "pass123", "TitleSearcher", model.UserRoleUser, model.UserStatusApproved)

	// Insert images with varying titles
	insertImage(t, db, user.ID, "Beautiful Sunset", "A nice sunset photo", []string{"nature", "sunset"}, "https://img.example.com/1.jpg")
	insertImage(t, db, user.ID, "Sunset Beach", "Beach view", []string{"travel"}, "https://img.example.com/2.jpg")
	insertImage(t, db, user.ID, "City Life", "Urban landscape", []string{"city"}, "https://img.example.com/3.jpg")

	rec := doRequest(e, http.MethodGet, "/api/images/search?q=Sunset", "")
	assert.Equal(t, http.StatusOK, rec.Code)

	result := parseResponse(t, rec)
	assert.Equal(t, float64(http.StatusOK), result["code"])

	data, ok := result["data"].(map[string]interface{})
	require.True(t, ok, "data should be a map")

	images, ok := data["images"].([]interface{})
	require.True(t, ok, "images should be an array")
	assert.Len(t, images, 2, "should find 2 images with 'Sunset' in title")

	total, ok := data["total"].(float64)
	require.True(t, ok)
	assert.Equal(t, float64(2), total)
}

func TestSearchImages_ByTag(t *testing.T) {
	e, _, db, _, teardown := setupImageTest(t, 10)
	defer teardown()

	user := insertUser(t, db, "search-tag@test.com", "pass123", "TagSearcher", model.UserRoleUser, model.UserStatusApproved)

	insertImage(t, db, user.ID, "Mountain View", "A mountain panorama", []string{"nature", "mountain"}, "https://img.example.com/mt.jpg")
	insertImage(t, db, user.ID, "Ocean View", "Wide ocean scenery", []string{"nature", "ocean"}, "https://img.example.com/oc.jpg")
	insertImage(t, db, user.ID, "Downtown", "City buildings", []string{"city", "architecture"}, "https://img.example.com/dt.jpg")

	// Search by tag "ocean" (stored as part of JSON array text)
	rec := doRequest(e, http.MethodGet, "/api/images/search?q=ocean", "")
	assert.Equal(t, http.StatusOK, rec.Code)

	result := parseResponse(t, rec)
	assert.Equal(t, float64(http.StatusOK), result["code"])

	data, ok := result["data"].(map[string]interface{})
	require.True(t, ok, "data should be a map")

	images, ok := data["images"].([]interface{})
	require.True(t, ok, "images should be an array")
	assert.Len(t, images, 1, "should find 1 image with 'ocean' tag")
}

func TestSearchImages_NoResults(t *testing.T) {
	e, _, db, _, teardown := setupImageTest(t, 10)
	defer teardown()

	user := insertUser(t, db, "search-none@test.com", "pass123", "NoResults", model.UserRoleUser, model.UserStatusApproved)
	insertImage(t, db, user.ID, "Something", "Some description", []string{"tag1"}, "https://img.example.com/s.jpg")

	// Search for something that doesn't exist
	rec := doRequest(e, http.MethodGet, "/api/images/search?q=NonExistentXYZ", "")
	assert.Equal(t, http.StatusOK, rec.Code)

	result := parseResponse(t, rec)
	assert.Equal(t, float64(http.StatusOK), result["code"])

	data, ok := result["data"].(map[string]interface{})
	require.True(t, ok, "data should be a map")

	images, ok := data["images"].([]interface{})
	require.True(t, ok, "images should be an array")
	assert.Len(t, images, 0, "should return empty array when no results match")

	total, ok := data["total"].(float64)
	require.True(t, ok)
	assert.Equal(t, float64(0), total)
}

func TestSearchImages_Paginated(t *testing.T) {
	e, _, db, _, teardown := setupImageTest(t, 10)
	defer teardown()

	user := insertUser(t, db, "search-page@test.com", "pass123", "Paginated", model.UserRoleUser, model.UserStatusApproved)

	// Insert 5 images all sharing the word "Photo"
	for i := 1; i <= 5; i++ {
		title := fmt.Sprintf("Photo Number %d", i)
		insertImage(t, db, user.ID, title, "A test photo", []string{"test"}, fmt.Sprintf("https://img.example.com/p%d.jpg", i))
	}

	// Page 1 with limit 2
	rec := doRequest(e, http.MethodGet, "/api/images/search?q=Photo&page=1&limit=2", "")
	assert.Equal(t, http.StatusOK, rec.Code)

	result := parseResponse(t, rec)
	data, ok := result["data"].(map[string]interface{})
	require.True(t, ok, "data should be a map")

	images, ok := data["images"].([]interface{})
	require.True(t, ok, "images should be an array")
	assert.Len(t, images, 2, "page 1 should have 2 images")
	assert.Equal(t, float64(5), data["total"], "total should be 5")
	assert.Equal(t, float64(1), data["page"], "page should be 1")
	assert.Equal(t, float64(2), data["limit"], "limit should be 2")

	// Page 3 with limit 2 (should return 1 image)
	rec2 := doRequest(e, http.MethodGet, "/api/images/search?q=Photo&page=3&limit=2", "")
	assert.Equal(t, http.StatusOK, rec2.Code)

	result2 := parseResponse(t, rec2)
	data2, ok := result2["data"].(map[string]interface{})
	require.True(t, ok, "data should be a map")

	images2, ok := data2["images"].([]interface{})
	require.True(t, ok, "images should be an array")
	assert.Len(t, images2, 1, "page 3 should have 1 remaining image")
	assert.Equal(t, float64(5), data2["total"], "total should still be 5")
}

// ---------------------------------------------------------------------------
// CRUD tests — List, Get, Update, Delete
// ---------------------------------------------------------------------------

func TestListImages_Paginated(t *testing.T) {
	e, _, db, _, teardown := setupImageTest(t, 10)
	defer teardown()

	user := insertUser(t, db, "list-page@test.com", "pass123", "PageUser", model.UserRoleUser, model.UserStatusApproved)

	// Insert 5 images
	for i := 1; i <= 5; i++ {
		insertImage(t, db, user.ID, fmt.Sprintf("Photo %d", i), "", []string{"test"}, fmt.Sprintf("https://img.example.com/p%d.jpg", i))
	}

	// Page 1, limit 2
	rec := doRequest(e, http.MethodGet, "/api/images?page=1&limit=2", "")
	assert.Equal(t, http.StatusOK, rec.Code)

	result := parseResponse(t, rec)
	assert.Equal(t, float64(http.StatusOK), result["code"])

	data, ok := result["data"].(map[string]interface{})
	require.True(t, ok, "data should be a map")

	list, ok := data["list"].([]interface{})
	require.True(t, ok, "list should be an array")
	assert.Len(t, list, 2, "page 1 should have 2 images")
	assert.Equal(t, float64(5), data["total"], "total should be 5")
	assert.Equal(t, float64(1), data["page"], "page should be 1")
	assert.Equal(t, float64(2), data["limit"], "limit should be 2")

	// Page 3, limit 2 (should return 1 image)
	rec2 := doRequest(e, http.MethodGet, "/api/images?page=3&limit=2", "")
	assert.Equal(t, http.StatusOK, rec2.Code)

	result2 := parseResponse(t, rec2)
	data2, ok := result2["data"].(map[string]interface{})
	require.True(t, ok)

	list2, ok := data2["list"].([]interface{})
	require.True(t, ok)
	assert.Len(t, list2, 1, "page 3 should have 1 image")
}

func TestListImages_FilterByAlbum(t *testing.T) {
	e, _, db, _, teardown := setupImageTest(t, 10)
	defer teardown()

	user := insertUser(t, db, "album-filter@test.com", "pass123", "AlbumFilter", model.UserRoleUser, model.UserStatusApproved)

	// Insert an album
	album := insertAlbum(t, db, "My Album", "test album", user.ID)

	// Insert images: 2 in the album, 1 without
	insertImageWithAlbum(t, db, user.ID, album.ID, "Album Photo 1", []string{"album"}, "https://img.example.com/a1.jpg")
	insertImageWithAlbum(t, db, user.ID, album.ID, "Album Photo 2", []string{"album"}, "https://img.example.com/a2.jpg")
	insertImage(t, db, user.ID, "No Album Photo", "", []string{"noalbum"}, "https://img.example.com/na.jpg")

	rec := doRequest(e, http.MethodGet, fmt.Sprintf("/api/images?album_id=%d", album.ID), "")
	assert.Equal(t, http.StatusOK, rec.Code)

	result := parseResponse(t, rec)
	data, ok := result["data"].(map[string]interface{})
	require.True(t, ok)

	list, ok := data["list"].([]interface{})
	require.True(t, ok)
	assert.Len(t, list, 2, "should find 2 images in the album")
	assert.Equal(t, float64(2), data["total"])
}

func TestListImages_FilterByTag(t *testing.T) {
	e, _, db, _, teardown := setupImageTest(t, 10)
	defer teardown()

	user := insertUser(t, db, "tag-filter@test.com", "pass123", "TagFilter", model.UserRoleUser, model.UserStatusApproved)

	insertImage(t, db, user.ID, "Sunset Mountain", "", []string{"顾夏", "mountain"}, "https://img.example.com/sm.jpg")
	insertImage(t, db, user.ID, "Ocean View", "", []string{"顾夏", "ocean"}, "https://img.example.com/ov.jpg")
	insertImage(t, db, user.ID, "City Life", "", []string{"city"}, "https://img.example.com/cl.jpg")

	rec := doRequest(e, http.MethodGet, "/api/images?tag=顾夏", "")
	assert.Equal(t, http.StatusOK, rec.Code)

	result := parseResponse(t, rec)
	data, ok := result["data"].(map[string]interface{})
	require.True(t, ok)

	list, ok := data["list"].([]interface{})
	require.True(t, ok)
	assert.Len(t, list, 2, "should find 2 images with tag 顾夏")
	assert.Equal(t, float64(2), data["total"])
}

func TestGetImage_Detail(t *testing.T) {
	e, _, db, _, teardown := setupImageTest(t, 10)
	defer teardown()

	user := insertUser(t, db, "detail@test.com", "pass123", "DetailUser", model.UserRoleUser, model.UserStatusApproved)
	img := insertImage(t, db, user.ID, "Detail Photo", "A detailed photo", []string{"nature"}, "https://img.example.com/detail.jpg")

	// Insert comments for this image
	_, err := db.Exec(`INSERT INTO comments (image_id, user_id, content, created_at) VALUES (?, ?, ?, ?)`,
		img.ID, user.ID, "Comment 1", time.Now())
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO comments (image_id, user_id, content, created_at) VALUES (?, ?, ?, ?)`,
		img.ID, user.ID, "Comment 2", time.Now())
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO comments (image_id, user_id, content, created_at) VALUES (?, ?, ?, ?)`,
		img.ID, user.ID, "Comment 3", time.Now())
	require.NoError(t, err)

	rec := doRequest(e, http.MethodGet, fmt.Sprintf("/api/images/%d", img.ID), "")
	assert.Equal(t, http.StatusOK, rec.Code)

	result := parseResponse(t, rec)
	assert.Equal(t, float64(http.StatusOK), result["code"])

	data, ok := result["data"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Detail Photo", data["title"])
	assert.Equal(t, "A detailed photo", data["description"])
	assert.Equal(t, float64(3), data["comment_count"], "should have 3 comments")
}

func TestGetImage_NotFound(t *testing.T) {
	e, _, _, _, teardown := setupImageTest(t, 10)
	defer teardown()

	rec := doRequest(e, http.MethodGet, "/api/images/99999", "")
	assert.Equal(t, http.StatusNotFound, rec.Code)

	result := parseResponse(t, rec)
	assert.Equal(t, float64(http.StatusNotFound), result["code"])
	assert.Equal(t, "图片不存在", result["message"])
}

func TestUpdateImage_Owner(t *testing.T) {
	e, _, db, jwtSvc, teardown := setupImageTest(t, 10)
	defer teardown()

	user := insertUser(t, db, "update-owner@test.com", "pass123", "UpdateOwner", model.UserRoleUser, model.UserStatusApproved)
	token := generateAccessToken(t, jwtSvc, user.ID, model.UserRoleUser)

	img := insertImage(t, db, user.ID, "Old Title", "Old Description", []string{"old"}, "https://img.example.com/old.jpg")

	body := `{"title":"Updated Title","description":"Updated Description","tags":["new","updated"],"album_id":null}`
	rec := doAuthRequest(e, http.MethodPut, fmt.Sprintf("/api/images/%d", img.ID), body, token)
	assert.Equal(t, http.StatusOK, rec.Code)

	result := parseResponse(t, rec)
	assert.Equal(t, float64(http.StatusOK), result["code"])

	data, ok := result["data"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Updated Title", data["title"])
	assert.Equal(t, "Updated Description", data["description"])

	tags, ok := data["tags"].([]interface{})
	require.True(t, ok)
	assert.Len(t, tags, 2)
}

func TestUpdateImage_NotOwner(t *testing.T) {
	e, _, db, jwtSvc, teardown := setupImageTest(t, 10)
	defer teardown()

	owner := insertUser(t, db, "update-owner2@test.com", "pass123", "Owner2", model.UserRoleUser, model.UserStatusApproved)
	other := insertUser(t, db, "update-other@test.com", "pass123", "Other", model.UserRoleUser, model.UserStatusApproved)

	img := insertImage(t, db, owner.ID, "Owner's Photo", "", []string{"private"}, "https://img.example.com/private.jpg")

	// Other user tries to update
	otherToken := generateAccessToken(t, jwtSvc, other.ID, model.UserRoleUser)
	body := `{"title":"Hacked Title"}`
	rec := doAuthRequest(e, http.MethodPut, fmt.Sprintf("/api/images/%d", img.ID), body, otherToken)
	assert.Equal(t, http.StatusForbidden, rec.Code)

	result := parseResponse(t, rec)
	assert.Equal(t, float64(http.StatusForbidden), result["code"])
	assert.Equal(t, "没有权限编辑此图片", result["message"])
}

func TestUpdateImage_Admin(t *testing.T) {
	e, _, db, jwtSvc, teardown := setupImageTest(t, 10)
	defer teardown()

	owner := insertUser(t, db, "update-admin-owner@test.com", "pass123", "AdminOwner", model.UserRoleUser, model.UserStatusApproved)
	admin := insertUser(t, db, "update-admin@test.com", "pass123", "Admin", model.UserRoleAdmin, model.UserStatusApproved)

	img := insertImage(t, db, owner.ID, "User Photo", "User's image", []string{"user"}, "https://img.example.com/user.jpg")

	// Admin updates the image
	adminToken := generateAccessToken(t, jwtSvc, admin.ID, model.UserRoleAdmin)
	body := `{"title":"Admin Updated"}`
	rec := doAuthRequest(e, http.MethodPut, fmt.Sprintf("/api/images/%d", img.ID), body, adminToken)
	assert.Equal(t, http.StatusOK, rec.Code)

	result := parseResponse(t, rec)
	assert.Equal(t, float64(http.StatusOK), result["code"])

	data, ok := result["data"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Admin Updated", data["title"])
}

func TestDeleteImage_Owner(t *testing.T) {
	e, _, db, jwtSvc, teardown := setupImageTest(t, 10)
	defer teardown()

	user := insertUser(t, db, "delete-owner@test.com", "pass123", "DeleteOwner", model.UserRoleUser, model.UserStatusApproved)
	token := generateAccessToken(t, jwtSvc, user.ID, model.UserRoleUser)

	img := insertImage(t, db, user.ID, "To Delete", "", []string{"temp"}, "https://img.example.com/temp.jpg")

	rec := doAuthRequest(e, http.MethodDelete, fmt.Sprintf("/api/images/%d", img.ID), "", token)
	assert.Equal(t, http.StatusOK, rec.Code)

	result := parseResponse(t, rec)
	assert.Equal(t, float64(http.StatusOK), result["code"])

	data, ok := result["data"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, true, data["deleted"])
}

func TestDeleteImage_NotOwner(t *testing.T) {
	e, _, db, jwtSvc, teardown := setupImageTest(t, 10)
	defer teardown()

	owner := insertUser(t, db, "delete-owner2@test.com", "pass123", "DeleteOwner2", model.UserRoleUser, model.UserStatusApproved)
	other := insertUser(t, db, "delete-other@test.com", "pass123", "DeleteOther", model.UserRoleUser, model.UserStatusApproved)

	img := insertImage(t, db, owner.ID, "Owner's Image", "", []string{"private"}, "https://img.example.com/private2.jpg")

	// Other user tries to delete
	otherToken := generateAccessToken(t, jwtSvc, other.ID, model.UserRoleUser)
	rec := doAuthRequest(e, http.MethodDelete, fmt.Sprintf("/api/images/%d", img.ID), "", otherToken)
	assert.Equal(t, http.StatusForbidden, rec.Code)

	result := parseResponse(t, rec)
	assert.Equal(t, float64(http.StatusForbidden), result["code"])
	assert.Equal(t, "没有权限删除此图片", result["message"])
}
