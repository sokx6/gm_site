package handler

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"

	"gm_site/internal/logger"
	"gm_site/internal/middleware"
	"gm_site/internal/model"
	"gm_site/internal/repository"
	"gm_site/internal/service"
)

// CommentHandler handles comment-related HTTP requests.
type CommentHandler struct {
	commentRepo      *repository.CommentRepository
	imageRepo        *repository.ImageRepository
	userRepo         *repository.UserRepository
	notificationRepo *repository.NotificationRepository
	emailSvc         service.EmailService
	db               *sql.DB
}

// NewCommentHandler creates a new CommentHandler with the given dependencies.
func NewCommentHandler(
	commentRepo *repository.CommentRepository,
	imageRepo *repository.ImageRepository,
	userRepo *repository.UserRepository,
	notificationRepo *repository.NotificationRepository,
	emailSvc service.EmailService,
	db *sql.DB,
) *CommentHandler {
	return &CommentHandler{
		commentRepo:      commentRepo,
		imageRepo:        imageRepo,
		userRepo:         userRepo,
		notificationRepo: notificationRepo,
		emailSvc:         emailSvc,
		db:               db,
	}
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
		middleware.GetLogger(c).Error("failed to parse image id", "err", err)
		return Error(c, http.StatusBadRequest, "无效的图片ID")
	}

	userID, ok := c.Get(middleware.UserIDKey).(int64)
	if !ok {
		return Error(c, http.StatusUnauthorized, "未登录")
	}

	var req model.CreateCommentRequest
	if err := c.Bind(&req); err != nil {
		middleware.GetLogger(c).Error("failed to bind comment request", "err", err)
		return Error(c, http.StatusBadRequest, "无效的请求数据")
	}
	if err := validate.Struct(req); err != nil {
		middleware.GetLogger(c).Error("comment validation failed", "err", err)
		return Error(c, http.StatusBadRequest, "评论内容不能为空且不超过500字")
	}

	// XSS sanitize
	content := sanitizeContent(req.Content)

	comment := &model.Comment{
		ImageID:  imageID,
		UserID:   userID,
		Content:  content,
		ParentID: req.ParentID,
	}

	if err := h.commentRepo.Create(comment); err != nil {
		middleware.GetLogger(c).Error("failed to create comment", "err", err)
		return Error(c, http.StatusInternalServerError, "创建评论失败")
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.L.Error("comment notification: panic recovered", "panic", r, "imageID", imageID, "userID", userID)
			}
		}()
		logger.L.Debug("comment notification: goroutine started", "imageID", imageID, "userID", userID)

		img, err := h.imageRepo.FindByID(imageID)
		if err != nil {
			logger.L.Error("comment notification: failed to find image", "err", err, "imageID", imageID)
			return
		}
		if img.UploadedBy == userID {
			logger.L.Debug("comment notification: self-comment, skipping", "imageID", imageID, "userID", userID)
			return
		}

		logger.L.Debug("comment notification: found image and uploader", "uploaderID", img.UploadedBy, "imageTitle", img.Title)

		uploader, err := h.userRepo.FindByID(img.UploadedBy)
		if err != nil {
			logger.L.Error("comment notification: failed to find uploader", "err", err, "uploaderID", img.UploadedBy)
			return
		}

		existing, err := h.notificationRepo.FindTodayByImageAndType(imageID, "image_comment")
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			logger.L.Error("comment notification: rate-limit check failed", "err", err)
			return
		}

		logger.L.Debug("comment notification: rate-limit check done", "existing", existing != nil)

		commenter, err := h.userRepo.FindByID(userID)
		if err != nil {
			logger.L.Error("comment notification: failed to find commenter", "err", err, "userID", userID)
			return
		}
		commenterName := commenter.Nickname

		if existing != nil {
			logger.L.Debug("comment notification: appending to existing", "notifID", existing.ID)
			newLine := fmt.Sprintf("\n%s: %s", commenterName, content)
			if err := h.notificationRepo.AppendContent(existing.ID, newLine); err != nil {
				logger.L.Error("comment notification: failed to append content", "err", err)
			}
		} else {
			logger.L.Debug("comment notification: creating new notification")
			notif := &model.Notification{
				UserID:  img.UploadedBy,
				Type:    "image_comment",
				Title:   "你的图片有新评论",
				Content: fmt.Sprintf("%s: %s", commenterName, content),
				ImageID: &imageID,
			}
			if err := h.notificationRepo.Create(notif); err != nil {
				logger.L.Error("comment notification: failed to create notification", "err", err)
				return
			}

			if err := h.emailSvc.SendImageCommentNotification(uploader.Email, commenterName, img.Title, content); err != nil {
				logger.L.Error("comment notification: failed to send email", "err", err, "to", uploader.Email)
			} else {
				logger.L.Debug("comment notification: email sent", "to", uploader.Email, "image", img.Title)
			}
		}
	}()

	return Created(c, comment)
}

