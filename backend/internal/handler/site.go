package handler

import (
	"github.com/labstack/echo/v4"
)

// SiteHandler handles site info endpoints.
type SiteHandler struct {
	name string
}

// NewSiteHandler creates a new SiteHandler.
func NewSiteHandler(name string) *SiteHandler {
	return &SiteHandler{name}
}

// Info returns the site name.
func (h *SiteHandler) Info(c echo.Context) error {
	return Success(c, map[string]string{"name": h.name})
}
