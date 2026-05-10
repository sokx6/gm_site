package service

import (
	"database/sql"
	"errors"
	"fmt"

	"gm_site/internal/logger"
	"gm_site/internal/model"
	"gm_site/internal/repository"
)

type FriendService struct {
	friendRepo       *repository.FriendRepository
	userRepo         *repository.UserRepository
	emailSvc         EmailService
	notificationRepo *repository.NotificationRepository
}

func NewFriendService(
	friendRepo *repository.FriendRepository,
	userRepo *repository.UserRepository,
	emailSvc EmailService,
	notificationRepo *repository.NotificationRepository,
) *FriendService {
	return &FriendService{
		friendRepo:       friendRepo,
		userRepo:         userRepo,
		emailSvc:         emailSvc,
		notificationRepo: notificationRepo,
	}
}

func (s *FriendService) SendRequest(fromUserID, toUserID int64) (*model.FriendRequest, error) {
	req, err := s.friendRepo.CreateFriendRequest(fromUserID, toUserID)
	if err != nil {
		return nil, fmt.Errorf("service: send request failed: %w", err)
	}

	go func() {
		toUser, err := s.userRepo.FindByID(toUserID)
		if err != nil {
			logger.L.Error("find to_user for friend request notification failed", "err", err, "toUserID", toUserID)
			return
		}
		fromUser, err := s.userRepo.FindByID(fromUserID)
		if err != nil {
			logger.L.Error("find from_user for friend request notification failed", "err", err, "fromUserID", fromUserID)
			return
		}
		if err := s.emailSvc.SendFriendRequestNotification(toUser.Email, fromUser.Nickname); err != nil {
			logger.L.Error("SendFriendRequestNotification failed", "err", err, "toEmail", toUser.Email, "fromNickname", fromUser.Nickname)
		}
	}()

	return req, nil
}

func (s *FriendService) AcceptRequest(requestID, userID int64) error {
	req, err := s.friendRepo.GetFriendRequestByID(requestID)
	if err != nil {
		return fmt.Errorf("service: get request failed: %w", err)
	}

	if req.ToUserID != userID {
		return errors.New("service: only the recipient can accept a friend request")
	}

	if req.Status != "pending" {
		return errors.New("service: friend request is not pending")
	}

	if err := s.friendRepo.UpdateFriendRequestStatus(requestID, "accepted"); err != nil {
		return fmt.Errorf("service: update request status failed: %w", err)
	}

	if _, err := s.friendRepo.CreateFriendship(req.FromUserID, req.ToUserID); err != nil {
		return fmt.Errorf("service: create friendship failed: %w", err)
	}

	notification := &model.Notification{
		UserID:    req.FromUserID,
		Type:      "friend_accepted",
		Title:     "好友请求已接受",
		Content:   fmt.Sprintf("用户 %s 接受了您的好友请求", req.ToNickname),
		RelatedID: &requestID,
		IsRead:    false,
	}
	if err := s.notificationRepo.Create(notification); err != nil {
		logger.L.Error("create acceptance notification failed", "err", err)
	}

	go func() {
		fromUser, err := s.userRepo.FindByID(req.FromUserID)
		if err != nil {
			logger.L.Error("find from_user for friend accepted notification failed", "err", err, "fromUserID", req.FromUserID)
			return
		}
		if err := s.emailSvc.SendFriendAcceptedNotification(fromUser.Email, req.ToNickname); err != nil {
			logger.L.Error("SendFriendAcceptedNotification failed", "err", err, "toEmail", fromUser.Email, "fromNickname", req.ToNickname)
		}
	}()

	return nil
}

func (s *FriendService) RejectRequest(requestID, userID int64) error {
	req, err := s.friendRepo.GetFriendRequestByID(requestID)
	if err != nil {
		return fmt.Errorf("service: get request failed: %w", err)
	}

	if req.ToUserID != userID {
		return errors.New("service: only the recipient can reject a friend request")
	}

	if req.Status != "pending" {
		return errors.New("service: friend request is not pending")
	}

	if err := s.friendRepo.UpdateFriendRequestStatus(requestID, "rejected"); err != nil {
		return fmt.Errorf("service: update request status failed: %w", err)
	}

	notification := &model.Notification{
		UserID:    req.FromUserID,
		Type:      "friend_rejected",
		Title:     "好友请求已拒绝",
		Content:   fmt.Sprintf("用户 %s 拒绝了您的好友请求", req.ToNickname),
		RelatedID: &requestID,
		IsRead:    false,
	}
	if err := s.notificationRepo.Create(notification); err != nil {
		logger.L.Error("create rejection notification failed", "err", err)
	}

	go func() {
		fromUser, err := s.userRepo.FindByID(req.FromUserID)
		if err != nil {
			logger.L.Error("find from_user for friend rejected notification failed", "err", err, "fromUserID", req.FromUserID)
			return
		}
		if err := s.emailSvc.SendFriendRejectedNotification(fromUser.Email, req.ToNickname); err != nil {
			logger.L.Error("SendFriendRejectedNotification failed", "err", err, "toEmail", fromUser.Email, "fromNickname", req.ToNickname)
		}
	}()

	return nil
}

func (s *FriendService) GetPendingRequests(userID int64) ([]model.FriendRequest, error) {
	return s.friendRepo.GetPendingRequestsForUser(userID)
}

func (s *FriendService) GetFriends(userID int64) ([]model.Friend, error) {
	return s.friendRepo.GetFriends(userID)
}

func (s *FriendService) RemoveFriend(userID, friendID int64) error {
	err := s.friendRepo.DeleteFriendship(userID, friendID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("service: friendship not found")
		}
		return fmt.Errorf("service: remove friend failed: %w", err)
	}
	return nil
}
