// Package apperror provides custom error types for HTTP API error handling.
//
// AppError wraps a message, HTTP status code, and an optional underlying cause,
// enabling structured error responses and standard Go error-chain traversal
// via errors.Is / errors.As / errors.Unwrap.
package apperror

import (
	"errors"
	"net/http"
)

// AppError is a structured error that carries an HTTP status code, a
// human-readable message, and an optional underlying cause.
//
// It implements:
//   - error (Error() returns Message)
//   - interface{ Unwrap() error } for error-chain traversal
type AppError struct {
	Code    int    // HTTP status code
	Message string // Human-readable message
	Cause   error  // Underlying cause (may be nil)
}

// Error returns the human-readable message. This makes AppError satisfy the
// standard error interface so it can be used anywhere an error is expected.
func (e *AppError) Error() string {
	return e.Message
}

// Unwrap returns the underlying cause of this error. It implements the
// single-Unwrap interface used by errors.Is, errors.As, and errors.Unwrap
// to walk the error chain.
func (e *AppError) Unwrap() error {
	return e.Cause
}

// newAppError is the internal constructor used by all factory functions.
func newAppError(code int, msg string, cause error) *AppError {
	return &AppError{
		Code:    code,
		Message: msg,
		Cause:   cause,
	}
}

// NewNotFound creates a 404 Not Found error.
func NewNotFound(msg string, cause error) *AppError {
	return newAppError(http.StatusNotFound, msg, cause)
}

// NewValidation creates a 400 Bad Request / validation error.
func NewValidation(msg string, cause error) *AppError {
	return newAppError(http.StatusBadRequest, msg, cause)
}

// NewAuth creates a 401 Unauthorized / authentication error.
func NewAuth(msg string, cause error) *AppError {
	return newAppError(http.StatusUnauthorized, msg, cause)
}

// NewInternal creates a 500 Internal Server Error.
func NewInternal(msg string, cause error) *AppError {
	return newAppError(http.StatusInternalServerError, msg, cause)
}

// StatusCode walks the error chain via errors.As looking for an *AppError.
// If found its Code field is returned. Otherwise 500 (Internal Server Error)
// is returned as a safe default.
func StatusCode(err error) int {
	var ae *AppError
	if errors.As(err, &ae) {
		return ae.Code
	}
	return http.StatusInternalServerError
}
