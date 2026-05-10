package repository

import (
	"database/sql"
	"fmt"
	"time"

	"gm_site/internal/model"
)

// CommentRepository handles database operations for comments.
type CommentRepository struct {
	db *sql.DB
}

// NewCommentRepository creates a new CommentRepository.
func NewCommentRepository(db *sql.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

// Create inserts a new comment and populates the ID and CreatedAt.
func (r *CommentRepository) Create(comment *model.Comment) error {
	now := time.Now()
	comment.CreatedAt = now

	result, err := r.db.Exec(
		`INSERT INTO comments (image_id, user_id, content, parent_id, created_at)
		 VALUES (?, ?, ?, ?, ?)`,
		comment.ImageID, comment.UserID, comment.Content, comment.ParentID, comment.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("repository: create comment failed: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("repository: get last insert id failed: %w", err)
	}
	comment.ID = id
	return nil
}

// FindByImageID retrieves comments for a given image with pagination.
// Returns the list of comments and the total count matching the image.
func (r *CommentRepository) FindByImageID(imageID int64, page, limit int) ([]model.Comment, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	offset := (page - 1) * limit

	// Get total count of top-level comments
	var total int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM comments WHERE image_id = ? AND parent_id IS NULL`, imageID,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("repository: count comments failed: %w", err)
	}

	// Get paginated list of top-level comments with reply count
	rows, err := r.db.Query(
		`SELECT id, image_id, user_id, parent_id, content, created_at,
		        (SELECT COUNT(*) FROM comments AS replies WHERE replies.parent_id = c.id) AS reply_count
		 FROM comments AS c WHERE image_id = ? AND parent_id IS NULL
		 ORDER BY created_at ASC
		 LIMIT ? OFFSET ?`,
		imageID, limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("repository: find comments by image_id failed: %w", err)
	}
	defer rows.Close()

	var comments []model.Comment
	for rows.Next() {
		var c model.Comment
		if err := rows.Scan(&c.ID, &c.ImageID, &c.UserID, &c.ParentID, &c.Content, &c.CreatedAt, &c.ReplyCount); err != nil {
			return nil, 0, fmt.Errorf("repository: scan comment failed: %w", err)
		}
		comments = append(comments, c)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("repository: iterate comments failed: %w", err)
	}

	if comments == nil {
		comments = []model.Comment{}
	}

	return comments, total, nil
}

// FindReplies retrieves child comments for a given parent comment with pagination.
// Returns the list of replies and the total count.
func (r *CommentRepository) FindReplies(parentID int64, page, limit int) ([]model.Comment, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	offset := (page - 1) * limit

	var total int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM comments WHERE parent_id = ?`, parentID,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("repository: count replies failed: %w", err)
	}

	rows, err := r.db.Query(
		`SELECT id, image_id, user_id, parent_id, content, created_at,
		        (SELECT COUNT(*) FROM comments AS sub WHERE sub.parent_id = c.id) AS reply_count
		 FROM comments AS c WHERE parent_id = ?
		 ORDER BY created_at ASC
		 LIMIT ? OFFSET ?`,
		parentID, limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("repository: find replies failed: %w", err)
	}
	defer rows.Close()

	var replies []model.Comment
	for rows.Next() {
		var c model.Comment
		if err := rows.Scan(&c.ID, &c.ImageID, &c.UserID, &c.ParentID, &c.Content, &c.CreatedAt, &c.ReplyCount); err != nil {
			return nil, 0, fmt.Errorf("repository: scan reply failed: %w", err)
		}
		replies = append(replies, c)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("repository: iterate replies failed: %w", err)
	}

	if replies == nil {
		replies = []model.Comment{}
	}

	return replies, total, nil
}

// Delete removes a comment by ID. Returns sql.ErrNoRows if not found.
func (r *CommentRepository) Delete(id int64) error {
	result, err := r.db.Exec(
		`DELETE FROM comments WHERE id = ?`, id,
	)
	if err != nil {
		return fmt.Errorf("repository: delete comment failed: %w", err)
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

// FindByID retrieves a comment by ID. Returns sql.ErrNoRows if not found.
func (r *CommentRepository) FindByID(id int64) (*model.Comment, error) {
	c := &model.Comment{}
	err := r.db.QueryRow(
		`SELECT id, image_id, user_id, parent_id, content, created_at
		 FROM comments WHERE id = ?`, id,
	).Scan(&c.ID, &c.ImageID, &c.UserID, &c.ParentID, &c.Content, &c.CreatedAt)
	if err != nil {
		return nil, err
	}
	return c, nil
}
