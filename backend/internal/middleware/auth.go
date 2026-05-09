package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"gm_site/internal/service"
)

const (
	// UserIDKey is the context key for storing the authenticated user's ID.
	UserIDKey = "userID"
	// UserRoleKey is the context key for storing the authenticated user's role.
	UserRoleKey = "userRole"
)

// AuthRequired returns an Echo middleware that requires a valid JWT access token.
// It extracts the token from the "Authorization: Bearer <token>" header,
// validates it using the provided JWTService, and injects the user ID and role
// into the request context.
//
// OPTIONS requests are always allowed through without authentication.
func AuthRequired(jwtService *service.JWTService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Skip authentication for OPTIONS requests (CORS preflight)
			if c.Request().Method == http.MethodOptions {
				return next(c)
			}

			authHeader := c.Request().Header.Get(echo.HeaderAuthorization)
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, map[string]any{
					"code":    http.StatusUnauthorized,
					"message": "未登录或token已过期",
				})
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenStr == authHeader {
				// No "Bearer " prefix found
				return c.JSON(http.StatusUnauthorized, map[string]any{
					"code":    http.StatusUnauthorized,
					"message": "未登录或token已过期",
				})
			}

			claims, err := jwtService.ValidateAccessToken(tokenStr)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]any{
					"code":    http.StatusUnauthorized,
					"message": "未登录或token已过期",
				})
			}

			c.Set(UserIDKey, claims.UserID)
			c.Set(UserRoleKey, claims.Role)
			return next(c)
		}
	}
}

// AdminRequired returns an Echo middleware that checks if the authenticated user
// has the "admin" role. This middleware MUST be used after AuthRequired.
//
// Returns 403 if the user is not an admin or if no user role is found in context.
func AdminRequired() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			role, ok := c.Get(UserRoleKey).(string)
			if !ok || role != "admin" {
				return c.JSON(http.StatusForbidden, map[string]any{
					"code":    http.StatusForbidden,
					"message": "需要管理员权限",
				})
			}
			return next(c)
		}
	}
}

// OptionalAuth returns an Echo middleware that optionally validates a JWT access token.
// If a valid token is present in the "Authorization: Bearer <token>" header, it injects
// the user ID and role into the context. If no token is present or the token is invalid,
// the request proceeds as a guest (no user info in context).
func OptionalAuth(jwtService *service.JWTService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get(echo.HeaderAuthorization)
			if authHeader == "" {
				return next(c)
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenStr == authHeader {
				return next(c)
			}

			claims, err := jwtService.ValidateAccessToken(tokenStr)
			if err != nil {
				// Invalid token — continue as guest
				return next(c)
			}

			c.Set(UserIDKey, claims.UserID)
			c.Set(UserRoleKey, claims.Role)
			return next(c)
		}
	}
}
