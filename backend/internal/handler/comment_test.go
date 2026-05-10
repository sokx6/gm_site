package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
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

// setupCommentTestDB creates a temporary SQLite database with all required tables.
func setupCommentTestDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()

	db, teardown := setupTestDB(t)

	_, err := db.Exec(`
		CREATE TABLE images (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			album_id INTEGER,
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
		CREATE TABLE comments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			image_id INTEGER NOT NULL REFERENCES images(id) ON DELETE CASCADE,
			user_id INTEGER NOT NULL REFERENCES users(id),
			content TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX idx_comments_image_id ON comments(image_id);
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
	require.NoError(t, err, "failed to create additional tables")

	return db, teardown
}

// insertTestImage inserts a test image and returns its ID.
func insertTestImage(t *testing.T, db *sql.DB, uploadedBy int64) *model.Image {
	t.Helper()

	now := time.Now()
	result, err := db.Exec(
		`INSERT INTO images (title, lsky_url, uploaded_by, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
		"test image", "https://example.com/img.jpg", uploadedBy, now, now,
	)
	require.NoError(t, err, "failed to insert test image")

	id, err := result.LastInsertId()
	require.NoError(t, err, "failed to get last insert id")

	return &model.Image{
		ID:         id,
		Title:      "test image",
		LskyURL:    "https://example.com/img.jpg",
		UploadedBy: uploadedBy,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// insertTestComment inserts a test comment and returns it.
func insertTestComment(t *testing.T, db *sql.DB, imageID, userID int64, content string) *model.Comment {
	t.Helper()

	now := time.Now()
	result, err := db.Exec(
		`INSERT INTO comments (image_id, user_id, content, created_at) VALUES (?, ?, ?, ?)`,
		imageID, userID, content, now,
	)
	require.NoError(t, err, "failed to insert test comment")

	id, err := result.LastInsertId()
	require.NoError(t, err, "failed to get last insert id")

	return &model.Comment{
		ID:        id,
		ImageID:   imageID,
		UserID:    userID,
		Content:   content,
		CreatedAt: now,
	}
}

// setupCommentTest creates all dependencies and registers comment routes for testing.
func setupCommentTest(t *testing.T) (*echo.Echo, *CommentHandler, *sql.DB, *service.JWTService, func()) {
	t.Helper()

	db, teardown := setupCommentTestDB(t)
	commentRepo := repository.NewCommentRepository(db)
	imageRepo := repository.NewImageRepository(db)
	userRepo := repository.NewUserRepository(db)
	notificationRepo := repository.NewNotificationRepository(db)
	emailSvc := service.NewMockEmailService()
	jwtSvc := newTestJWTService()
	handler := NewCommentHandler(commentRepo, imageRepo, userRepo, notificationRepo, emailSvc, db)

	e := echo.New()
	// Public route
	e.GET("/api/images/:id/comments", handler.ListByImage)
	// Authenticated routes
	g := e.Group("/api", middleware.AuthRequired(jwtSvc))
	g.POST("/images/:id/comments", handler.Create)
	g.DELETE("/comments/:id", handler.Delete)

	return e, handler, db, jwtSvc, teardown
}

// parseListResponse decodes the JSON response into a map and returns comments array and data map.
func parseListResponse(t *testing.T, rec *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	var result map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&result)
	require.NoError(t, err)
	return result
}

// assertSuccessCode asserts that the response has code 200 or 201.
func assertSuccessCode(t *testing.T, rec *httptest.ResponseRecorder, expectedCode int) {
	t.Helper()
	result := parseListResponse(t, rec)
	assert.Equal(t, float64(expectedCode), result["code"])
}

// ---------------------------------------------------------------------------
// Create tests
// ---------------------------------------------------------------------------

func TestCommentCreate_Success(t *testing.T) {
	e, _, db, jwtSvc, teardown := setupCommentTest(t)
	defer teardown()

	user := insertUser(t, db, "commenter@test.com", "pass", "Commenter", model.UserRoleUser, model.UserStatusApproved)
	image := insertTestImage(t, db, user.ID)
	token := generateAccessToken(t, jwtSvc, user.ID, user.Role)

	body := `{"content":"这是一条评论"}`
	rec := doAuthRequest(e, http.MethodPost, fmt.Sprintf("/api/images/%d/comments", image.ID), body, token)
	assert.Equal(t, http.StatusCreated, rec.Code)

	result := parseListResponse(t, rec)
	assert.Equal(t, float64(http.StatusCreated), result["code"])
	data, ok := result["data"].(map[string]interface{})
	require.True(t, ok)
	assert.NotZero(t, data["id"])
	assert.Equal(t, "这是一条评论", data["content"])
}

func TestCommentCreate_XSS(t *testing.T) {
	e, _, db, jwtSvc, teardown := setupCommentTest(t)
	defer teardown()

	user := insertUser(t, db, "xss@test.com", "pass", "XSS", model.UserRoleUser, model.UserStatusApproved)
	image := insertTestImage(t, db, user.ID)
	token := generateAccessToken(t, jwtSvc, user.ID, user.Role)

	body := `{"content":"<script>alert('xss')</script>"}`
	rec := doAuthRequest(e, http.MethodPost, fmt.Sprintf("/api/images/%d/comments", image.ID), body, token)
	assert.Equal(t, http.StatusCreated, rec.Code)

	result := parseListResponse(t, rec)
	data, ok := result["data"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "&lt;script&gt;alert('xss')&lt;/script&gt;", data["content"])
}

func TestCommentCreate_Unauthenticated(t *testing.T) {
	e, _, db, _, teardown := setupCommentTest(t)
	defer teardown()

	user := insertUser(t, db, "anon@test.com", "pass", "Anon", model.UserRoleUser, model.UserStatusApproved)
	image := insertTestImage(t, db, user.ID)

	body := `{"content":"test"}`
	rec := doRequest(e, http.MethodPost, fmt.Sprintf("/api/images/%d/comments", image.ID), body)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestCommentCreate_EmptyContent(t *testing.T) {
	e, _, db, jwtSvc, teardown := setupCommentTest(t)
	defer teardown()

	user := insertUser(t, db, "empty@test.com", "pass", "Empty", model.UserRoleUser, model.UserStatusApproved)
	image := insertTestImage(t, db, user.ID)
	token := generateAccessToken(t, jwtSvc, user.ID, user.Role)

	body := `{"content":""}`
	rec := doAuthRequest(e, http.MethodPost, fmt.Sprintf("/api/images/%d/comments", image.ID), body, token)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCommentCreate_InvalidImageID(t *testing.T) {
	e, _, db, jwtSvc, teardown := setupCommentTest(t)
	defer teardown()

	user := insertUser(t, db, "badid@test.com", "pass", "BadID", model.UserRoleUser, model.UserStatusApproved)
	token := generateAccessToken(t, jwtSvc, user.ID, user.Role)

	body := `{"content":"test"}`
	rec := doAuthRequest(e, http.MethodPost, "/api/images/abc/comments", body, token)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// ---------------------------------------------------------------------------
// List tests
// ---------------------------------------------------------------------------

func TestCommentList_Success(t *testing.T) {
	e, _, db, _, teardown := setupCommentTest(t)
	defer teardown()

	user := insertUser(t, db, "listuser@test.com", "pass", "ListUser", model.UserRoleUser, model.UserStatusApproved)
	image := insertTestImage(t, db, user.ID)

	// Insert 3 comments
	for i := 0; i < 3; i++ {
		insertTestComment(t, db, image.ID, user.ID, fmt.Sprintf("comment %d", i+1))
	}

	rec := doRequest(e, http.MethodGet, fmt.Sprintf("/api/images/%d/comments?page=1&limit=2", image.ID), "")
	assert.Equal(t, http.StatusOK, rec.Code)

	result := parseListResponse(t, rec)
	assert.Equal(t, float64(http.StatusOK), result["code"])

	data, ok := result["data"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, float64(3), data["total"])
	assert.Equal(t, float64(1), data["page"])
	assert.Equal(t, float64(2), data["limit"])

	comments, ok := data["comments"].([]interface{})
	require.True(t, ok)
	assert.Len(t, comments, 2)
}

func TestCommentList_DefaultPagination(t *testing.T) {
	e, _, db, _, teardown := setupCommentTest(t)
	defer teardown()

	user := insertUser(t, db, "defaultpage@test.com", "pass", "Default", model.UserRoleUser, model.UserStatusApproved)
	image := insertTestImage(t, db, user.ID)

	// Insert 25 comments
	for i := 0; i < 25; i++ {
		insertTestComment(t, db, image.ID, user.ID, fmt.Sprintf("comment %d", i+1))
	}

	// No pagination params → defaults to page=1, limit=20
	rec := doRequest(e, http.MethodGet, fmt.Sprintf("/api/images/%d/comments", image.ID), "")
	assert.Equal(t, http.StatusOK, rec.Code)

	result := parseListResponse(t, rec)
	data, ok := result["data"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, float64(25), data["total"])
	assert.Equal(t, float64(1), data["page"])
	assert.Equal(t, float64(20), data["limit"])

	comments, ok := data["comments"].([]interface{})
	require.True(t, ok)
	assert.Len(t, comments, 20)
}

func TestCommentList_NoComments(t *testing.T) {
	e, _, db, _, teardown := setupCommentTest(t)
	defer teardown()

	user := insertUser(t, db, "nocomments@test.com", "pass", "NoComments", model.UserRoleUser, model.UserStatusApproved)
	image := insertTestImage(t, db, user.ID)

	rec := doRequest(e, http.MethodGet, fmt.Sprintf("/api/images/%d/comments", image.ID), "")
	assert.Equal(t, http.StatusOK, rec.Code)

	result := parseListResponse(t, rec)
	data, ok := result["data"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, float64(0), data["total"])
	assert.Equal(t, float64(1), data["page"])
	assert.Equal(t, float64(20), data["limit"])

	comments, ok := data["comments"].([]interface{})
	require.True(t, ok)
	assert.Empty(t, comments)
}

func TestCommentList_InvalidImageID(t *testing.T) {
	e, _, _, _, teardown := setupCommentTest(t)
	defer teardown()

	rec := doRequest(e, http.MethodGet, "/api/images/abc/comments", "")
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// ---------------------------------------------------------------------------
// Delete tests
// ---------------------------------------------------------------------------

func TestCommentDelete_AsAuthor(t *testing.T) {
	e, _, db, jwtSvc, teardown := setupCommentTest(t)
	defer teardown()

	author := insertUser(t, db, "author@test.com", "pass", "Author", model.UserRoleUser, model.UserStatusApproved)
	image := insertTestImage(t, db, author.ID)
	comment := insertTestComment(t, db, image.ID, author.ID, "my comment")
	token := generateAccessToken(t, jwtSvc, author.ID, author.Role)

	rec := doAuthRequest(e, http.MethodDelete, fmt.Sprintf("/api/comments/%d", comment.ID), "", token)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestCommentDelete_AsImageUploader(t *testing.T) {
	e, _, db, jwtSvc, teardown := setupCommentTest(t)
	defer teardown()

	uploader := insertUser(t, db, "uploader@test.com", "pass", "Uploader", model.UserRoleUser, model.UserStatusApproved)
	commenter := insertUser(t, db, "commenter2@test.com", "pass", "Commenter", model.UserRoleUser, model.UserStatusApproved)
	image := insertTestImage(t, db, uploader.ID)
	comment := insertTestComment(t, db, image.ID, commenter.ID, "someone else's comment")
	token := generateAccessToken(t, jwtSvc, uploader.ID, uploader.Role)

	rec := doAuthRequest(e, http.MethodDelete, fmt.Sprintf("/api/comments/%d", comment.ID), "", token)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestCommentDelete_AsAdmin(t *testing.T) {
	e, _, db, jwtSvc, teardown := setupCommentTest(t)
	defer teardown()

	admin := insertUser(t, db, "admin@test.com", "pass", "Admin", model.UserRoleAdmin, model.UserStatusApproved)
	user := insertUser(t, db, "regular@test.com", "pass", "Regular", model.UserRoleUser, model.UserStatusApproved)
	image := insertTestImage(t, db, user.ID)
	comment := insertTestComment(t, db, image.ID, user.ID, "regular comment")
	token := generateAccessToken(t, jwtSvc, admin.ID, admin.Role)

	rec := doAuthRequest(e, http.MethodDelete, fmt.Sprintf("/api/comments/%d", comment.ID), "", token)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestCommentDelete_Forbidden(t *testing.T) {
	e, _, db, jwtSvc, teardown := setupCommentTest(t)
	defer teardown()

	author := insertUser(t, db, "author3@test.com", "pass", "Author", model.UserRoleUser, model.UserStatusApproved)
	other := insertUser(t, db, "other@test.com", "pass", "Other", model.UserRoleUser, model.UserStatusApproved)
	image := insertTestImage(t, db, author.ID)
	comment := insertTestComment(t, db, image.ID, author.ID, "my comment")
	token := generateAccessToken(t, jwtSvc, other.ID, other.Role)

	rec := doAuthRequest(e, http.MethodDelete, fmt.Sprintf("/api/comments/%d", comment.ID), "", token)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestCommentDelete_NotFound(t *testing.T) {
	e, _, db, jwtSvc, teardown := setupCommentTest(t)
	defer teardown()

	user := insertUser(t, db, "notfound@test.com", "pass", "NotFound", model.UserRoleUser, model.UserStatusApproved)
	token := generateAccessToken(t, jwtSvc, user.ID, user.Role)

	rec := doAuthRequest(e, http.MethodDelete, "/api/comments/99999", "", token)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestCommentDelete_Unauthenticated(t *testing.T) {
	e, _, db, jwtSvc, teardown := setupCommentTest(t)
	defer teardown()

	user := insertUser(t, db, "unauth@test.com", "pass", "Unauth", model.UserRoleUser, model.UserStatusApproved)
	image := insertTestImage(t, db, user.ID)
	comment := insertTestComment(t, db, image.ID, user.ID, "test")
	_ = generateAccessToken(t, jwtSvc, user.ID, user.Role)

	// Delete without token
	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/comments/%d", comment.ID), nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