// Reply handles POST /api/comments/:id/reply.
//
// It requires authentication. Creates a child comment with parent_id set.
func (h *CommentHandler) Reply(c echo.Context) error {
	idStr := c.Param("id")
	parentID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		middleware.GetLogger(c).Error("failed to parse parent comment id", "err", err)
		return Error(c, http.StatusBadRequest, "无效的评论ID")
	}

	userID, ok := c.Get(middleware.UserIDKey).(int64)
	if !ok {
		return Error(c, http.StatusUnauthorized, "未登录")
	}

	var req model.CreateCommentRequest
	if err := c.Bind(&req); err != nil {
		middleware.GetLogger(c).Error("failed to bind reply request", "err", err)
		return Error(c, http.StatusBadRequest, "无效的请求数据")
	}
	if err := validate.Struct(req); err != nil {
		middleware.GetLogger(c).Error("reply validation failed", "err", err)
		return Error(c, http.StatusBadRequest, "回复内容不能为空且不超过500字")
	}

	// Fetch parent to get image_id
	parent, err := h.commentRepo.FindByID(parentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Error(c, http.StatusNotFound, "父评论不存在")
		}
		middleware.GetLogger(c).Error("failed to find parent comment", "err", err)
		return Error(c, http.StatusInternalServerError, "服务器内部错误")
	}

	content := sanitizeContent(req.Content)

	comment := &model.Comment{
		ImageID:  parent.ImageID,
		UserID:   userID,
		Content:  content,
		ParentID: &parentID,
	}

	if err := h.commentRepo.Create(comment); err != nil {
		middleware.GetLogger(c).Error("failed to create reply", "err", err)
		return Error(c, http.StatusInternalServerError, "创建回复失败")
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
		middleware.GetLogger(c).Error("failed to parse image id", "err", err)
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

	parentIDStr := c.QueryParam("parent_id")
	var comments []model.Comment
	var total int

	if parentIDStr != "" {
		parentID, err := strconv.ParseInt(parentIDStr, 10, 64)
		if err != nil {
			middleware.GetLogger(c).Error("failed to parse parent_id", "err", err)
			return Error(c, http.StatusBadRequest, "无效的父评论ID")
		}
		comments, total, err = h.commentRepo.FindReplies(parentID, page, limit)
		if err != nil {
			middleware.GetLogger(c).Error("failed to list replies", "err", err)
			return Error(c, http.StatusInternalServerError, "获取回复列表失败")
		}
	} else {
		comments, total, err = h.commentRepo.FindByImageID(imageID, page, limit)
		if err != nil {
			middleware.GetLogger(c).Error("failed to list comments", "err", err)
			return Error(c, http.StatusInternalServerError, "获取评论列表失败")
		}
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
		middleware.GetLogger(c).Error("failed to parse comment id", "err", err)
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
		middleware.GetLogger(c).Error("failed to find comment for deletion", "err", err)
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
			if err != nil {
				middleware.GetLogger(c).Error("failed to query image uploader", "err", err)
			}
			return Error(c, http.StatusForbidden, "无权删除此评论")
		}
	}

	if err := h.commentRepo.Delete(id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Error(c, http.StatusNotFound, "评论不存在")
		}
		middleware.GetLogger(c).Error("failed to delete comment", "err", err)
		return Error(c, http.StatusInternalServerError, "删除评论失败")
	}

	return Success(c, nil)
}
