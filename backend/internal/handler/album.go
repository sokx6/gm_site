package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"gm_site/internal/middleware"
	"gm_site/internal/model"
	"gm_site/internal/repository"
)

// AlbumHandler handles album-related HTTP requests.
type AlbumHandler struct {
	albumRepo *repository.AlbumRepository
}

// NewAlbumHandler creates a new AlbumHandler with the given dependencies.
func NewAlbumHandler(albumRepo *repository.AlbumRepository) *AlbumHandler {
	return &AlbumHandler{albumRepo: albumRepo}
}

// ListAlbums handles GET /api/albums.
//
// Returns all albums ordered by created_at descending. This endpoint is public.
// OptionalAuth middleware injects user info when a valid token is present.
func (h *AlbumHandler) ListAlbums(c echo.Context) error {
	viewerID, _ := c.Get(middleware.UserIDKey).(int64)
	isAdmin := false
	if role, ok := c.Get(middleware.UserRoleKey).(string); ok && role == model.UserRoleAdmin {
		isAdmin = true
	}

	albums, err := h.albumRepo.FindAll(viewerID, isAdmin)
	if err != nil {
		middleware.GetLogger(c).Error("failed to list albums", "err", err)
		return Error(c, http.StatusInternalServerError, "获取相册列表失败")
	}
	return Success(c, albums)
}

// CreateAlbum handles POST /api/albums. Requires authentication.
//
// Request body: {"name": "...", "description": "..."}
// The album is created with the authenticated user as the creator.
func (h *AlbumHandler) CreateAlbum(c echo.Context) error {
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := c.Bind(&req); err != nil {
		middleware.GetLogger(c).Error("failed to bind create album request", "err", err)
		return Error(c, http.StatusBadRequest, "无效的请求参数")
	}
	if req.Name == "" {
		return Error(c, http.StatusBadRequest, "相册名称不能为空")
	}

	userID, ok := c.Get(middleware.UserIDKey).(int64)
	if !ok {
		return Error(c, http.StatusUnauthorized, "未登录")
	}

	album := &model.Album{
		Name:        req.Name,
		Description: req.Description,
		CreatedBy:   userID,
	}

	if err := h.albumRepo.Create(album); err != nil {
		middleware.GetLogger(c).Error("failed to create album", "err", err)
		return Error(c, http.StatusInternalServerError, "创建相册失败")
	}

	return Created(c, album)
}

// UpdateAlbum handles PUT /api/albums/:id. Requires authentication.
//
// Only the album creator or an admin can edit the album.
// Request body: {"name": "...", "description": "..."}
func (h *AlbumHandler) UpdateAlbum(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		middleware.GetLogger(c).Error("failed to parse album id", "err", err)
		return Error(c, http.StatusBadRequest, "无效的相册ID")
	}

	userID, ok := c.Get(middleware.UserIDKey).(int64)
	if !ok {
		return Error(c, http.StatusUnauthorized, "未登录")
	}
	role, _ := c.Get(middleware.UserRoleKey).(string)

	album, err := h.albumRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Error(c, http.StatusNotFound, "相册不存在")
		}
		middleware.GetLogger(c).Error("failed to find album for update", "err", err)
		return Error(c, http.StatusInternalServerError, "服务器内部错误")
	}

	// Only the creator or an admin can edit
	if album.CreatedBy != userID && role != model.UserRoleAdmin {
		return Error(c, http.StatusForbidden, "没有权限编辑此相册")
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := c.Bind(&req); err != nil {
		middleware.GetLogger(c).Error("failed to bind update album request", "err", err)
		return Error(c, http.StatusBadRequest, "无效的请求参数")
	}

	album.Name = req.Name
	album.Description = req.Description

	if err := h.albumRepo.Update(album); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Error(c, http.StatusNotFound, "相册不存在")
		}
		middleware.GetLogger(c).Error("failed to update album", "err", err)
		return Error(c, http.StatusInternalServerError, "更新相册失败")
	}

	return Success(c, album)
}

// UpdatePrivacy handles PUT /api/albums/:id/privacy. Requires authentication.
//
// Only the album creator or an admin can change the privacy setting.
// Request body: {"privacy": "public"|"friends"|"private"}
func (h *AlbumHandler) UpdatePrivacy(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		middleware.GetLogger(c).Error("failed to parse album id", "err", err)
		return Error(c, http.StatusBadRequest, "无效的相册ID")
	}

	userID, ok := c.Get(middleware.UserIDKey).(int64)
	if !ok {
		return Error(c, http.StatusUnauthorized, "未登录")
	}
	role, _ := c.Get(middleware.UserRoleKey).(string)

	album, err := h.albumRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Error(c, http.StatusNotFound, "相册不存在")
		}
		middleware.GetLogger(c).Error("failed to get album for privacy update", "err", err)
		return Error(c, http.StatusInternalServerError, "服务器内部错误")
	}

	// Only the creator or an admin can update privacy
	if album.CreatedBy != userID && role != model.UserRoleAdmin {
		return Error(c, http.StatusForbidden, "没有权限修改此相册的隐私设置")
	}

	var req struct {
		Privacy string `json:"privacy"`
	}
	if err := c.Bind(&req); err != nil {
		middleware.GetLogger(c).Error("failed to bind privacy update request", "err", err)
		return Error(c, http.StatusBadRequest, "无效的请求参数")
	}

	// Validate privacy value
	validPrivacy := map[string]bool{
		"public":  true,
		"friends": true,
		"private": true,
	}
	if !validPrivacy[req.Privacy] {
		return Error(c, http.StatusBadRequest, "无效的隐私设置，仅支持: public/friends/private")
	}

	if err := h.albumRepo.UpdatePrivacy(id, req.Privacy); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Error(c, http.StatusNotFound, "相册不存在")
		}
		middleware.GetLogger(c).Error("failed to update album privacy", "err", err)
		return Error(c, http.StatusInternalServerError, "更新隐私设置失败")
	}

	album.Privacy = req.Privacy
	return Success(c, album)
}

// DeleteAlbum handles DELETE /api/albums/:id. Requires authentication.
//
// Only the album creator or an admin can delete the album.
// Returns 400 if the album still has associated images.
func (h *AlbumHandler) DeleteAlbum(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		middleware.GetLogger(c).Error("failed to parse album id", "err", err)
		return Error(c, http.StatusBadRequest, "无效的相册ID")
	}

	userID, ok := c.Get(middleware.UserIDKey).(int64)
	if !ok {
		return Error(c, http.StatusUnauthorized, "未登录")
	}
	role, _ := c.Get(middleware.UserRoleKey).(string)

	album, err := h.albumRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Error(c, http.StatusNotFound, "相册不存在")
		}
		middleware.GetLogger(c).Error("failed to find album for deletion", "err", err)
		return Error(c, http.StatusInternalServerError, "服务器内部错误")
	}

	// Only the creator or an admin can delete
	if album.CreatedBy != userID && role != model.UserRoleAdmin {
		return Error(c, http.StatusForbidden, "没有权限删除此相册")
	}

	if err := h.albumRepo.Delete(id); err != nil {
		if errors.Is(err, repository.ErrAlbumHasImages) {
			return Error(c, http.StatusBadRequest, "相册下有图片，无法删除")
		}
		if errors.Is(err, sql.ErrNoRows) {
			return Error(c, http.StatusNotFound, "相册不存在")
		}
		middleware.GetLogger(c).Error("failed to delete album", "err", err)
		return Error(c, http.StatusInternalServerError, "删除相册失败")
	}

	return Success(c, map[string]any{"deleted": true})
}
