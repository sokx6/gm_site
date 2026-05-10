package handler

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"

	"gm_site/internal/middleware"
	"gm_site/internal/model"
	"gm_site/internal/repository"
	"gm_site/internal/service"
)

// ImageHandler handles image-related HTTP requests.
type ImageHandler struct {
	imageRepo  *repository.ImageRepository
	lskyClient *service.LskyClient
	maxSizeMB  int
}

// NewImageHandler creates a new ImageHandler with the given dependencies.
func NewImageHandler(imageRepo *repository.ImageRepository, lskyClient *service.LskyClient, maxSizeMB int) *ImageHandler {
	return &ImageHandler{
		imageRepo:  imageRepo,
		lskyClient: lskyClient,
		maxSizeMB:  maxSizeMB,
	}
}

// UploadImage handles POST /api/images/upload. Requires authentication.
//
// Accepts multipart/form-data with fields:
//   - file: the image file (required)
//   - title: image title (required)
//   - album_id: optional album ID
//   - tags: comma-separated tag string
//
// Validates file size and MIME type before uploading to Lsky and saving to DB.
func (h *ImageHandler) UploadImage(c echo.Context) error {
	// 1. Parse multipart form fields
	fileHeader, err := c.FormFile("file")
	if err != nil {
		middleware.GetLogger(c).Error("failed to get form file", "err", err)
		return Error(c, http.StatusBadRequest, "请选择图片文件")
	}

	title := c.FormValue("title")
	if title == "" {
		return Error(c, http.StatusBadRequest, "标题不能为空")
	}

	// Parse optional album_id
	var albumID *int64
	if albumIDStr := c.FormValue("album_id"); albumIDStr != "" {
		id, err := strconv.ParseInt(albumIDStr, 10, 64)
		if err != nil {
			middleware.GetLogger(c).Error("failed to parse album_id", "err", err)
			return Error(c, http.StatusBadRequest, "无效的相册ID")
		}
		albumID = &id
	}

	// Parse tags from comma-separated string
	tags := make([]string, 0)
	if tagsStr := c.FormValue("tags"); tagsStr != "" {
		for _, t := range strings.Split(tagsStr, ",") {
			trimmed := strings.TrimSpace(t)
			if trimmed != "" {
				tags = append(tags, trimmed)
			}
		}
	}

	// Parse optional privacy (defaults to "public")
	privacy := c.FormValue("privacy")
	if privacy == "" {
		privacy = "public"
	}
	validPrivacy := map[string]bool{
		"public":  true,
		"friends": true,
		"private": true,
	}
	if !validPrivacy[privacy] {
		return Error(c, http.StatusBadRequest, "无效的隐私设置，仅支持: public/friends/private")
	}

	// 2. Validate file size
	maxBytes := int64(h.maxSizeMB) * 1024 * 1024
	if fileHeader.Size > maxBytes {
		return Error(c, http.StatusRequestEntityTooLarge, fmt.Sprintf("文件大小不能超过%dMB", h.maxSizeMB))
	}

	// 3. Validate MIME type by reading first 512 bytes
	file, err := fileHeader.Open()
	if err != nil {
		middleware.GetLogger(c).Error("failed to open uploaded file", "err", err)
		return Error(c, http.StatusInternalServerError, "无法读取文件")
	}
	defer file.Close()

	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		middleware.GetLogger(c).Error("failed to read file content for MIME detection", "err", err)
		return Error(c, http.StatusBadRequest, "无法读取文件内容")
	}

	contentType := http.DetectContentType(buf[:n])
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}
	if !allowedTypes[contentType] {
		return Error(c, http.StatusBadRequest, "不支持的图片格式，仅支持 JPEG/PNG/GIF/WebP")
	}

	// Seek back to beginning before passing to Lsky uploader
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		middleware.GetLogger(c).Error("failed to seek file to beginning", "err", err)
		return Error(c, http.StatusInternalServerError, "无法处理文件")
	}

	// 4. Upload to Lsky
	lskyURL, err := h.lskyClient.UploadImage(file, fileHeader)
	if err != nil {
		middleware.GetLogger(c).Error("image upload to lsky failed", "err", err)
		return Error(c, http.StatusInternalServerError, fmt.Sprintf("图片上传失败: %v", err))
	}

	// 5. Save to DB
	userID, ok := c.Get(middleware.UserIDKey).(int64)
	if !ok {
		return Error(c, http.StatusUnauthorized, "未登录")
	}

	image := &model.Image{
		AlbumID:    albumID,
		Title:      title,
		Tags:       tags,
		LskyURL:    lskyURL,
		UploadedBy: userID,
		Privacy:    privacy,
	}

	if err := h.imageRepo.Create(image); err != nil {
		middleware.GetLogger(c).Error("failed to save image to database", "err", err)
		return Error(c, http.StatusInternalServerError, "保存图片信息失败")
	}

	return Created(c, image)
}

