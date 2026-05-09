package model

import "time"

// Comment represents a comment on an image.
type Comment struct {
	ID        int64     `json:"id" db:"id"`
	ImageID   int64     `json:"image_id" db:"image_id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	Content   string    `json:"content" db:"content"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
