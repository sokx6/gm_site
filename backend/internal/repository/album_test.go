package repository

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"gm_site/internal/model"

	_ "modernc.org/sqlite"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

// setupAlbumTestDB creates a temporary SQLite database with users, albums, and
// images tables, and returns the db handle and a teardown function.
func setupAlbumTestDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	dsn := "file:" + dbPath + "?cache=shared&_journal_mode=WAL"

	db, err := sql.Open("sqlite", dsn)
	require.NoError(t, err, "failed to open test database")

	_, err = db.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			nickname TEXT NOT NULL DEFAULT '',
			role TEXT NOT NULL DEFAULT 'user',
			status TEXT NOT NULL DEFAULT 'pending',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE albums (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			created_by INTEGER NOT NULL REFERENCES users(id),
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE images (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			album_id INTEGER REFERENCES albums(id) ON DELETE SET NULL,
			title TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			tags TEXT NOT NULL DEFAULT '[]',
			lsky_url TEXT NOT NULL,
			thumbnail_url TEXT NOT NULL DEFAULT '',
			uploaded_by INTEGER NOT NULL REFERENCES users(id),
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`)
	require.NoError(t, err, "failed to create tables")

	teardown := func() {
		db.Close()
	}

	return db, teardown
}

// insertAlbumTestUser inserts a test user and returns the populated model.
func insertAlbumTestUser(t *testing.T, db *sql.DB) *model.User {
	t.Helper()

	now := time.Now()
	result, err := db.Exec(
		`INSERT INTO users (email, password_hash, nickname, role, status, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		"albumuser@test.com", "hash", "AlbumUser", "user", "approved", now, now,
	)
	require.NoError(t, err, "failed to insert test user")

	id, err := result.LastInsertId()
	require.NoError(t, err, "failed to get last insert id")

	return &model.User{
		ID:           id,
		Email:        "albumuser@test.com",
		PasswordHash: "hash",
		Nickname:     "AlbumUser",
		Role:         "user",
		Status:       "approved",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// insertTestAlbum inserts an album directly and returns the populated model.
func insertTestAlbum(t *testing.T, db *sql.DB, name, description string, createdBy int64) *model.Album {
	t.Helper()

	now := time.Now()
	result, err := db.Exec(
		`INSERT INTO albums (name, description, created_by, created_at)
		 VALUES (?, ?, ?, ?)`,
		name, description, createdBy, now,
	)
	require.NoError(t, err, "failed to insert test album")

	id, err := result.LastInsertId()
	require.NoError(t, err, "failed to get last insert id")

	return &model.Album{
		ID:          id,
		Name:        name,
		Description: description,
		CreatedBy:   createdBy,
		CreatedAt:   now,
	}
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestAlbumRepository_Create(t *testing.T) {
	db, teardown := setupAlbumTestDB(t)
	defer teardown()

	repo := NewAlbumRepository(db)
	user := insertAlbumTestUser(t, db)

	album := &model.Album{
		Name:        "My Album",
		Description: "A test album",
		CreatedBy:   user.ID,
	}

	err := repo.Create(album)
	require.NoError(t, err, "Create should not error")
	assert.Greater(t, album.ID, int64(0), "ID should be auto-incremented")
	assert.False(t, album.CreatedAt.IsZero(), "CreatedAt should be set")

	// Verify by querying directly
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM albums WHERE id = ?", album.ID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestAlbumRepository_FindAll(t *testing.T) {
	db, teardown := setupAlbumTestDB(t)
	defer teardown()

	repo := NewAlbumRepository(db)
	user := insertAlbumTestUser(t, db)

	a1 := insertTestAlbum(t, db, "Album 1", "First", user.ID)
	a2 := insertTestAlbum(t, db, "Album 2", "Second", user.ID)

	albums, err := repo.FindAll()
	require.NoError(t, err, "FindAll should not error")
	assert.Len(t, albums, 2, "should return 2 albums")

	// Should be ordered by created_at DESC (most recent first)
	assert.Equal(t, a2.ID, albums[0].ID, "first should be latest")
	assert.Equal(t, a1.ID, albums[1].ID, "second should be earliest")
}

func TestAlbumRepository_FindAll_Empty(t *testing.T) {
	db, teardown := setupAlbumTestDB(t)
	defer teardown()

	repo := NewAlbumRepository(db)

	albums, err := repo.FindAll()
	require.NoError(t, err, "FindAll should not error")
	assert.Empty(t, albums, "should return empty slice")
}

func TestAlbumRepository_FindByID(t *testing.T) {
	db, teardown := setupAlbumTestDB(t)
	defer teardown()

	repo := NewAlbumRepository(db)
	user := insertAlbumTestUser(t, db)
	inserted := insertTestAlbum(t, db, "Find Me", "Description", user.ID)

	found, err := repo.FindByID(inserted.ID)
	require.NoError(t, err, "FindByID should not error")
	assert.Equal(t, inserted.ID, found.ID)
	assert.Equal(t, inserted.Name, found.Name)
	assert.Equal(t, inserted.Description, found.Description)
	assert.Equal(t, inserted.CreatedBy, found.CreatedBy)
}

func TestAlbumRepository_FindByID_NotFound(t *testing.T) {
	db, teardown := setupAlbumTestDB(t)
	defer teardown()

	repo := NewAlbumRepository(db)

	_, err := repo.FindByID(99999)
	assert.ErrorIs(t, err, sql.ErrNoRows, "should return sql.ErrNoRows for non-existent ID")
}

func TestAlbumRepository_Update(t *testing.T) {
	db, teardown := setupAlbumTestDB(t)
	defer teardown()

	repo := NewAlbumRepository(db)
	user := insertAlbumTestUser(t, db)
	inserted := insertTestAlbum(t, db, "Original Name", "Original Desc", user.ID)

	inserted.Name = "Updated Name"
	inserted.Description = "Updated Description"

	err := repo.Update(inserted)
	require.NoError(t, err, "Update should not error")

	// Verify by fetching again
	updated, err := repo.FindByID(inserted.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", updated.Name)
	assert.Equal(t, "Updated Description", updated.Description)
	assert.Equal(t, user.ID, updated.CreatedBy) // created_by should not change
}

func TestAlbumRepository_Update_NotFound(t *testing.T) {
	db, teardown := setupAlbumTestDB(t)
	defer teardown()

	repo := NewAlbumRepository(db)

	album := &model.Album{
		ID:          99999,
		Name:        "Ghost",
		Description: "Does not exist",
	}

	err := repo.Update(album)
	assert.ErrorIs(t, err, sql.ErrNoRows, "should return sql.ErrNoRows for non-existent ID")
}

func TestAlbumRepository_Delete(t *testing.T) {
	db, teardown := setupAlbumTestDB(t)
	defer teardown()

	repo := NewAlbumRepository(db)
	user := insertAlbumTestUser(t, db)
	inserted := insertTestAlbum(t, db, "To Delete", "Will be deleted", user.ID)

	err := repo.Delete(inserted.ID)
	require.NoError(t, err, "Delete should not error")

	// Verify it's gone
	_, err = repo.FindByID(inserted.ID)
	assert.ErrorIs(t, err, sql.ErrNoRows, "album should be deleted")
}

func TestAlbumRepository_Delete_NotFound(t *testing.T) {
	db, teardown := setupAlbumTestDB(t)
	defer teardown()

	repo := NewAlbumRepository(db)

	err := repo.Delete(99999)
	assert.ErrorIs(t, err, sql.ErrNoRows, "should return sql.ErrNoRows for non-existent ID")
}

func TestAlbumRepository_Delete_HasImages(t *testing.T) {
	db, teardown := setupAlbumTestDB(t)
	defer teardown()

	repo := NewAlbumRepository(db)
	user := insertAlbumTestUser(t, db)
	album := insertTestAlbum(t, db, "Has Images", "Cannot delete", user.ID)

	// Insert an image associated with this album
	now := time.Now()
	_, err := db.Exec(
		`INSERT INTO images (album_id, title, description, tags, lsky_url, thumbnail_url, uploaded_by, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		album.ID, "Test Image", "", "[]", "https://example.com/img.jpg", "", user.ID, now, now,
	)
	require.NoError(t, err, "failed to insert test image")

	// Attempt to delete should fail
	err = repo.Delete(album.ID)
	assert.ErrorIs(t, err, ErrAlbumHasImages, "should return ErrAlbumHasImages when images exist")
}