// ListImages handles GET /api/images. Public.
//
// Query params:
//   - page (default 1)
//   - limit (default 12)
//   - album_id (optional)
//   - tag (optional)
//
// Returns paginated list with total count.
func (h *ImageHandler) ListImages(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit < 1 || limit > 100 {
		limit = 12
	}

	var albumID *int64
	if albumIDStr := c.QueryParam("album_id"); albumIDStr != "" {
		id, err := strconv.ParseInt(albumIDStr, 10, 64)
		if err != nil {
			middleware.GetLogger(c).Error("failed to parse album_id in query", "err", err)
			return Error(c, http.StatusBadRequest, "无效的相册ID")
		}
		albumID = &id
	}

	tag := c.QueryParam("tag")

	viewerID, _ := c.Get(middleware.UserIDKey).(int64)
	isAdmin := false
	if role, ok := c.Get(middleware.UserRoleKey).(string); ok && role == model.UserRoleAdmin {
		isAdmin = true
	}

	images, total, err := h.imageRepo.FindAll(page, limit, albumID, tag, viewerID, isAdmin)
	if err != nil {
		middleware.GetLogger(c).Error("failed to list images", "err", err)
		return Error(c, http.StatusInternalServerError, "获取图片列表失败")
	}

	return Success(c, map[string]any{
		"list":  images,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// GetImage handles GET /api/images/:id. Public.
//
// Returns single image detail including comment count.
func (h *ImageHandler) GetImage(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		middleware.GetLogger(c).Error("failed to parse image id", "err", err)
		return Error(c, http.StatusBadRequest, "无效的图片ID")
	}

	image, err := h.imageRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Error(c, http.StatusNotFound, "图片不存在")
		}
		middleware.GetLogger(c).Error("failed to get image by id", "err", err)
		return Error(c, http.StatusInternalServerError, "服务器内部错误")
	}

	return Success(c, image)
}

// UpdateImage handles PUT /api/images/:id. Requires authentication.
//
// Only the uploader or an admin can update the image.
// Request body: {"title": "...", "description": "...", "tags": ["..."], "album_id": 1}
func (h *ImageHandler) UpdateImage(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		middleware.GetLogger(c).Error("failed to parse image id", "err", err)
		return Error(c, http.StatusBadRequest, "无效的图片ID")
	}

	userID, ok := c.Get(middleware.UserIDKey).(int64)
	if !ok {
		return Error(c, http.StatusUnauthorized, "未登录")
	}
	role, _ := c.Get(middleware.UserRoleKey).(string)

	image, err := h.imageRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Error(c, http.StatusNotFound, "图片不存在")
		}
		middleware.GetLogger(c).Error("failed to get image for update", "err", err)
		return Error(c, http.StatusInternalServerError, "服务器内部错误")
	}

	// Only the uploader or an admin can update
	if image.UploadedBy != userID && role != model.UserRoleAdmin {
		return Error(c, http.StatusForbidden, "没有权限编辑此图片")
	}

	var req struct {
		Title       string   `json:"title"`
		Description string   `json:"description"`
		Tags        []string `json:"tags"`
		AlbumID     *int64   `json:"album_id"`
	}
	if err := c.Bind(&req); err != nil {
		middleware.GetLogger(c).Error("failed to bind update image request", "err", err)
		return Error(c, http.StatusBadRequest, "无效的请求参数")
	}

	image.Title = req.Title
	image.Description = req.Description
	image.Tags = req.Tags
	image.AlbumID = req.AlbumID

	if err := h.imageRepo.Update(image); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Error(c, http.StatusNotFound, "图片不存在")
		}
		middleware.GetLogger(c).Error("failed to update image", "err", err)
		return Error(c, http.StatusInternalServerError, "更新图片失败")
	}

	return Success(c, image)
}

