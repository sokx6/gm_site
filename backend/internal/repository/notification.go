package repository

import (
	"database/sql"
	"fmt"
	"time"

	"gm_site/internal/model"
)

type NotificationRepository struct {
	db *sql.DB
}

func NewNotificationRepository(db *sql.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) Create(notification *model.Notification) error {
	now := time.Now()
	notification.CreatedAt = now

	result, err := r.db.Exec(
		`INSERT INTO notifications (user_id, type, title, content, related_id, image_id, is_read, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		notification.UserID, notification.Type, notification.Title, notification.Content,
		notification.RelatedID, notification.ImageID, notification.IsRead, notification.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("repository: create notification failed: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("repository: get last insert id failed: %w", err)
	}
	notification.ID = id
	return nil
}

func (r *NotificationRepository) GetByUser(userID int64) ([]model.Notification, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, type, title, content, related_id, image_id, is_read, created_at
		 FROM notifications WHERE user_id = ? ORDER BY created_at DESC`, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("repository: get notifications failed: %w", err)
	}
	defer rows.Close()

	var notifications []model.Notification
	for rows.Next() {
		var n model.Notification
		if err := rows.Scan(&n.ID, &n.UserID, &n.Type, &n.Title, &n.Content, &n.RelatedID, &n.ImageID, &n.IsRead, &n.CreatedAt); err != nil {
			return nil, fmt.Errorf("repository: scan notification failed: %w", err)
		}
		notifications = append(notifications, n)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: iterate notifications failed: %w", err)
	}

	if notifications == nil {
		notifications = []model.Notification{}
	}
	return notifications, nil
}

func (r *NotificationRepository) MarkRead(id, userID int64) error {
	result, err := r.db.Exec(
		`UPDATE notifications SET is_read = 1 WHERE id = ? AND user_id = ?`,
		id, userID,
	)
	if err != nil {
		return fmt.Errorf("repository: mark read failed: %w", err)
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

// FindTodayByImageAndType returns the first notification for the given image and type
// that was created today, or sql.ErrNoRows if none exists.
func (r *NotificationRepository) FindTodayByImageAndType(imageID int64, notifType string) (*model.Notification, error) {
	today := time.Now().Truncate(24 * time.Hour)

	var n model.Notification
	err := r.db.QueryRow(
		`SELECT id, user_id, type, title, content, related_id, image_id, is_read, created_at
		 FROM notifications
		 WHERE image_id = ? AND type = ? AND created_at > ?
		 ORDER BY created_at DESC LIMIT 1`,
		imageID, notifType, today,
	).Scan(&n.ID, &n.UserID, &n.Type, &n.Title, &n.Content, &n.RelatedID, &n.ImageID, &n.IsRead, &n.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &n, nil
}

// AppendContent appends newContent to the notification's content (with a newline separator)
// and marks it as unread.
func (r *NotificationRepository) AppendContent(id int64, newContent string) error {
	result, err := r.db.Exec(
		`UPDATE notifications SET content = content || '\n' || ?, is_read = 0 WHERE id = ?`,
		newContent, id,
	)
	if err != nil {
		return fmt.Errorf("repository: append content failed: %w", err)
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
