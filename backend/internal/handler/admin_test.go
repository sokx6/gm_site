package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"

	"gm_site/internal/logger"
	"gm_site/internal/middleware"
	"gm_site/internal/model"
	"gm_site/internal/repository"
	"gm_site/internal/service"
)

// TestMain initializes the global logger for all handler tests.
func TestMain(m *testing.M) {
	logger.L = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	os.Exit(m.Run())
}

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

// mockEmailService simulates the email service for tests.
type mockEmailService struct{}

func (m *mockEmailService) SendAdminNotification(_, _ string) error { return nil }
func (m *mockEmailService) SendApprovalNotification(_ string, _ bool) error { return nil }
func (m *mockEmailService) SendRegisterApprovedNotification(_ string) error { return nil }
func (m *mockEmailService) SendRegisterRejectedNotification(_ string) error { return nil }
func (m *mockEmailService) SendCommentReplyNotification(_, _, _ string) error { return nil }
func (m *mockEmailService) SendFriendRequestNotification(_, _ string) error { return nil }
func (m *mockEmailService) SendFriendAcceptedNotification(_, _ string) error { return nil }
func (m *mockEmailService) SendFriendRejectedNotification(_, _ string) error { return nil }
func (m *mockEmailService) SendImageCommentNotification(_, _, _, _ string) error { return nil }

// generateAccessToken creates a valid JWT access token for testing.
func generateAccessToken(t *testing.T, jwtSvc *service.JWTService, userID int64, role string) string {
	t.Helper()
	token, _, err := jwtSvc.GenerateAccessToken(userID, role)
	require.NoError(t, err)
	return token
}

// doAuthRequest performs an HTTP request with a Bearer token and returns the response recorder.
func doAuthRequest(e *echo.Echo, method, path, body, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

// setupAdminTest creates all dependencies and registers admin routes with middleware.
func setupAdminTest(t *testing.T) (*echo.Echo, *AdminHandler, *sql.DB, *service.JWTService, func()) {
	t.Helper()

	db, teardownDB := setupTestDB(t)
	userRepo := repository.NewUserRepository(db)
	jwtSvc := newTestJWTService()
	emailSvc := &mockEmailService{}
	handler := NewAdminHandler(userRepo, emailSvc)

	e := echo.New()
	g := e.Group("/api/admin", middleware.AuthRequired(jwtSvc), middleware.AdminRequired())
	g.GET("/users/pending", handler.ListPending)
	g.PUT("/users/:id/approve", handler.ApproveUser)
	g.PUT("/users/:id/reject", handler.RejectUser)

	return e, handler, db, jwtSvc, teardownDB
}

// ---------------------------------------------------------------------------
// ListPending tests
// ---------------------------------------------------------------------------

func TestListPending_Admin(t *testing.T) {
	e, _, db, jwtSvc, teardown := setupAdminTest(t)
	defer teardown()

	admin := insertUser(t, db, "admin@test.com", "admin123", "Admin", model.UserRoleAdmin, model.UserStatusApproved)

	// Create pending users
	insertUser(t, db, "pending1@test.com", "pass123", "User1", model.UserRoleUser, model.UserStatusPending)
	insertUser(t, db, "pending2@test.com", "pass123", "User2", model.UserRoleUser, model.UserStatusPending)

	// Create a non-pending user that should NOT appear in results
	insertUser(t, db, "approved@test.com", "pass123", "Approved", model.UserRoleUser, model.UserStatusApproved)

	token := generateAccessToken(t, jwtSvc, admin.ID, model.UserRoleAdmin)
	rec := doAuthRequest(e, http.MethodGet, "/api/admin/users/pending", "", token)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, float64(http.StatusOK), resp["code"])

	data, ok := resp["data"].([]interface{})
	require.True(t, ok, "data should be an array")
	assert.Len(t, data, 2)

	// Verify password_hash is not exposed in the response
	for _, item := range data {
		u, ok := item.(map[string]interface{})
		require.True(t, ok, "each item should be a map")
		_, hasPassword := u["password_hash"]
		assert.False(t, hasPassword, "password_hash must not appear in response")
	}
}

