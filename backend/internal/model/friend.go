package model

import "time"

type FriendRequest struct {
	ID         int64     `json:"id"`
	FromUserID int64     `json:"from_user_id"`
	ToUserID   int64     `json:"to_user_id"`
	Status     string    `json:"status"` // pending/accepted/rejected
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	// joined fields
	FromNickname string `json:"from_nickname,omitempty"`
	FromEmail    string `json:"from_email,omitempty"`
	ToNickname   string `json:"to_nickname,omitempty"`
	ToEmail      string `json:"to_email,omitempty"`
}

type Friend struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	FriendID  int64     `json:"friend_id"`
	CreatedAt time.Time `json:"created_at"`
	// joined
	FriendNickname string `json:"friend_nickname,omitempty"`
	FriendEmail    string `json:"friend_email,omitempty"`
}
