package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"gm_site/internal/model"
)

// ErrAlbumHasImages is returned when attempting to delete an album that still
// has associated images.
var ErrAlbumHasImages = errors.New("album has associated images")

// AlbumRepository handles database operations for albums.
type AlbumRepository struct {
	db *sql.DB
}

// NewAlbumRepository creates a new AlbumRepository.
func NewAlbumRepository(db *sql.DB) *AlbumRepository {
	return &AlbumRepository{db: db}
}

// Create inserts a new album and populates the ID and CreatedAt fields.
func (r *AlbumRepository) Create(album *model.Album) error {
	now := time.Now()
	album.CreatedAt = now

	result, err := r.db.Exec(
		`INSERT INTO albums (name, description, created_by, created_at)
		 VALUES (?, ?, ?, ?)`,
		album.Name, album.Description, album.CreatedBy, album.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("repository: create album failed: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("repository: get last insert id failed: %w", err)
	}
	album.ID = id
	return nil
}

// FindAll retrieves all albums ordered by created_at descending (most recent first).
func (r *AlbumRepository) FindAll() ([]model.Album, error) {
	rows, err := r.db.Query(
		`SELECT id, name, description, created_by, created_at
		 FROM albums ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("repository: find all albums failed: %w", err)
	}
	defer rows.Close()

	var albums []model.Album
	for rows.Next() {
		var a model.Album
		if err := rows.Scan(&a.ID, &a.Name, &a.Description, &a.CreatedBy, &a.CreatedAt); err != nil {
			return nil, fmt.Errorf("repository: scan album failed: %w", err)
		}
		albums = append(albums, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: rows iteration failed: %w", err)
	}

	// Return empty slice (not nil) for consistency
	if albums == nil {
		albums = []model.Album{}
	}
	return albums, nil
}

// FindByID retrieves an album by its ID.
// Returns sql.ErrNoRows if not found.
func (r *AlbumRepository) FindByID(id int64) (*model.Album, error) {
	album := &model.Album{}
	err := r.db.QueryRow(
		`SELECT id, name, description, created_by, created_at
		 FROM albums WHERE id = ?`, id,
	).Scan(&album.ID, &album.Name, &album.Description, &album.CreatedBy, &album.CreatedAt)
	if err != nil {
		return nil, err
	}
	return album, nil
}

// Update modifies the name and description of an existing album.
// Returns sql.ErrNoRows if the album does not exist.
func (r *AlbumRepository) Update(album *model.Album) error {
	result, err := r.db.Exec(
		`UPDATE albums SET name = ?, description = ? WHERE id = ?`,
		album.Name, album.Description, album.ID,
	)
	if err != nil {
		return fmt.Errorf("repository: update album failed: %w", err)
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

// Delete removes an album by its ID. It first checks whether the album has
// any associated images. If it does, ErrAlbumHasImages is returned.
// Returns sql.ErrNoRows if the album does not exist.
func (r *AlbumRepository) Delete(id int64) error {
	// Check for associated images
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM images WHERE album_id = ?", id).Scan(&count)
	if err != nil {
		return fmt.Errorf("repository: check images for album failed: %w", err)
	}
	if count > 0 {
		return ErrAlbumHasImages
	}

	result, err := r.db.Exec("DELETE FROM albums WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("repository: delete album failed: %w", err)
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
