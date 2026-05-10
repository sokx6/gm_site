package model

import "time"

type Notification struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Type      string    `json:"type"` // friend_request/friend_accepted/friend_rejected/register_approved/register_rejected/comment_reply/image_comment
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	RelatedID *int64    `json:"related_id"`
	ImageID   *int64    `json:"image_id"`
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}
