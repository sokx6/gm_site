package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"

	"gm_site/internal/middleware"
	"gm_site/internal/model"
	"gm_site/internal/repository"
)

// CommentHandler handles comment-related HTTP requests.
type CommentHandler struct {
	commentRepo *repository.CommentRepository
	db          *sql.DB
}

// NewCommentHandler creates a new CommentHandler with the given dependencies.
func NewCommentHandler(commentRepo *repository.CommentRepository, db *sql.DB) *CommentHandler {
	return &CommentHandler{commentRepo: commentRepo, db: db}
}

// sanitizeContent performs basic XSS protection by escaping angle brackets.
func sanitizeContent(content string) string {
	content = strings.ReplaceAll(content, "<", "&lt;")
	content = strings.ReplaceAll(content, ">", "&gt;")
	return content
}

// commentListResponse is the response body for listing comments.
type commentListResponse struct {
	Comments []model.Comment `json:"comments"`
	Total    int             `json:"total"`
	Page     int             `json:"page"`
	Limit    int             `json:"limit"`
}

// Create handles POST /api/images/:id/comments.
//
// It requires authentication. The comment content is sanitized for XSS.
func (h *CommentHandler) Create(c echo.Context) error {
	imageIDStr := c.Param("id")
	imageID, err := strconv.ParseInt(imageIDStr, 10, 64)
	if err != nil {
		return Error(c, http.StatusBadRequest, "无效的图片ID")
	}

	userID, ok := c.Get(middleware.UserIDKey).(int64)
	if !ok {
		return Error(c, http.StatusUnauthorized, "未登录")
	}

	var req model.CreateCommentRequest
	if err := c.Bind(&req); err != nil {
		return Error(c, http.StatusBadRequest, "无效的请求数据")
	}
	if err := validate.Struct(req); err != nil {
		return Error(c, http.StatusBadRequest, "评论内容不能为空且不超过500字")
	}

	// XSS sanitize
	content := sanitizeContent(req.Content)

	comment := &model.Comment{
		ImageID: imageID,
		UserID:  userID,
		Content: content,
	}

	if err := h.commentRepo.Create(comment); err != nil {
		return Error(c, http.StatusInternalServerError, "创建评论失败")
	}

	return Created(c, comment)
}

// ListByImage handles GET /api/images/:id/comments.
//
// It is publicly accessible. Supports pagination via query params.
func (h *CommentHandler) ListByImage(c echo.Context) error {
	imageIDStr := c.Param("id")
	imageID, err := strconv.ParseInt(imageIDStr, 10, 64)
	if err != nil {
		return Error(c, http.StatusBadRequest, "无效的图片ID")
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	comments, total, err := h.commentRepo.FindByImageID(imageID, page, limit)
	if err != nil {
		return Error(c, http.StatusInternalServerError, "获取评论列表失败")
	}

	resp := commentListResponse{
		Comments: comments,
		Total:    total,
		Page:     page,
		Limit:    limit,
	}
	return Success(c, resp)
}

// Delete handles DELETE /api/comments/:id.
//
// It requires authentication. Only the comment author, the image uploader,
// or an admin can delete a comment.
func (h *CommentHandler) Delete(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return Error(c, http.StatusBadRequest, "无效的评论ID")
	}

	userID, ok := c.Get(middleware.UserIDKey).(int64)
	if !ok {
		return Error(c, http.StatusUnauthorized, "未登录")
	}
	role, _ := c.Get(middleware.UserRoleKey).(string)

	// Fetch the comment
	comment, err := h.commentRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Error(c, http.StatusNotFound, "评论不存在")
		}
		return Error(c, http.StatusInternalServerError, "服务器内部错误")
	}

	// Permission check: comment author OR image uploader OR admin
	if comment.UserID == userID || role == "admin" {
		// Allowed
	} else {
		// Check if current user is the image uploader
		var uploaderID int64
		err := h.db.QueryRow(
			`SELECT uploaded_by FROM images WHERE id = ?`, comment.ImageID,
		).Scan(&uploaderID)
		if err != nil || uploaderID != userID {
			return Error(c, http.StatusForbidden, "无权删除此评论")
		}
	}

	if err := h.commentRepo.Delete(id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Error(c, http.StatusNotFound, "评论不存在")
		}
		return Error(c, http.StatusInternalServerError, "删除评论失败")
	}

	return Success(c, nil)
}
