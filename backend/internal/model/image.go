package model

import "time"

// Image represents an uploaded image.
type Image struct {
	ID           int64     `json:"id" db:"id"`
	AlbumID      *int64    `json:"album_id" db:"album_id"`
	Title        string    `json:"title" db:"title"`
	Description  string    `json:"description" db:"description"`
	Tags         []string  `json:"tags" db:"tags"`
	LskyURL      string    `json:"lsky_url" db:"lsky_url"`
	ThumbnailURL string    `json:"thumbnail_url" db:"thumbnail_url"`
	UploadedBy   int64     `json:"uploaded_by" db:"uploaded_by"`
	CommentCount int64     `json:"comment_count" db:"-"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}