func TestListPending_NonAdmin(t *testing.T) {
	e, _, db, jwtSvc, teardown := setupAdminTest(t)
	defer teardown()

	user := insertUser(t, db, "user@test.com", "pass123", "Regular", model.UserRoleUser, model.UserStatusApproved)

	token := generateAccessToken(t, jwtSvc, user.ID, model.UserRoleUser)
	rec := doAuthRequest(e, http.MethodGet, "/api/admin/users/pending", "", token)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

// ---------------------------------------------------------------------------
// Approve tests
// ---------------------------------------------------------------------------

func TestApproveUser_Success(t *testing.T) {
	e, _, db, jwtSvc, teardown := setupAdminTest(t)
	defer teardown()

	admin := insertUser(t, db, "admin@test.com", "admin123", "Admin", model.UserRoleAdmin, model.UserStatusApproved)
	pending := insertUser(t, db, "pending@test.com", "pass123", "Pending", model.UserRoleUser, model.UserStatusPending)

	token := generateAccessToken(t, jwtSvc, admin.ID, model.UserRoleAdmin)
	rec := doAuthRequest(e, http.MethodPut, fmt.Sprintf("/api/admin/users/%d/approve", pending.ID), "", token)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, float64(http.StatusOK), resp["code"])

	data, ok := resp["data"].(map[string]interface{})
	require.True(t, ok, "data should be a map")
	assert.Equal(t, float64(pending.ID), data["id"])
	assert.Equal(t, model.UserStatusApproved, data["status"])

	// Verify database was updated
	var status string
	err = db.QueryRow("SELECT status FROM users WHERE id = ?", pending.ID).Scan(&status)
	require.NoError(t, err)
	assert.Equal(t, model.UserStatusApproved, status)
}

func TestApproveUser_NotFound(t *testing.T) {
	e, _, db, jwtSvc, teardown := setupAdminTest(t)
	defer teardown()

	admin := insertUser(t, db, "admin@test.com", "admin123", "Admin", model.UserRoleAdmin, model.UserStatusApproved)

	token := generateAccessToken(t, jwtSvc, admin.ID, model.UserRoleAdmin)
	rec := doAuthRequest(e, http.MethodPut, "/api/admin/users/99999/approve", "", token)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestApproveUser_NonAdmin(t *testing.T) {
	e, _, db, jwtSvc, teardown := setupAdminTest(t)
	defer teardown()

	regular := insertUser(t, db, "user@test.com", "pass123", "Regular", model.UserRoleUser, model.UserStatusApproved)
	pending := insertUser(t, db, "pending@test.com", "pass123", "Pending", model.UserRoleUser, model.UserStatusPending)

	token := generateAccessToken(t, jwtSvc, regular.ID, model.UserRoleUser)
	rec := doAuthRequest(e, http.MethodPut, fmt.Sprintf("/api/admin/users/%d/approve", pending.ID), "", token)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestApproveUser_AlreadyApproved(t *testing.T) {
	e, _, db, jwtSvc, teardown := setupAdminTest(t)
	defer teardown()

	admin := insertUser(t, db, "admin@test.com", "admin123", "Admin", model.UserRoleAdmin, model.UserStatusApproved)
	approved := insertUser(t, db, "approved@test.com", "pass123", "Approved", model.UserRoleUser, model.UserStatusApproved)

	token := generateAccessToken(t, jwtSvc, admin.ID, model.UserRoleAdmin)
	rec := doAuthRequest(e, http.MethodPut, fmt.Sprintf("/api/admin/users/%d/approve", approved.ID), "", token)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// ---------------------------------------------------------------------------
// Reject tests
// ---------------------------------------------------------------------------

func TestRejectUser_Success(t *testing.T) {
	e, _, db, jwtSvc, teardown := setupAdminTest(t)
	defer teardown()

	admin := insertUser(t, db, "admin@test.com", "admin123", "Admin", model.UserRoleAdmin, model.UserStatusApproved)
	pending := insertUser(t, db, "pending@test.com", "pass123", "Pending", model.UserRoleUser, model.UserStatusPending)

	token := generateAccessToken(t, jwtSvc, admin.ID, model.UserRoleAdmin)
	rec := doAuthRequest(e, http.MethodPut, fmt.Sprintf("/api/admin/users/%d/reject", pending.ID), "", token)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, float64(http.StatusOK), resp["code"])

	data, ok := resp["data"].(map[string]interface{})
	require.True(t, ok, "data should be a map")
	assert.Equal(t, float64(pending.ID), data["id"])
	assert.Equal(t, model.UserStatusRejected, data["status"])

	// Verify database was updated
	var status string
	err = db.QueryRow("SELECT status FROM users WHERE id = ?", pending.ID).Scan(&status)
	require.NoError(t, err)
	assert.Equal(t, model.UserStatusRejected, status)
}

func TestRejectUser_AlreadyRejected(t *testing.T) {
	e, _, db, jwtSvc, teardown := setupAdminTest(t)
	defer teardown()

	admin := insertUser(t, db, "admin@test.com", "admin123", "Admin", model.UserRoleAdmin, model.UserStatusApproved)
	rejected := insertUser(t, db, "rejected@test.com", "pass123", "Rejected", model.UserRoleUser, model.UserStatusRejected)

	token := generateAccessToken(t, jwtSvc, admin.ID, model.UserRoleAdmin)
	rec := doAuthRequest(e, http.MethodPut, fmt.Sprintf("/api/admin/users/%d/reject", rejected.ID), "", token)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
