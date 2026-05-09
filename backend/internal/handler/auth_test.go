package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"

	"gm_site/internal/model"
	"gm_site/internal/repository"
	"gm_site/internal/service"
)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

// setupTestDB creates a temporary SQLite database with the users table
// and returns the db handle and a teardown function.
func setupTestDB(t *testing.T) (*sql.DB, func()) {
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
	`)
	require.NoError(t, err, "failed to create users table")

	teardown := func() {
		db.Close()
	}

	return db, teardown
}

// insertUser inserts a user with the given fields and a bcrypt-hashed password.
func insertUser(t *testing.T, db *sql.DB, email, password, nickname, role, status string) *model.User {
	t.Helper()

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	require.NoError(t, err, "failed to hash password")

	now := time.Now()
	result, err := db.Exec(
		`INSERT INTO users (email, password_hash, nickname, role, status, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		email, string(hash), nickname, role, status, now, now,
	)
	require.NoError(t, err, "failed to insert test user")

	id, err := result.LastInsertId()
	require.NoError(t, err, "failed to get last insert id")

	return &model.User{
		ID:           id,
		Email:        email,
		PasswordHash: string(hash),
		Nickname:     nickname,
		Role:         role,
		Status:       status,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// newTestJWTService creates a JWTService suitable for tests.
func newTestJWTService() *service.JWTService {
	return service.NewJWTService(
		"test-access-secret",
		"test-refresh-secret",
		15*time.Minute,
		168*time.Hour,
	)
}

// setupAuthTest creates all dependencies and returns the Echo instance,
// AuthHandler, sql.DB, and teardown function.
func setupAuthTest(t *testing.T) (*echo.Echo, *AuthHandler, *sql.DB, func()) {
	t.Helper()

	db, teardownDB := setupTestDB(t)
	userRepo := repository.NewUserRepository(db)
	jwtSvc := newTestJWTService()
	emailSvc := service.NewMockEmailService()
	handler := NewAuthHandler(userRepo, jwtSvc, emailSvc, "")

	e := echo.New()
	e.POST("/api/auth/login", handler.Login)
	e.POST("/api/auth/refresh", handler.Refresh)

	return e, handler, db, teardownDB
}

// setupAuthRegisterTest creates all dependencies for Register tests and returns
// the Echo instance, AuthHandler, sql.DB, MockEmailService, and teardown function.
func setupAuthRegisterTest(t *testing.T, adminEmail string) (*echo.Echo, *AuthHandler, *sql.DB, *service.MockEmailService, func()) {
	t.Helper()

	db, teardownDB := setupTestDB(t)
	userRepo := repository.NewUserRepository(db)
	jwtSvc := newTestJWTService()
	emailSvc := service.NewMockEmailService()
	handler := NewAuthHandler(userRepo, jwtSvc, emailSvc, adminEmail)

	e := echo.New()
	e.POST("/api/auth/register", handler.Register)

	return e, handler, db, emailSvc, teardownDB
}

// doRequest performs an HTTP request against the Echo instance and returns the response recorder.
func doRequest(e *echo.Echo, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

// parseResponse decodes the JSON response body into a map.
func parseResponse(t *testing.T, rec *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	var result map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&result)
	require.NoError(t, err, "failed to decode response body")
	return result
}

// generateExpiredRefreshToken creates a refresh token that is already expired.
func generateExpiredRefreshToken(subject string) string {
	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		Subject:   subject,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte("test-refresh-secret"))
	return signed
}

// ---------------------------------------------------------------------------
// Login tests
// ---------------------------------------------------------------------------

func TestLogin_Success(t *testing.T) {
	e, _, db, teardown := setupAuthTest(t)
	defer teardown()

	// Insert an approved user
	insertUser(t, db, "alice@test.com", "password123", "Alice", "user", model.UserStatusApproved)

	body := `{"email":"alice@test.com","password":"password123"}`
	rec := doRequest(e, http.MethodPost, "/api/auth/login", body)
	assert.Equal(t, http.StatusOK, rec.Code)

	result := parseResponse(t, rec)
	assert.Equal(t, float64(http.StatusOK), result["code"])
	assert.Equal(t, "success", result["message"])

	data, ok := result["data"].(map[string]interface{})
	require.True(t, ok, "data should be a map")
	assert.NotEmpty(t, data["access_token"])
	assert.NotEmpty(t, data["refresh_token"])
	assert.NotZero(t, data["expires_in"])
}

func TestLogin_WrongPassword(t *testing.T) {
	e, _, db, teardown := setupAuthTest(t)
	defer teardown()

	insertUser(t, db, "bob@test.com", "correct-password", "Bob", "user", model.UserStatusApproved)

	body := `{"email":"bob@test.com","password":"wrong-password"}`
	rec := doRequest(e, http.MethodPost, "/api/auth/login", body)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	result := parseResponse(t, rec)
	assert.Equal(t, float64(http.StatusUnauthorized), result["code"])
	assert.Equal(t, "邮箱或密码错误", result["message"])
}

func TestLogin_UserNotFound(t *testing.T) {
	e, _, _, teardown := setupAuthTest(t)
	defer teardown()

	body := `{"email":"nonexistent@test.com","password":"password123"}`
	rec := doRequest(e, http.MethodPost, "/api/auth/login", body)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	result := parseResponse(t, rec)
	assert.Equal(t, float64(http.StatusUnauthorized), result["code"])
	assert.Equal(t, "邮箱或密码错误", result["message"])
}

func TestLogin_PendingUser(t *testing.T) {
	e, _, db, teardown := setupAuthTest(t)
	defer teardown()

	insertUser(t, db, "pending@test.com", "password123", "PendingUser", "user", model.UserStatusPending)

	body := `{"email":"pending@test.com","password":"password123"}`
	rec := doRequest(e, http.MethodPost, "/api/auth/login", body)
	assert.Equal(t, http.StatusForbidden, rec.Code)

	result := parseResponse(t, rec)
	assert.Equal(t, float64(http.StatusForbidden), result["code"])
	assert.Equal(t, "账户正在审核中", result["message"])
}

func TestLogin_RejectedUser(t *testing.T) {
	e, _, db, teardown := setupAuthTest(t)
	defer teardown()

	insertUser(t, db, "rejected@test.com", "password123", "RejectedUser", "user", model.UserStatusRejected)

	body := `{"email":"rejected@test.com","password":"password123"}`
	rec := doRequest(e, http.MethodPost, "/api/auth/login", body)
	assert.Equal(t, http.StatusForbidden, rec.Code)

	result := parseResponse(t, rec)
	assert.Equal(t, float64(http.StatusForbidden), result["code"])
	assert.Equal(t, "账户已被拒绝", result["message"])
}

// ---------------------------------------------------------------------------
// Refresh tests
// ---------------------------------------------------------------------------

func TestRefresh_Success(t *testing.T) {
	e, handler, db, teardown := setupAuthTest(t)
	defer teardown()

	user := insertUser(t, db, "refresh@test.com", "password123", "RefreshUser", "user", model.UserStatusApproved)

	// Generate a valid token pair to get a refresh token
	pair, err := handler.jwtSvc.GenerateTokenPair(user.ID, user.Role)
	require.NoError(t, err)

	body := `{"refresh_token":"` + pair.RefreshToken + `"}`
	rec := doRequest(e, http.MethodPost, "/api/auth/refresh", body)
	assert.Equal(t, http.StatusOK, rec.Code)

	result := parseResponse(t, rec)
	assert.Equal(t, float64(http.StatusOK), result["code"])
	assert.Equal(t, "success", result["message"])

	data, ok := result["data"].(map[string]interface{})
	require.True(t, ok, "data should be a map")
	assert.NotEmpty(t, data["access_token"])
	assert.NotEmpty(t, data["refresh_token"])
	assert.NotZero(t, data["expires_in"])
}

func TestRefresh_InvalidToken(t *testing.T) {
	e, _, _, teardown := setupAuthTest(t)
	defer teardown()

	body := `{"refresh_token":"this-is-not-a-valid-token"}`
	rec := doRequest(e, http.MethodPost, "/api/auth/refresh", body)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	result := parseResponse(t, rec)
	assert.Equal(t, float64(http.StatusUnauthorized), result["code"])
	assert.Equal(t, "无效的刷新令牌", result["message"])
}

func TestRefresh_ExpiredToken(t *testing.T) {
	e, _, _, teardown := setupAuthTest(t)
	defer teardown()

	// Generate an expired refresh token with subject "1" (user ID)
	expiredToken := generateExpiredRefreshToken("1")

	body := `{"refresh_token":"` + expiredToken + `"}`
	rec := doRequest(e, http.MethodPost, "/api/auth/refresh", body)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	result := parseResponse(t, rec)
	assert.Equal(t, float64(http.StatusUnauthorized), result["code"])
	assert.Equal(t, "刷新令牌已过期", result["message"])
}

// ---------------------------------------------------------------------------
// Register tests
// ---------------------------------------------------------------------------

func TestRegister_Success(t *testing.T) {
	e, _, db, emailSvc, teardown := setupAuthRegisterTest(t, "admin@test.com")
	defer teardown()

	body := `{"email":"newuser@test.com","password":"pass123","nickname":"NewUser"}`
	rec := doRequest(e, http.MethodPost, "/api/auth/register", body)
	assert.Equal(t, http.StatusCreated, rec.Code)

	result := parseResponse(t, rec)
	data, ok := result["data"].(map[string]interface{})
	require.True(t, ok)
	msg, ok := data["message"].(string)
	require.True(t, ok, "data.message should be a string: %v", data["message"])
	assert.Equal(t, "注册成功，请等待管理员审核", msg)

	// Verify user was created in DB with pending status
	user, err := db.Query("SELECT email, role, status FROM users WHERE email = ?", "newuser@test.com")
	require.NoError(t, err)
	defer user.Close()
	require.True(t, user.Next(), "user should exist in database")
	var email, role, status string
	require.NoError(t, user.Scan(&email, &role, &status))
	assert.Equal(t, "newuser@test.com", email)
	assert.Equal(t, "user", role)
	assert.Equal(t, "pending", status)

	// Verify email was queued for admin
	require.Len(t, emailSvc.Messages, 1)
	assert.Equal(t, "admin_notification", emailSvc.Messages[0].Type)
	assert.Equal(t, "newuser@test.com", emailSvc.Messages[0].Email)
}

func TestRegister_DuplicateEmail(t *testing.T) {
	e, _, db, _, teardown := setupAuthRegisterTest(t, "")
	defer teardown()

	// Pre-insert a user with the same email
	insertUser(t, db, "dup@test.com", "password123", "Dup", "user", model.UserStatusApproved)

	body := `{"email":"dup@test.com","password":"newpass123","nickname":"Duplicate"}`
	rec := doRequest(e, http.MethodPost, "/api/auth/register", body)
	assert.Equal(t, http.StatusConflict, rec.Code)

	result := parseResponse(t, rec)
	assert.Equal(t, float64(http.StatusConflict), result["code"])
	assert.Equal(t, "该邮箱已被注册", result["message"])
}

func TestRegister_InvalidEmail(t *testing.T) {
	e, _, _, _, teardown := setupAuthRegisterTest(t, "")
	defer teardown()

	body := `{"email":"not-an-email","password":"pass123","nickname":"BadEmail"}`
	rec := doRequest(e, http.MethodPost, "/api/auth/register", body)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	result := parseResponse(t, rec)
	assert.Equal(t, float64(http.StatusBadRequest), result["code"])
	assert.Equal(t, "请求参数校验失败", result["message"])
	data, ok := result["data"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "email", data["Email"])
}

func TestRegister_WeakPassword(t *testing.T) {
	e, _, _, _, teardown := setupAuthRegisterTest(t, "")
	defer teardown()

	body := `{"email":"weak@test.com","password":"12345","nickname":"WeakPass"}`
	rec := doRequest(e, http.MethodPost, "/api/auth/register", body)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	result := parseResponse(t, rec)
	assert.Equal(t, float64(http.StatusBadRequest), result["code"])
	assert.Equal(t, "请求参数校验失败", result["message"])
	data, ok := result["data"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "min", data["Password"])
}

func TestRegister_AdminEmail(t *testing.T) {
	adminEmail := "admin@example.com"
	e, _, db, _, teardown := setupAuthRegisterTest(t, adminEmail)
	defer teardown()

	body := `{"email":"admin@example.com","password":"pass123","nickname":"Admin"}`
	rec := doRequest(e, http.MethodPost, "/api/auth/register", body)
	assert.Equal(t, http.StatusCreated, rec.Code)

	// Verify user has admin role
	user, err := db.Query("SELECT email, role FROM users WHERE email = ?", adminEmail)
	require.NoError(t, err)
	defer user.Close()
	require.True(t, user.Next(), "admin user should exist")
	var email, role string
	require.NoError(t, user.Scan(&email, &role))
	assert.Equal(t, "admin", role)
}

func TestRegister_EmailSentToAdmin(t *testing.T) {
	e, _, _, emailSvc, teardown := setupAuthRegisterTest(t, "admin@test.com")
	defer teardown()

	body := `{"email":"newuser@test.com","password":"pass123","nickname":"NewUser"}`
	rec := doRequest(e, http.MethodPost, "/api/auth/register", body)
	assert.Equal(t, http.StatusCreated, rec.Code)

	// Verify email service was called (may need brief wait for goroutine)
	require.Eventually(t, func() bool {
		return len(emailSvc.Messages) > 0
	}, time.Second, 10*time.Millisecond, "email should have been sent")

	require.Len(t, emailSvc.Messages, 1)
	assert.Equal(t, "admin_notification", emailSvc.Messages[0].Type)
	assert.Equal(t, "newuser@test.com", emailSvc.Messages[0].Email)
	assert.Equal(t, "NewUser", emailSvc.Messages[0].Nickname)
}
