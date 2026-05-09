package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"gm_site/internal/model"
)

// JSON writes a standard API response with the given status code, message, and data.
func JSON(c echo.Context, code int, message string, data interface{}) error {
	return c.JSON(code, model.APIResponse{
		Code:    code,
		Message: message,
		Data:    data,
	})
}

// Success writes a 200 response with the given data.
func Success(c echo.Context, data interface{}) error {
	return JSON(c, http.StatusOK, "success", data)
}

// Created writes a 201 response with the given data.
func Created(c echo.Context, data interface{}) error {
	return JSON(c, http.StatusCreated, "created", data)
}

// Error writes an error response with the given status code and message.
func Error(c echo.Context, code int, message string) error {
	return JSON(c, code, message, nil)
}
