package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"gm_site/internal/model"
)

// ImageRepository handles database operations for images.
type ImageRepository struct {
	db *sql.DB
}

// NewImageRepository creates a new ImageRepository.
func NewImageRepository(db *sql.DB) *ImageRepository {
	return &ImageRepository{db: db}
}

// Create inserts a new image record. It populates the ID, CreatedAt, and UpdatedAt
// fields on success. The Tags slice is marshalled to a JSON array for storage.
func (r *ImageRepository) Create(image *model.Image) error {
	now := time.Now()
	image.CreatedAt = now
	image.UpdatedAt = now

	tagsJSON, err := json.Marshal(image.Tags)
	if err != nil {
		return fmt.Errorf("repository: marshal tags failed: %w", err)
	}

	result, err := r.db.Exec(
		`INSERT INTO images (album_id, title, description, tags, lsky_url, thumbnail_url, uploaded_by, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		image.AlbumID,
		image.Title,
		image.Description,
		string(tagsJSON),
		image.LskyURL,
		image.ThumbnailURL,
		image.UploadedBy,
		image.CreatedAt,
		image.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("repository: create image failed: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("repository: get last insert id failed: %w", err)
	}
	image.ID = id
	return nil
}

// FindAll retrieves paginated images with optional album/tag filters.
// Images are ordered by created_at DESC. Returns the matching images and the total count.
func (r *ImageRepository) FindAll(page, limit int, albumID *int64, tag string) ([]model.Image, int, error) {
	// Build WHERE clauses and args dynamically
	var conditions []string
	var args []interface{}

	if albumID != nil {
		conditions = append(conditions, "album_id = ?")
		args = append(args, *albumID)
	}

	if tag != "" {
		conditions = append(conditions, "tags LIKE ?")
		args = append(args, `%"`+tag+`"%`)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + conditions[0]
		for i := 1; i < len(conditions); i++ {
			whereClause += " AND " + conditions[i]
		}
	}

	// Count total matching rows
	var total int
	countQuery := "SELECT COUNT(*) FROM images " + whereClause
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("repository: count images failed: %w", err)
	}

	// Fetch paginated results
	offset := (page - 1) * limit
	selectQuery := `SELECT id, album_id, title, description, tags, lsky_url, thumbnail_url, uploaded_by, created_at, updated_at
		FROM images ` + whereClause + ` ORDER BY created_at DESC LIMIT ? OFFSET ?`
	queryArgs := append(args, limit, offset)

	rows, err := r.db.Query(selectQuery, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("repository: find all images failed: %w", err)
	}
	defer rows.Close()

	var images []model.Image
	for rows.Next() {
		var img model.Image
		var tagsJSON string
		if err := rows.Scan(
			&img.ID, &img.AlbumID, &img.Title, &img.Description, &tagsJSON,
			&img.LskyURL, &img.ThumbnailURL, &img.UploadedBy, &img.CreatedAt, &img.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("repository: scan image failed: %w", err)
		}
		if err := json.Unmarshal([]byte(tagsJSON), &img.Tags); err != nil {
			img.Tags = []string{}
		}
		images = append(images, img)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("repository: rows iteration failed: %w", err)
	}

	if images == nil {
		images = []model.Image{}
	}
	return images, total, nil
}

// FindByID retrieves an image by its ID, including its comment count.
// Returns sql.ErrNoRows if not found.
func (r *ImageRepository) FindByID(id int64) (*model.Image, error) {
	img := &model.Image{}
	var tagsJSON string
	err := r.db.QueryRow(
		`SELECT i.id, i.album_id, i.title, i.description, i.tags, i.lsky_url, i.thumbnail_url,
			i.uploaded_by, i.created_at, i.updated_at,
			COALESCE((SELECT COUNT(*) FROM comments WHERE image_id = i.id), 0) AS comment_count
		 FROM images i WHERE i.id = ?`, id,
	).Scan(
		&img.ID, &img.AlbumID, &img.Title, &img.Description, &tagsJSON,
		&img.LskyURL, &img.ThumbnailURL, &img.UploadedBy, &img.CreatedAt, &img.UpdatedAt,
		&img.CommentCount,
	)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(tagsJSON), &img.Tags); err != nil {
		img.Tags = []string{}
	}
	return img, nil
}

// Update modifies the title, description, tags, and album_id of an existing image.
// Returns sql.ErrNoRows if the image does not exist.
func (r *ImageRepository) Update(image *model.Image) error {
	tagsJSON, err := json.Marshal(image.Tags)
	if err != nil {
		return fmt.Errorf("repository: marshal tags failed: %w", err)
	}

	result, err := r.db.Exec(
		`UPDATE images SET album_id = ?, title = ?, description = ?, tags = ?, updated_at = ? WHERE id = ?`,
		image.AlbumID, image.Title, image.Description, string(tagsJSON), time.Now(), image.ID,
	)
	if err != nil {
		return fmt.Errorf("repository: update image failed: %w", err)
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

// Delete removes an image by its ID.
// Returns sql.ErrNoRows if the image does not exist.
func (r *ImageRepository) Delete(id int64) error {
	result, err := r.db.Exec("DELETE FROM images WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("repository: delete image failed: %w", err)
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

// SearchImages searches images by title, description, or tags using LIKE.
// Returns the matching images, total count, and any error.
// Results are ordered by created_at descending with pagination.
func (r *ImageRepository) SearchImages(query string, page, limit int) ([]model.Image, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 12
	}

	like := "%" + query + "%"
	offset := (page - 1) * limit

	// Count total matching images
	var total int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM images WHERE title LIKE ? OR description LIKE ? OR tags LIKE ?`,
		like, like, like,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("repository: count search images failed: %w", err)
	}

	// Fetch paginated results
	rows, err := r.db.Query(
		`SELECT id, album_id, title, description, tags, lsky_url, thumbnail_url, uploaded_by, created_at, updated_at
		 FROM images
		 WHERE title LIKE ? OR description LIKE ? OR tags LIKE ?
		 ORDER BY created_at DESC
		 LIMIT ? OFFSET ?`,
		like, like, like, limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("repository: search images failed: %w", err)
	}
	defer rows.Close()

	var images []model.Image
	for rows.Next() {
		var img model.Image
		var tagsJSON string
		err := rows.Scan(
			&img.ID, &img.AlbumID, &img.Title, &img.Description, &tagsJSON,
			&img.LskyURL, &img.ThumbnailURL, &img.UploadedBy,
			&img.CreatedAt, &img.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("repository: scan search image row failed: %w", err)
		}
		if err := json.Unmarshal([]byte(tagsJSON), &img.Tags); err != nil {
			img.Tags = []string{}
		}
		images = append(images, img)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("repository: iterate search image rows failed: %w", err)
	}

	if images == nil {
		images = []model.Image{}
	}

	return images, total, nil
}
