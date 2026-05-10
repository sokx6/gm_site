package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"gm_site/internal/logger"
	"gm_site/internal/middleware"
	"gm_site/internal/model"
	"gm_site/internal/repository"
	"gm_site/internal/service"
)

// AdminHandler handles admin-only HTTP requests for user review.
type AdminHandler struct {
	userRepo     *repository.UserRepository
	emailService service.EmailService
}

// NewAdminHandler creates a new AdminHandler with the given dependencies.
func NewAdminHandler(userRepo *repository.UserRepository, emailService service.EmailService) *AdminHandler {
	return &AdminHandler{
		userRepo:     userRepo,
		emailService: emailService,
	}
}

// ListPending handles GET /api/admin/users/pending.
//
// Returns all users with "pending" status. password_hash is excluded
// automatically via the json:"-" tag on model.User.
func (h *AdminHandler) ListPending(c echo.Context) error {
	users, err := h.userRepo.ListPending()
	if err != nil {
		middleware.GetLogger(c).Error("failed to list pending users", "err", err)
		return Error(c, http.StatusInternalServerError, "获取待审核用户列表失败")
	}
	return Success(c, users)
}

// ApproveUser handles PUT /api/admin/users/:id/approve.
func (h *AdminHandler) ApproveUser(c echo.Context) error {
	return h.updateUserStatus(c, model.UserStatusApproved)
}

// RejectUser handles PUT /api/admin/users/:id/reject.
func (h *AdminHandler) RejectUser(c echo.Context) error {
	return h.updateUserStatus(c, model.UserStatusRejected)
}

// updateUserStatus is the shared logic for approve/reject.
//
// It validates the target user exists, has not already been processed,
// updates the status, and asynchronously sends a notification email.
func (h *AdminHandler) updateUserStatus(c echo.Context, status string) error {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		middleware.GetLogger(c).Error("failed to parse user id", "err", err)
		return Error(c, http.StatusBadRequest, "无效的用户ID")
	}

	// Fetch user to verify existence and current status
	user, err := h.userRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Error(c, http.StatusNotFound, "用户不存在")
		}
		middleware.GetLogger(c).Error("failed to find user for status update", "err", err)
		return Error(c, http.StatusInternalServerError, "服务器内部错误")
	}

	// Prevent re-review of already processed users
	if user.Status != model.UserStatusPending {
		return Error(c, http.StatusBadRequest, "该用户已处理，不可重复审核")
	}

	// Update status
	if err := h.userRepo.UpdateStatus(id, status); err != nil {
		middleware.GetLogger(c).Error("failed to update user status", "err", err)
		return Error(c, http.StatusInternalServerError, "更新用户状态失败")
	}

	// Asynchronously send notification email
	go func() {
		if status == model.UserStatusApproved {
			if err := h.emailService.SendRegisterApprovedNotification(user.Email); err != nil {
				logger.L.Error("approval email failed", "err", err)
			}
		} else {
			if err := h.emailService.SendRegisterRejectedNotification(user.Email); err != nil {
				logger.L.Error("rejection email failed", "err", err)
			}
		}
	}()

	return Success(c, map[string]any{
		"id":     id,
		"status": status,
	})
}
