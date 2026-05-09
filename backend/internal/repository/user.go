package repository

import (
	"database/sql"
	"fmt"
	"time"

	"gm_site/internal/model"
)

// UserRepository handles database operations for users.
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create inserts a new user and populates the ID and timestamps.
func (r *UserRepository) Create(user *model.User) error {
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	result, err := r.db.Exec(
		`INSERT INTO users (email, password_hash, nickname, role, status, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		user.Email, user.PasswordHash, user.Nickname, user.Role, user.Status, user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("repository: create user failed: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("repository: get last insert id failed: %w", err)
	}
	user.ID = id
	return nil
}

// FindByEmail retrieves a user by email.
// Returns sql.ErrNoRows if not found.
func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	user := &model.User{}
	err := r.db.QueryRow(
		`SELECT id, email, password_hash, nickname, role, status, created_at, updated_at
		 FROM users WHERE email = ?`, email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Nickname, &user.Role, &user.Status, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// FindByID retrieves a user by ID.
// Returns sql.ErrNoRows if not found.
func (r *UserRepository) FindByID(id int64) (*model.User, error) {
	user := &model.User{}
	err := r.db.QueryRow(
		`SELECT id, email, password_hash, nickname, role, status, created_at, updated_at
		 FROM users WHERE id = ?`, id,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Nickname, &user.Role, &user.Status, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// UpdateStatus updates the user's status and updated_at timestamp.
func (r *UserRepository) UpdateStatus(id int64, status string) error {
	result, err := r.db.Exec(
		`UPDATE users SET status = ?, updated_at = ? WHERE id = ?`,
		string(status), time.Now(), id,
	)
	if err != nil {
		return fmt.Errorf("repository: update status failed: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("repository: rows affected failed: %w", err)
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// ListPending returns all users with pending status.
func (r *UserRepository) ListPending() ([]model.User, error) {
	rows, err := r.db.Query(
		`SELECT id, email, password_hash, nickname, role, status, created_at, updated_at
		 FROM users WHERE status = ? ORDER BY created_at ASC`, string(model.UserStatusPending),
	)
	if err != nil {
		return nil, fmt.Errorf("repository: list pending users failed: %w", err)
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Nickname, &u.Role, &u.Status, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, fmt.Errorf("repository: scan user failed: %w", err)
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: iterate users failed: %w", err)
	}

	// Return empty slice instead of nil for consistency
	if users == nil {
		users = []model.User{}
	}
	return users, nil
}

// CountByStatus returns the number of users with the given status.
func (r *UserRepository) CountByStatus(status string) (int, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM users WHERE status = ?`, status,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("repository: count by status failed: %w", err)
	}
	return count, nil
}

// UpdateUser updates a user's mutable fields (nickname, role, updated_at).
// This is a reserved method for future use (e.g., nickname changes).
func (r *UserRepository) UpdateUser(user *model.User) error {
	user.UpdatedAt = time.Now()
	result, err := r.db.Exec(
		`UPDATE users SET nickname = ?, role = ?, status = ?, updated_at = ? WHERE id = ?`,
		user.Nickname, user.Role, user.Status, user.UpdatedAt, user.ID,
	)
	if err != nil {
		return fmt.Errorf("repository: update user failed: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("repository: rows affected failed: %w", err)
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}
