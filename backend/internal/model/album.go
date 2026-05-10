package model

import "time"

// Album represents a photo album.
type Album struct {
	ID          int64     `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	CreatedBy     int64     `json:"created_by" db:"created_by"`
	Privacy       string    `json:"privacy" db:"privacy"`
	IsFriendAlbum bool      `json:"is_friend_album" db:"is_friend_album"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}
