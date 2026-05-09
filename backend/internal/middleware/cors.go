package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// CORS returns an Echo middleware that sets CORS headers for all requests.
// It allows requests from the Vite dev server (http://localhost:5173) and
// handles OPTIONS preflight requests by returning 204 No Content.
func CORS() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
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
