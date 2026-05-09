package model

import "time"

// UserRole constants
const (
	UserRoleAdmin = "admin"
	UserRoleUser  = "user"
)

// UserStatus constants
const (
	UserStatusPending  = "pending"
	UserStatusApproved = "approved"
	UserStatusRejected = "rejected"
)

// User represents a user account.
type User struct {
	ID           int64     `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	Nickname     string    `json:"nickname" db:"nickname"`
	Role         string    `json:"role" db:"role"`
	Status       string    `json:"status" db:"status"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}
