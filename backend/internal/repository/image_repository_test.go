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

// setupImageTestDB creates a temporary SQLite database with users and images
// tables, and returns the db handle and a teardown function.
func setupImageTestDB(t *testing.T) (*sql.DB, func()) {
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

// insertImageTestUser inserts a test user and returns the populated model.
func insertImageTestUser(t *testing.T, db *sql.DB) *model.User {
	t.Helper()

	now := time.Now()
	result, err := db.Exec(
		`INSERT INTO users (email, password_hash, nickname, role, status, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		"imageuser@test.com", "hash", "ImageUser", "user", "approved", now, now,
	)
	require.NoError(t, err, "failed to insert test user")

	id, err := result.LastInsertId()
	require.NoError(t, err, "failed to get last insert id")

	return &model.User{
		ID:           id,
		Email:        "imageuser@test.com",
		PasswordHash: "hash",
		Nickname:     "ImageUser",
		Role:         "user",
		Status:       "approved",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestImageRepo_Create(t *testing.T) {
	db, teardown := setupImageTestDB(t)
	defer teardown()

	repo := NewImageRepository(db)
	user := insertImageTestUser(t, db)

	image := &model.Image{
		Title:      "Test Image",
		Tags:       []string{"tag1", "tag2"},
		LskyURL:    "https://images.example.com/test.png",
		UploadedBy: user.ID,
	}

	err := repo.Create(image)
	require.NoError(t, err, "Create should not error")
	assert.Greater(t, image.ID, int64(0), "ID should be auto-incremented")
	assert.False(t, image.CreatedAt.IsZero(), "CreatedAt should be set")
	assert.False(t, image.UpdatedAt.IsZero(), "UpdatedAt should be set")

	// Verify by querying directly
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM images WHERE id = ?", image.ID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	// Verify tags are stored as JSON
	var tagsJSON string
	err = db.QueryRow("SELECT tags FROM images WHERE id = ?", image.ID).Scan(&tagsJSON)
	require.NoError(t, err)
	assert.JSONEq(t, `["tag1","tag2"]`, tagsJSON)
}
