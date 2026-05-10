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
	"gm_site/internal/service"
)

// FriendHandler handles friend-related and notification HTTP requests.
type FriendHandler struct {
	friendSvc        *service.FriendService
	userRepo         *repository.UserRepository
	notificationRepo *repository.NotificationRepository
}

// NewFriendHandler creates a new FriendHandler with the given dependencies.
func NewFriendHandler(
	friendSvc *service.FriendService,
	userRepo *repository.UserRepository,
	notificationRepo *repository.NotificationRepository,
) *FriendHandler {
	return &FriendHandler{
		friendSvc:        friendSvc,
		userRepo:         userRepo,
		notificationRepo: notificationRepo,
	}
}

// SendRequest handles POST /api/friends/request.
//
// It requires authentication. Creates a friend request from the current user
// to the user with the specified email.
func (h *FriendHandler) SendRequest(c echo.Context) error {
	userID, ok := c.Get(middleware.UserIDKey).(int64)
	if !ok {
		return Error(c, http.StatusUnauthorized, "未登录")
	}

	var req model.SendFriendRequest
	if err := c.Bind(&req); err != nil {
		middleware.GetLogger(c).Error("failed to bind friend request", "err", err)
		return Error(c, http.StatusBadRequest, "无效的请求数据")
	}
	if err := validate.Struct(req); err != nil {
		middleware.GetLogger(c).Error("friend request validation failed", "err", err)
		return Error(c, http.StatusBadRequest, "好友请求数据无效")
	}

	targetUser, err := h.userRepo.FindByEmail(req.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Error(c, http.StatusNotFound, "用户不存在")
		}
		middleware.GetLogger(c).Error("failed to look up user by email", "err", err)
		return Error(c, http.StatusInternalServerError, "查找用户失败")
	}

	result, err := h.friendSvc.SendRequest(userID, targetUser.ID)
	if err != nil {
		middleware.GetLogger(c).Error("failed to send friend request", "err", err)
		return Error(c, http.StatusInternalServerError, "发送好友请求失败")
	}

	return Created(c, result)
}

// AcceptRequest handles PUT /api/friends/request/:id/accept.
//
// It requires authentication. Only the recipient can accept a friend request.
func (h *FriendHandler) AcceptRequest(c echo.Context) error {
	userID, ok := c.Get(middleware.UserIDKey).(int64)
	if !ok {
		return Error(c, http.StatusUnauthorized, "未登录")
	}

	idStr := c.Param("id")
	requestID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		middleware.GetLogger(c).Error("failed to parse friend request id", "err", err)
		return Error(c, http.StatusBadRequest, "无效的请求ID")
	}

	if err := h.friendSvc.AcceptRequest(requestID, userID); err != nil {
		middleware.GetLogger(c).Error("failed to accept friend request", "err", err)
		if err.Error() == "service: only the recipient can accept a friend request" {
			return Error(c, http.StatusForbidden, "只有请求接收者才能接受好友请求")
		}
		if err.Error() == "service: friend request is not pending" {
			return Error(c, http.StatusBadRequest, "该好友请求已处理")
		}
		return Error(c, http.StatusInternalServerError, "接受好友请求失败")
	}

	return Success(c, nil)
}

// RejectRequest handles PUT /api/friends/request/:id/reject.
//
// It requires authentication. Only the recipient can reject a friend request.
func (h *FriendHandler) RejectRequest(c echo.Context) error {
	userID, ok := c.Get(middleware.UserIDKey).(int64)
	if !ok {
		return Error(c, http.StatusUnauthorized, "未登录")
	}

	idStr := c.Param("id")
	requestID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		middleware.GetLogger(c).Error("failed to parse friend request id", "err", err)
		return Error(c, http.StatusBadRequest, "无效的请求ID")
	}

	if err := h.friendSvc.RejectRequest(requestID, userID); err != nil {
		middleware.GetLogger(c).Error("failed to reject friend request", "err", err)
		if err.Error() == "service: only the recipient can reject a friend request" {
			return Error(c, http.StatusForbidden, "只有请求接收者才能拒绝好友请求")
		}
		if err.Error() == "service: friend request is not pending" {
			return Error(c, http.StatusBadRequest, "该好友请求已处理")
		}
		return Error(c, http.StatusInternalServerError, "拒绝好友请求失败")
	}

	return Success(c, nil)
}

