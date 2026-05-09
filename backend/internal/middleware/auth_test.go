package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"

	"gm_site/internal/service"
)

// setupJWT creates a JWTService with test credentials for testing.
func setupJWT() *service.JWTService {
	return service.NewJWTService(
		"test-access-secret",
		"test-refresh-secret",
		time.Hour,
		24*time.Hour,
	)
}

// generateTestToken generates a valid access token for testing.
func generateTestToken(t *testing.T, svc *service.JWTService, userID int64, role string) string {
	t.Helper()
	token, _, err := svc.GenerateAccessToken(userID, role)
	if err != nil {
		t.Fatalf("failed to generate test token: %v", err)
	}
	return token
}

// newEchoContext creates a new Echo context with the given method and headers.
func newEchoContext(method, target string, headers map[string]string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(method, target, http.NoBody)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

// successHandler returns an Echo handler that checks context values and returns 200.
func successHandler(expectedUserID int64, expectedRole string, check bool) echo.HandlerFunc {
	return func(c echo.Context) error {
		if check {
			if uid, ok := c.Get(UserIDKey).(int64); !ok || uid != expectedUserID {
				return c.JSON(http.StatusInternalServerError, map[string]any{
					"code":    http.StatusInternalServerError,
					"message": "userID mismatch",
				})
			}
			if role, ok := c.Get(UserRoleKey).(string); !ok || role != expectedRole {
				return c.JSON(http.StatusInternalServerError, map[string]any{
					"code":    http.StatusInternalServerError,
					"message": "userRole mismatch",
				})
			}
		}
		return c.NoContent(http.StatusOK)
	}
}

// --- AuthRequired Tests ---

func TestAuthRequired_NoToken(t *testing.T) {
	svc := setupJWT()
	c, rec := newEchoContext(http.MethodGet, "/", nil)

	middleware := AuthRequired(svc)
	handler := middleware(successHandler(0, "", false))

	err := handler(c)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestAuthRequired_InvalidToken(t *testing.T) {
	svc := setupJWT()
	c, rec := newEchoContext(http.MethodGet, "/", map[string]string{
		echo.HeaderAuthorization: "Bearer invalid.token.here",
	})

	middleware := AuthRequired(svc)
	handler := middleware(successHandler(0, "", false))

	err := handler(c)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestAuthRequired_ValidToken(t *testing.T) {
	svc := setupJWT()
	token := generateTestToken(t, svc, 42, "user")

	c, rec := newEchoContext(http.MethodGet, "/", map[string]string{
		echo.HeaderAuthorization: "Bearer " + token,
	})

	middleware := AuthRequired(svc)
	handler := middleware(successHandler(42, "user", true))

	err := handler(c)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestAuthRequired_OPTIONS(t *testing.T) {
	svc := setupJWT()
	c, rec := newEchoContext(http.MethodOptions, "/", nil)

	middleware := AuthRequired(svc)
	handler := middleware(successHandler(0, "", false))

	err := handler(c)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

// --- AdminRequired Tests ---

func TestAdminRequired_NotAdmin(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	// Set a non-admin role
	c.Set(UserRoleKey, "user")

	middleware := AdminRequired()
	handler := middleware(successHandler(0, "", false))

	err := handler(c)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, rec.Code)
	}
}

func TestAdminRequired_NoRole(t *testing.T) {
	// No role set in context (bypassing AuthRequired)
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	middleware := AdminRequired()
	handler := middleware(successHandler(0, "", false))

	err := handler(c)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, rec.Code)
	}
}

func TestAdminRequired_Admin(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(UserRoleKey, "admin")

	middleware := AdminRequired()
	handler := middleware(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	err := handler(c)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

// --- OptionalAuth Tests ---

func TestOptionalAuth_NoToken(t *testing.T) {
	svc := setupJWT()
	c, rec := newEchoContext(http.MethodGet, "/", nil)

	middleware := OptionalAuth(svc)
	handler := middleware(func(c echo.Context) error {
		// Verify no user info in context
		if _, ok := c.Get(UserIDKey).(int64); ok {
			return c.JSON(http.StatusInternalServerError, map[string]any{
				"code":    http.StatusInternalServerError,
				"message": "UserID should not be in context",
			})
		}
		return c.NoContent(http.StatusOK)
	})

	err := handler(c)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestOptionalAuth_ValidToken(t *testing.T) {
	svc := setupJWT()
	token := generateTestToken(t, svc, 99, "editor")

	c, rec := newEchoContext(http.MethodGet, "/", map[string]string{
		echo.HeaderAuthorization: "Bearer " + token,
	})

	middleware := OptionalAuth(svc)
	handler := middleware(successHandler(99, "editor", true))

	err := handler(c)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestOptionalAuth_InvalidToken(t *testing.T) {
	// Invalid token should be silently ignored (guest mode)
	svc := setupJWT()
	c, rec := newEchoContext(http.MethodGet, "/", map[string]string{
		echo.HeaderAuthorization: "Bearer invalid.token.here",
	})

	middleware := OptionalAuth(svc)
	handler := middleware(func(c echo.Context) error {
		if _, ok := c.Get(UserIDKey).(int64); ok {
			return c.JSON(http.StatusInternalServerError, map[string]any{
				"code":    http.StatusInternalServerError,
				"message": "UserID should not be in context for invalid token",
			})
		}
		return c.NoContent(http.StatusOK)
	})

	err := handler(c)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestAuthRequired_NoBearerPrefix(t *testing.T) {
	// Token without "Bearer " prefix should be rejected
	svc := setupJWT()
	token := generateTestToken(t, svc, 1, "user")

	c, rec := newEchoContext(http.MethodGet, "/", map[string]string{
		echo.HeaderAuthorization: token, // no "Bearer " prefix
	})

	middleware := AuthRequired(svc)
	handler := middleware(successHandler(0, "", false))

	err := handler(c)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestAuthRequired_EmptyBearer(t *testing.T) {
	// "Bearer " with no actual token should be rejected
	svc := setupJWT()
	c, rec := newEchoContext(http.MethodGet, "/", map[string]string{
		echo.HeaderAuthorization: "Bearer ",
	})

	middleware := AuthRequired(svc)
	handler := middleware(successHandler(0, "", false))

	err := handler(c)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}