// DeleteImage handles DELETE /api/images/:id. Requires authentication.
//
// Only the uploader or an admin can delete the image.
func (h *ImageHandler) DeleteImage(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		middleware.GetLogger(c).Error("failed to parse image id", "err", err)
		return Error(c, http.StatusBadRequest, "无效的图片ID")
	}

	userID, ok := c.Get(middleware.UserIDKey).(int64)
	if !ok {
		return Error(c, http.StatusUnauthorized, "未登录")
	}
	role, _ := c.Get(middleware.UserRoleKey).(string)

	image, err := h.imageRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Error(c, http.StatusNotFound, "图片不存在")
		}
		middleware.GetLogger(c).Error("failed to get image for deletion", "err", err)
		return Error(c, http.StatusInternalServerError, "服务器内部错误")
	}

	// Only the uploader or an admin can delete
	if image.UploadedBy != userID && role != model.UserRoleAdmin {
		return Error(c, http.StatusForbidden, "没有权限删除此图片")
	}

	if err := h.imageRepo.Delete(id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Error(c, http.StatusNotFound, "图片不存在")
		}
		middleware.GetLogger(c).Error("failed to delete image", "err", err)
		return Error(c, http.StatusInternalServerError, "删除图片失败")
	}

	return Success(c, map[string]any{"deleted": true})
}

// UpdatePrivacy handles PUT /api/images/:id/privacy. Requires authentication.
//
// Only the uploader or an admin can change the privacy setting.
// Request body: {"privacy": "public"|"friends"|"private"}
func (h *ImageHandler) UpdatePrivacy(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		middleware.GetLogger(c).Error("failed to parse image id", "err", err)
		return Error(c, http.StatusBadRequest, "无效的图片ID")
	}

	userID, ok := c.Get(middleware.UserIDKey).(int64)
	if !ok {
		return Error(c, http.StatusUnauthorized, "未登录")
	}
	role, _ := c.Get(middleware.UserRoleKey).(string)

	image, err := h.imageRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Error(c, http.StatusNotFound, "图片不存在")
		}
		middleware.GetLogger(c).Error("failed to get image for privacy update", "err", err)
		return Error(c, http.StatusInternalServerError, "服务器内部错误")
	}

	// Only the uploader or an admin can update privacy
	if image.UploadedBy != userID && role != model.UserRoleAdmin {
		return Error(c, http.StatusForbidden, "没有权限修改此图片的隐私设置")
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

	if err := h.imageRepo.UpdatePrivacy(id, req.Privacy); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Error(c, http.StatusNotFound, "图片不存在")
		}
		middleware.GetLogger(c).Error("failed to update image privacy", "err", err)
		return Error(c, http.StatusInternalServerError, "更新隐私设置失败")
	}

	image.Privacy = req.Privacy
	return Success(c, image)
}

// SearchImages handles GET /api/images/search?q=keyword&page=1&limit=12.
// Public endpoint. Searches by title, description, and tags.
func (h *ImageHandler) SearchImages(c echo.Context) error {
	q := c.QueryParam("q")
	if q == "" {
		return Error(c, http.StatusBadRequest, "搜索关键词不能为空")
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit < 1 || limit > 100 {
		limit = 12
	}

	viewerID, _ := c.Get(middleware.UserIDKey).(int64)
	isAdmin := false
	if role, ok := c.Get(middleware.UserRoleKey).(string); ok && role == model.UserRoleAdmin {
		isAdmin = true
	}

	images, total, err := h.imageRepo.SearchImages(q, page, limit, viewerID, isAdmin)
	if err != nil {
		middleware.GetLogger(c).Error("failed to search images", "err", err)
		return Error(c, http.StatusInternalServerError, "搜索图片失败")
	}

	return Success(c, map[string]interface{}{
		"images": images,
		"total":  total,
		"page":   page,
		"limit":  limit,
	})
}