// GetRequests handles GET /api/friends/requests.
//
// It requires authentication. Returns pending friend requests for the current user.
func (h *FriendHandler) GetRequests(c echo.Context) error {
	userID, ok := c.Get(middleware.UserIDKey).(int64)
	if !ok {
		return Error(c, http.StatusUnauthorized, "未登录")
	}

	requests, err := h.friendSvc.GetPendingRequests(userID)
	if err != nil {
		middleware.GetLogger(c).Error("failed to get pending requests", "err", err)
		return Error(c, http.StatusInternalServerError, "获取好友请求失败")
	}

	return Success(c, requests)
}

// GetFriends handles GET /api/friends.
//
// It requires authentication. Returns the friend list for the current user.
func (h *FriendHandler) GetFriends(c echo.Context) error {
	userID, ok := c.Get(middleware.UserIDKey).(int64)
	if !ok {
		return Error(c, http.StatusUnauthorized, "未登录")
	}

	friends, err := h.friendSvc.GetFriends(userID)
	if err != nil {
		middleware.GetLogger(c).Error("failed to get friends", "err", err)
		return Error(c, http.StatusInternalServerError, "获取好友列表失败")
	}

	return Success(c, friends)
}

// DeleteFriend handles DELETE /api/friends/:id.
//
// It requires authentication. Removes a friend from the current user's friend list.
func (h *FriendHandler) DeleteFriend(c echo.Context) error {
	userID, ok := c.Get(middleware.UserIDKey).(int64)
	if !ok {
		return Error(c, http.StatusUnauthorized, "未登录")
	}

	idStr := c.Param("id")
	friendID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		middleware.GetLogger(c).Error("failed to parse friend id", "err", err)
		return Error(c, http.StatusBadRequest, "无效的好友ID")
	}

	if err := h.friendSvc.RemoveFriend(userID, friendID); err != nil {
		middleware.GetLogger(c).Error("failed to remove friend", "err", err)
		if errors.Is(err, sql.ErrNoRows) || err.Error() == "service: friendship not found" {
			return Error(c, http.StatusNotFound, "好友关系不存在")
		}
		return Error(c, http.StatusInternalServerError, "删除好友失败")
	}

	return Success(c, nil)
}

// GetNotifications handles GET /api/notifications.
//
// It requires authentication. Returns all notifications for the current user.
func (h *FriendHandler) GetNotifications(c echo.Context) error {
	userID, ok := c.Get(middleware.UserIDKey).(int64)
	if !ok {
		return Error(c, http.StatusUnauthorized, "未登录")
	}

	notifications, err := h.notificationRepo.GetByUser(userID)
	if err != nil {
		middleware.GetLogger(c).Error("failed to get notifications", "err", err)
		return Error(c, http.StatusInternalServerError, "获取通知失败")
	}

	return Success(c, notifications)
}

// MarkNotificationRead handles PUT /api/notifications/:id/read.
//
// It requires authentication. Marks a notification as read.
func (h *FriendHandler) MarkNotificationRead(c echo.Context) error {
	userID, ok := c.Get(middleware.UserIDKey).(int64)
	if !ok {
		return Error(c, http.StatusUnauthorized, "未登录")
	}

	idStr := c.Param("id")
	notificationID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		middleware.GetLogger(c).Error("failed to parse notification id", "err", err)
		return Error(c, http.StatusBadRequest, "无效的通知ID")
	}

	if err := h.notificationRepo.MarkRead(notificationID, userID); err != nil {
		middleware.GetLogger(c).Error("failed to mark notification read", "err", err)
		if errors.Is(err, sql.ErrNoRows) {
			return Error(c, http.StatusNotFound, "通知不存在")
		}
		return Error(c, http.StatusInternalServerError, "标记通知失败")
	}

	return Success(c, nil)
}
