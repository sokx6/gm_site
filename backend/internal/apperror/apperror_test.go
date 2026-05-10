package apperror

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewNotFound(t *testing.T) {
	ae := NewNotFound("resource not found", fmt.Errorf("db: %w", assert.AnError))
	assert.Equal(t, http.StatusNotFound, ae.Code)
	assert.Equal(t, "resource not found", ae.Message)
	assert.ErrorIs(t, ae.Cause, assert.AnError)
}

func TestNewInternal(t *testing.T) {
	ae := NewInternal("internal error", fmt.Errorf("upload: %w", assert.AnError))
	assert.Equal(t, http.StatusInternalServerError, ae.Code)
	assert.Equal(t, "internal error", ae.Message)
	assert.ErrorIs(t, ae.Cause, assert.AnError)
}

func TestNewValidation(t *testing.T) {
	ae := NewValidation("invalid input", nil)
	assert.Equal(t, http.StatusBadRequest, ae.Code)
	assert.Equal(t, "invalid input", ae.Message)
	assert.Nil(t, ae.Cause)
}

func TestNewAuth(t *testing.T) {
	ae := NewAuth("unauthorized", nil)
	assert.Equal(t, http.StatusUnauthorized, ae.Code)
	assert.Equal(t, "unauthorized", ae.Message)
	assert.Nil(t, ae.Cause)
}

func TestErrorMethod(t *testing.T) {
	ae := NewNotFound("item missing", nil)
	assert.Equal(t, "item missing", ae.Error())
}

func TestUnwrap(t *testing.T) {
	inner := fmt.Errorf("inner error")
	ae := NewInternal("wrapped", inner)

	assert.Equal(t, inner, ae.Unwrap())
	assert.ErrorIs(t, ae, inner)
}

func TestErrorsUnwrap(t *testing.T) {
	inner := fmt.Errorf("inner error")
	ae := NewInternal("wrapped", inner)

	assert.Equal(t, inner, errors.Unwrap(ae))
}

func TestUnwrapNilCause(t *testing.T) {
	ae := NewNotFound("no cause", nil)
	assert.Nil(t, ae.Unwrap())
	assert.Nil(t, ae.Unwrap())
}

func TestStatusCodeFindsAppError(t *testing.T) {
	inner := fmt.Errorf("inner error")
	ae := NewNotFound("not found", inner)

	assert.Equal(t, http.StatusNotFound, StatusCode(ae))

	wrapped := fmt.Errorf("context: %w", ae)
	assert.Equal(t, http.StatusNotFound, StatusCode(wrapped))
}

func TestStatusCodeReturns500WhenNoAppError(t *testing.T) {
	plainErr := errors.New("plain error")
	assert.Equal(t, http.StatusInternalServerError, StatusCode(plainErr))

	assert.Equal(t, http.StatusInternalServerError, StatusCode(nil))
}

func TestStatusCodeWalksChain(t *testing.T) {
	inner := errors.New("root")
	ae := NewInternal("mid", inner)
	wrapped := fmt.Errorf("top: %w", ae)

	assert.Equal(t, http.StatusInternalServerError, StatusCode(wrapped))
}

func TestErrorIsChain(t *testing.T) {
	root := errors.New("root cause")
	ae := NewInternal("internal", root)
	wrapped := fmt.Errorf("wrap: %w", ae)

	assert.True(t, errors.Is(wrapped, root))
	assert.True(t, errors.Is(wrapped, ae))
}
