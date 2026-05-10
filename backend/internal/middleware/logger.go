package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"time"

	"github.com/labstack/echo/v4"

	"gm_site/internal/logger"
)

// LoggerMiddleware returns an Echo middleware that injects a request-scoped
// slog.Logger into the context and logs the completed request afterwards.
func LoggerMiddleware(l *slog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			reqID := shortUUID()

			// Create request-scoped logger with request metadata
			requestLogger := l.With(
				"request_id", reqID,
				"method", c.Request().Method,
				"path", c.Path(),
			)

			// Store in context for downstream handlers
			c.Set("logger", requestLogger)

			// Process request
			err := next(c)

			// Log completion
			requestLogger.Info("request completed",
				"status", c.Response().Status,
				"duration", time.Since(start),
			)

			return err
		}
	}
}

// GetLogger retrieves the request-scoped logger from the Echo context.
// Falls back to the global logger [logger.L] if the context is nil or no
// logger was injected (e.g. when used outside of LoggerMiddleware).
func GetLogger(c echo.Context) *slog.Logger {
	if c == nil {
		return logger.L
	}
	if l, ok := c.Get("logger").(*slog.Logger); ok {
		return l
	}
	return logger.L
}

// shortUUID generates an 8-character hex string from crypto/rand.
func shortUUID() string {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return "00000000"
	}
	return hex.EncodeToString(b)
}
