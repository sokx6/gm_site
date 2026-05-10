package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
)

// ANSI color constants for console output.
const (
	colorDebug = "\033[36m"
	colorInfo  = "\033[32m"
	colorWarn  = "\033[33m"
	colorError = "\033[31m"
	colorReset = "\033[0m"
)

// L is the global logger instance, initialized by Init.
var L *slog.Logger

// levelString returns a 5-character right-aligned level label.
func levelString(l slog.Level) string {
	switch l {
	case slog.LevelDebug:
		return "DEBUG"
	case slog.LevelInfo:
		return "INFO "
	case slog.LevelWarn:
		return "WARN "
	case slog.LevelError:
		return "ERROR"
	default:
		s := l.String()
		if len(s) < 5 {
			return strings.Repeat(" ", 5-len(s)) + s
		}
		return s[:5]
	}
}

// colorFor returns the ANSI color escape for the given level.
func colorFor(l slog.Level) string {
	switch l {
	case slog.LevelDebug:
		return colorDebug
	case slog.LevelInfo:
		return colorInfo
	case slog.LevelWarn:
		return colorWarn
	case slog.LevelError:
		return colorError
	default:
		return ""
	}
}

// coloredHandler implements slog.Handler with ANSI-colored console output.
type coloredHandler struct {
	writer io.Writer
	level  slog.Leveler
	attrs  []slog.Attr
	groups []string
}

// Enabled reports whether the handler handles records at the given level.
func (h *coloredHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level.Level()
}

// Handle formats and writes a log record with ANSI colors.
// Format: HH:MM:SS [LEVEL] [group.]message key=value key=value
func (h *coloredHandler) Handle(_ context.Context, r slog.Record) error {
	var buf strings.Builder

	// Time — use local time as HH:MM:SS
	buf.WriteString(r.Time.Format("15:04:05"))
	buf.WriteByte(' ')

	// Colored level label in brackets
	buf.WriteString(colorFor(r.Level))
	buf.WriteByte('[')
	buf.WriteString(levelString(r.Level))
	buf.WriteByte(']')
	buf.WriteString(colorReset)
	buf.WriteByte(' ')

	// Group prefix
	for _, g := range h.groups {
		buf.WriteString(g)
		buf.WriteByte('.')
	}

	// Message
	buf.WriteString(r.Message)

	// Record-scoped attributes
	r.Attrs(func(a slog.Attr) bool {
		buf.WriteByte(' ')
		buf.WriteString(a.Key)
		buf.WriteByte('=')
		buf.WriteString(a.Value.String())
		return true
	})

	// Handler-level attributes
	for _, a := range h.attrs {
		buf.WriteByte(' ')
		buf.WriteString(a.Key)
		buf.WriteByte('=')
		buf.WriteString(a.Value.String())
	}

	buf.WriteByte('\n')

	_, err := h.writer.Write([]byte(buf.String()))
	return err
}

// WithAttrs returns a new handler with the given attributes appended.
func (h *coloredHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}
	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)
	return &coloredHandler{
		writer: h.writer,
		level:  h.level,
		attrs:  newAttrs,
		groups: h.groups, // share the same underlying slice (immutable usage)
	}
}

// WithGroup returns a new handler with the given group name appended.
func (h *coloredHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	newGroups := make([]string, len(h.groups)+1)
	copy(newGroups, h.groups)
	newGroups[len(h.groups)] = name
	return &coloredHandler{
		writer: h.writer,
		level:  h.level,
		attrs:  h.attrs,
		groups: newGroups,
	}
}

// multiHandler fans out log operations to multiple handlers (console + file).
type multiHandler struct {
	handlers []slog.Handler
}

// Enabled reports whether any handler is enabled for the given level.
func (m *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range m.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

// Handle fans out the record to all enabled handlers.
// Uses Record.Clone() to prevent attribute consumption by one handler
// from affecting subsequent handlers.
func (m *multiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, h := range m.handlers {
		if h.Enabled(ctx, r.Level) {
			r2 := r.Clone()
			if err := h.Handle(ctx, r2); err != nil {
				return err
			}
		}
	}
	return nil
}

// WithAttrs fans out WithAttrs to all contained handlers.
func (m *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		newHandlers[i] = h.WithAttrs(attrs)
	}
	return &multiHandler{handlers: newHandlers}
}

// WithGroup fans out WithGroup to all contained handlers.
func (m *multiHandler) WithGroup(name string) slog.Handler {
	newHandlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		newHandlers[i] = h.WithGroup(name)
	}
	return &multiHandler{handlers: newHandlers}
}

// redactPassword replaces values for keys containing "password" or "secret"
// with "[REDACTED]". Used as slog.HandlerOptions.ReplaceAttr.
func redactPassword(groups []string, a slog.Attr) slog.Attr {
	key := a.Key
	lower := strings.ToLower(key)
	if strings.Contains(lower, "password") || strings.Contains(lower, "secret") {
		return slog.String(key, "[REDACTED]")
	}
	return a
}

// Init initializes the global logger with dual output:
//   - ANSI-colored console (debug level and above)
//   - JSON file with password redaction (info level and above)
//
// The log file parent directory is created if it does not exist.
// Panics if the directory cannot be created or the file cannot be opened.
func Init(logFilePath string) *slog.Logger {
	// Create parent directory if needed
	dir := logFilePath[:strings.LastIndex(logFilePath, "/")]
	if err := os.MkdirAll(dir, 0755); err != nil {
		panic("failed to create log directory: " + err.Error())
	}

	// Open log file in append mode (create if not exists)
	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		panic("failed to open log file: " + err.Error())
	}

	// Console handler — debug level, colored
	console := &coloredHandler{
		writer: os.Stdout,
		level:  slog.LevelDebug,
	}

	// File handler — info level, JSON, redact sensitive fields
	fileHandler := slog.NewJSONHandler(file, &slog.HandlerOptions{
		Level:       slog.LevelInfo,
		ReplaceAttr: redactPassword,
	})

	// Combine handlers
	handler := &multiHandler{
		handlers: []slog.Handler{console, fileHandler},
	}

	logger := slog.New(handler)
	L = logger
	return logger
}

// Ensure coloredHandler implements slog.Handler.
var _ slog.Handler = (*coloredHandler)(nil)

// Ensure multiHandler implements slog.Handler.
var _ slog.Handler = (*multiHandler)(nil)
