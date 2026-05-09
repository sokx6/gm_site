package model

// RegisterRequest is the payload for user registration.
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Nickname string `json:"nickname" validate:"required"`
}

// LoginRequest is the payload for user login.
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// CreateAlbumRequest is the payload for creating an album.
type CreateAlbumRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
}

// CreateImageRequest is the payload for uploading an image.
type CreateImageRequest struct {
	Title       string `json:"title" validate:"required"`
	Description string `json:"description"`
	AlbumID     *int64 `json:"album_id"`
	Tags        string `json:"tags"` // comma-separated tags string
}

// UpdateImageRequest is the payload for updating an image.
type UpdateImageRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Tags        string `json:"tags"` // comma-separated tags string
}

// CreateCommentRequest is the payload for creating a comment.
type CreateCommentRequest struct {
	Content string `json:"content" validate:"required,max=500"`
}
