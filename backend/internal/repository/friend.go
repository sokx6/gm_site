package repository

import (
	"database/sql"
	"fmt"
	"time"

	"gm_site/internal/model"
)

type FriendRepository struct {
	db *sql.DB
}

func NewFriendRepository(db *sql.DB) *FriendRepository {
	return &FriendRepository{db: db}
}

func (r *FriendRepository) CreateFriendRequest(fromUserID, toUserID int64) (*model.FriendRequest, error) {
	now := time.Now()
	req := &model.FriendRequest{
		FromUserID: fromUserID,
		ToUserID:   toUserID,
		Status:     "pending",
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	result, err := r.db.Exec(
		`INSERT INTO friend_requests (from_user_id, to_user_id, status, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?)`,
		req.FromUserID, req.ToUserID, req.Status, req.CreatedAt, req.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("repository: create friend request failed: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("repository: get last insert id failed: %w", err)
	}
	req.ID = id
	return req, nil
}

func (r *FriendRepository) GetFriendRequestByID(id int64) (*model.FriendRequest, error) {
	req := &model.FriendRequest{}
	err := r.db.QueryRow(
		`SELECT fr.id, fr.from_user_id, fr.to_user_id, fr.status, fr.created_at, fr.updated_at,
		        fu.nickname, fu.email, tu.nickname, tu.email
		 FROM friend_requests fr
		 JOIN users fu ON fu.id = fr.from_user_id
		 JOIN users tu ON tu.id = fr.to_user_id
		 WHERE fr.id = ?`, id,
	).Scan(&req.ID, &req.FromUserID, &req.ToUserID, &req.Status, &req.CreatedAt, &req.UpdatedAt,
		&req.FromNickname, &req.FromEmail, &req.ToNickname, &req.ToEmail)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func (r *FriendRepository) GetPendingRequestsForUser(userID int64) ([]model.FriendRequest, error) {
	rows, err := r.db.Query(
		`SELECT fr.id, fr.from_user_id, fr.to_user_id, fr.status, fr.created_at, fr.updated_at,
		        u.nickname, u.email
		 FROM friend_requests fr
		 JOIN users u ON u.id = fr.from_user_id
		 WHERE fr.to_user_id = ? AND fr.status = 'pending'
		 ORDER BY fr.created_at DESC`, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("repository: get pending requests failed: %w", err)
	}
	defer rows.Close()

	var requests []model.FriendRequest
	for rows.Next() {
		var req model.FriendRequest
		if err := rows.Scan(&req.ID, &req.FromUserID, &req.ToUserID, &req.Status, &req.CreatedAt, &req.UpdatedAt,
			&req.FromNickname, &req.FromEmail); err != nil {
			return nil, fmt.Errorf("repository: scan friend request failed: %w", err)
		}
		requests = append(requests, req)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: iterate friend requests failed: %w", err)
	}

	if requests == nil {
		requests = []model.FriendRequest{}
	}
	return requests, nil
}

func (r *FriendRepository) UpdateFriendRequestStatus(id int64, status string) error {
	result, err := r.db.Exec(
		`UPDATE friend_requests SET status = ?, updated_at = ? WHERE id = ?`,
		status, time.Now(), id,
	)
	if err != nil {
		return fmt.Errorf("repository: update friend request status failed: %w", err)
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

func (r *FriendRepository) CreateFriendship(userID, friendID int64) (*model.Friend, error) {
	now := time.Now()
	friend := &model.Friend{
		UserID:    userID,
		FriendID:  friendID,
		CreatedAt: now,
	}

	tx, err := r.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("repository: begin tx failed: %w", err)
	}
	defer tx.Rollback()

	result, err := tx.Exec(
		`INSERT INTO friends (user_id, friend_id, created_at) VALUES (?, ?, ?)`,
		userID, friendID, now,
	)
	if err != nil {
		return nil, fmt.Errorf("repository: create friendship (forward) failed: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("repository: get last insert id failed: %w", err)
	}

	_, err = tx.Exec(
		`INSERT INTO friends (user_id, friend_id, created_at) VALUES (?, ?, ?)`,
		friendID, userID, now,
	)
	if err != nil {
		return nil, fmt.Errorf("repository: create friendship (reverse) failed: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("repository: commit tx failed: %w", err)
	}

	friend.ID = id
	return friend, nil
}

func (r *FriendRepository) DeleteFriendship(userID, friendID int64) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("repository: begin tx failed: %w", err)
	}
	defer tx.Rollback()

	result, err := tx.Exec(
		`DELETE FROM friends WHERE user_id = ? AND friend_id = ?`,
		userID, friendID,
	)
	if err != nil {
		return fmt.Errorf("repository: delete friendship (forward) failed: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("repository: rows affected failed: %w", err)
	}
	if rows == 0 {
		return sql.ErrNoRows
	}

	_, err = tx.Exec(
		`DELETE FROM friends WHERE user_id = ? AND friend_id = ?`,
		friendID, userID,
	)
	if err != nil {
		return fmt.Errorf("repository: delete friendship (reverse) failed: %w", err)
	}

	return tx.Commit()
}

func (r *FriendRepository) GetFriends(userID int64) ([]model.Friend, error) {
	rows, err := r.db.Query(
		`SELECT f.id, f.user_id, f.friend_id, f.created_at, u.nickname, u.email
		 FROM friends f
		 JOIN users u ON u.id = f.friend_id
		 WHERE f.user_id = ?
		 ORDER BY f.created_at DESC`, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("repository: get friends failed: %w", err)
	}
	defer rows.Close()

	var friends []model.Friend
	for rows.Next() {
		var f model.Friend
		if err := rows.Scan(&f.ID, &f.UserID, &f.FriendID, &f.CreatedAt, &f.FriendNickname, &f.FriendEmail); err != nil {
			return nil, fmt.Errorf("repository: scan friend failed: %w", err)
		}
		friends = append(friends, f)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: iterate friends failed: %w", err)
	}

	if friends == nil {
		friends = []model.Friend{}
	}
	return friends, nil
}

func (r *FriendRepository) AreFriends(userID1, userID2 int64) (bool, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM friends WHERE user_id = ? AND friend_id = ?`,
		userID1, userID2,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("repository: check friendship failed: %w", err)
	}
	return count > 0, nil
}
