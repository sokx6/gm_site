package model

// APIResponse is the standard JSON response envelope.
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// TokenPair contains access and refresh tokens.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// ImageDetail extends Image with additional computed fields for API responses.
type ImageDetail struct {
	Image
	CommentCount int `json:"comment_count"`
}

// Claims represents the JWT claims payload.
type Claims struct {
	UserID   int64  `json:"user_id"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	IssuedAt int64  `json:"iat"`
	Issuer   string `json:"iss"`
}

// RefreshRequest is the payload for refreshing tokens.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}
