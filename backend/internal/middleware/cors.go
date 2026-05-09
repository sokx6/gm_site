package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// CORS returns an Echo middleware that sets CORS headers for all requests.
// allowedOrigins is a comma-separated list of allowed origins.
// If empty, defaults to allowing the Vite dev server (http://localhost:5173).
// Handles OPTIONS preflight requests by returning 204 No Content.
func CORS(allowedOrigins string) echo.MiddlewareFunc {
	origins := parseOrigins(allowedOrigins)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			origin := c.Request().Header.Get("Origin")
			allowedOrigin := resolveAllowedOrigin(origin, origins)

			c.Response().Header().Set("Access-Control-Allow-Origin", allowedOrigin)
			c.Response().Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Response().Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			c.Response().Header().Set("Access-Control-Allow-Credentials", "true")
			if c.Request().Method == http.MethodOptions {
				return c.NoContent(http.StatusNoContent)
			}
			return next(c)
		}
	}
}

// parseOrigins splits a comma-separated origin string into a slice.
// Returns a default slice containing localhost:5173 if the input is empty.
func parseOrigins(s string) []string {
	if strings.TrimSpace(s) == "" {
		return []string{"http://localhost:5173"}
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// resolveAllowedOrigin checks if the request origin is in the allowed list.
// If matched, returns the origin; otherwise returns the first allowed origin.
// If the list contains "*", returns "*".
func resolveAllowedOrigin(origin string, origins []string) string {
	for _, o := range origins {
		if o == "*" {
			return "*"
		}
		if o == origin {
			return origin
		}
	}
	if len(origins) > 0 {
		return origins[0]
	}
	return "http://localhost:5173"
}
