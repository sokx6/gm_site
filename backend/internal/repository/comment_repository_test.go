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

// setupCommentTestDB creates a temporary SQLite database with users, images, and comments tables.
func setupCommentTestDB(t *testing.T) (*sql.DB, func()) {
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
			role TEXT NOT NULL DEFAULT 'user' CHECK(role IN ('admin','user')),
			status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending','approved','rejected')),
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE images (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			album_id INTEGER,
			title TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			tags TEXT NOT NULL DEFAULT '[]',
			lsky_url TEXT NOT NULL,
			thumbnail_url TEXT NOT NULL DEFAULT '',
			uploaded_by INTEGER NOT NULL REFERENCES users(id),
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE comments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			image_id INTEGER NOT NULL REFERENCES images(id) ON DELETE CASCADE,
			user_id INTEGER NOT NULL REFERENCES users(id),
			content TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX idx_comments_image_id ON comments(image_id);
	`)
	require.NoError(t, err, "failed to create tables")

	teardown := func() {
		db.Close()
	}

	return db, teardown
}

// insertTestComment inserts a comment with given fields and returns the populated model.
func insertTestComment(t *testing.T, db *sql.DB, imageID, userID int64, content string) *model.Comment {
	t.Helper()

	now := time.Now()
	result, err := db.Exec(
		`INSERT INTO comments (image_id, user_id, content, created_at) VALUES (?, ?, ?, ?)`,
		imageID, userID, content, now,
	)
	require.NoError(t, err, "failed to insert test comment")

	id, err := result.LastInsertId()
	require.NoError(t, err, "failed to get last insert id")

	return &model.Comment{
		ID:        id,
		ImageID:   imageID,
		UserID:    userID,
		Content:   content,
		CreatedAt: now,
	}
}

// insertTestImage inserts an image with minimal fields.
func insertTestImage(t *testing.T, db *sql.DB, uploadedBy int64) *model.Image {
	t.Helper()

	now := time.Now()
	result, err := db.Exec(
		`INSERT INTO images (title, lsky_url, uploaded_by, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
		"test image", "https://example.com/img.jpg", uploadedBy, now, now,
	)
	require.NoError(t, err, "failed to insert test image")

	id, err := result.LastInsertId()
	require.NoError(t, err, "failed to get last insert id")

	return &model.Image{
		ID:         id,
		Title:      "test image",
		LskyURL:    "https://example.com/img.jpg",
		UploadedBy: uploadedBy,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestCommentRepository_Create(t *testing.T) {
	db, teardown := setupCommentTestDB(t)
	defer teardown()

	user := insertTestUser(t, db, "commenter@test.com", "hash", "Commenter", model.UserRoleUser, model.UserStatusApproved)
	image := insertTestImage(t, db, user.ID)
	repo := NewCommentRepository(db)

	comment := &model.Comment{
		ImageID: image.ID,
		UserID:  user.ID,
		Content: "这是一条测试评论",
	}

	err := repo.Create(comment)
	require.NoError(t, err, "Create should not error")
	assert.Greater(t, comment.ID, int64(0), "ID should be auto-incremented")
	assert.False(t, comment.CreatedAt.IsZero(), "CreatedAt should be set")

	// Verify by querying directly
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM comments WHERE id = ?", comment.ID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestCommentRepository_FindByImageID(t *testing.T) {
	db, teardown := setupCommentTestDB(t)
	defer teardown()

	user := insertTestUser(t, db, "user@test.com", "hash", "User", model.UserRoleUser, model.UserStatusApproved)
	image1 := insertTestImage(t, db, user.ID)
	image2 := insertTestImage(t, db, user.ID)
	repo := NewCommentRepository(db)

	// Insert 5 comments for image1, 2 for image2
	for i := 0; i < 5; i++ {
		insertTestComment(t, db, image1.ID, user.ID, "comment")
	}
	for i := 0; i < 2; i++ {
		insertTestComment(t, db, image2.ID, user.ID, "comment")
	}

	// Test pagination: page 1, limit 2
	comments, total, err := repo.FindByImageID(image1.ID, 1, 2)
	require.NoError(t, err)
	assert.Equal(t, 5, total, "total should be 5")
	assert.Len(t, comments, 2, "should return 2 comments on page 1")

	// Test pagination: page 3, limit 2
	comments, total, err = repo.FindByImageID(image1.ID, 3, 2)
	require.NoError(t, err)
	assert.Equal(t, 5, total, "total should be 5")
	assert.Len(t, comments, 1, "should return 1 comment on page 3")

	// Test image with no comments
	emptyImg := insertTestImage(t, db, user.ID)
	comments, total, err = repo.FindByImageID(emptyImg.ID, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, 0, total, "total should be 0")
	assert.Empty(t, comments, "should return empty slice")
}

func TestCommentRepository_FindByImageID_DefaultPagination(t *testing.T) {
	db, teardown := setupCommentTestDB(t)
	defer teardown()

	user := insertTestUser(t, db, "user2@test.com", "hash", "User", model.UserRoleUser, model.UserStatusApproved)
	image := insertTestImage(t, db, user.ID)
	repo := NewCommentRepository(db)

	// Insert 30 comments
	for i := 0; i < 30; i++ {
		insertTestComment(t, db, image.ID, user.ID, "comment")
	}

	// page=0, limit=0 should default to page=1, limit=20
	comments, total, err := repo.FindByImageID(image.ID, 0, 0)
	require.NoError(t, err)
	assert.Equal(t, 30, total)
	assert.Len(t, comments, 20, "default limit is 20")
}

func TestCommentRepository_Delete(t *testing.T) {
	db, teardown := setupCommentTestDB(t)
	defer teardown()

	user := insertTestUser(t, db, "deluser@test.com", "hash", "DelUser", model.UserRoleUser, model.UserStatusApproved)
	image := insertTestImage(t, db, user.ID)
	repo := NewCommentRepository(db)

	comment := insertTestComment(t, db, image.ID, user.ID, "to be deleted")

	// Delete the comment
	err := repo.Delete(comment.ID)
	require.NoError(t, err, "Delete should not error")

	// Verify deletion
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM comments WHERE id = ?", comment.ID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)

	// Delete non-existent comment should return sql.ErrNoRows
	err = repo.Delete(99999)
	assert.ErrorIs(t, err, sql.ErrNoRows, "should return sql.ErrNoRows for non-existent comment")
}

func TestCommentRepository_FindByID(t *testing.T) {
	db, teardown := setupCommentTestDB(t)
	defer teardown()

	user := insertTestUser(t, db, "fbid@test.com", "hash", "FBID", model.UserRoleUser, model.UserStatusApproved)
	image := insertTestImage(t, db, user.ID)
	repo := NewCommentRepository(db)

	inserted := insertTestComment(t, db, image.ID, user.ID, "find me")

	found, err := repo.FindByID(inserted.ID)
	require.NoError(t, err)
	assert.Equal(t, inserted.ID, found.ID)
	assert.Equal(t, inserted.ImageID, found.ImageID)
	assert.Equal(t, inserted.UserID, found.UserID)
	assert.Equal(t, inserted.Content, found.Content)
}

func TestCommentRepository_FindByID_NotFound(t *testing.T) {
	db, teardown := setupCommentTestDB(t)
	defer teardown()

	repo := NewCommentRepository(db)

	_, err := repo.FindByID(99999)
	assert.ErrorIs(t, err, sql.ErrNoRows, "should return sql.ErrNoRows for non-existent comment")
}
