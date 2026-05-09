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

// setupTestDB creates a temporary SQLite database with the users table
// and returns the db handle and a teardown function.
func setupTestDB(t *testing.T) (*sql.DB, func()) {
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
		CREATE INDEX idx_users_email ON users(email);
		CREATE INDEX idx_users_status ON users(status);
	`)
	require.NoError(t, err, "failed to create users table")

	teardown := func() {
		db.Close()
	}

	return db, teardown
}

// insertTestUser inserts a user with the given fields and returns the populated model.
func insertTestUser(t *testing.T, db *sql.DB, email, passwordHash, nickname, role, status string) *model.User {
	t.Helper()

	now := time.Now()
	result, err := db.Exec(
		`INSERT INTO users (email, password_hash, nickname, role, status, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		email, passwordHash, nickname, role, status, now, now,
	)
	require.NoError(t, err, "failed to insert test user")

	id, err := result.LastInsertId()
	require.NoError(t, err, "failed to get last insert id")

	return &model.User{
		ID:           id,
		Email:        email,
		PasswordHash: passwordHash,
		Nickname:     nickname,
		Role:         role,
		Status:       status,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestUserRepository_Create(t *testing.T) {
	db, teardown := setupTestDB(t)
	defer teardown()

	repo := NewUserRepository(db)

	user := &model.User{
		Email:        "alice@test.com",
		PasswordHash: "hashed_pwd_123",
		Nickname:     "Alice",
		Role:         model.UserRoleUser,
		Status:       model.UserStatusPending,
	}

	err := repo.Create(user)
	require.NoError(t, err, "Create should not error")
	assert.Greater(t, user.ID, int64(0), "ID should be auto-incremented")
	assert.False(t, user.CreatedAt.IsZero(), "CreatedAt should be set")
	assert.False(t, user.UpdatedAt.IsZero(), "UpdatedAt should be set")

	// Verify by querying directly
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE id = ?", user.ID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestUserRepository_FindByEmail(t *testing.T) {
	db, teardown := setupTestDB(t)
	defer teardown()

	repo := NewUserRepository(db)
	inserted := insertTestUser(t, db, "bob@test.com", "hash_bob", "Bob", model.UserRoleUser, model.UserStatusPending)

	found, err := repo.FindByEmail("bob@test.com")
	require.NoError(t, err, "FindByEmail should not error")
	assert.Equal(t, inserted.ID, found.ID)
	assert.Equal(t, inserted.Email, found.Email)
	assert.Equal(t, inserted.PasswordHash, found.PasswordHash)
	assert.Equal(t, inserted.Nickname, found.Nickname)
	assert.Equal(t, inserted.Role, found.Role)
	assert.Equal(t, inserted.Status, found.Status)
}

func TestUserRepository_FindByEmail_NotFound(t *testing.T) {
	db, teardown := setupTestDB(t)
	defer teardown()

	repo := NewUserRepository(db)

	_, err := repo.FindByEmail("nonexistent@test.com")
	assert.ErrorIs(t, err, sql.ErrNoRows, "should return sql.ErrNoRows for non-existent email")
}

func TestUserRepository_FindByID(t *testing.T) {
	db, teardown := setupTestDB(t)
	defer teardown()

	repo := NewUserRepository(db)
	inserted := insertTestUser(t, db, "carol@test.com", "hash_carol", "Carol", model.UserRoleAdmin, model.UserStatusApproved)

	found, err := repo.FindByID(inserted.ID)
	require.NoError(t, err, "FindByID should not error")
	assert.Equal(t, inserted.ID, found.ID)
	assert.Equal(t, inserted.Email, found.Email)
	assert.Equal(t, inserted.PasswordHash, found.PasswordHash)
	assert.Equal(t, inserted.Nickname, found.Nickname)
	assert.Equal(t, inserted.Role, found.Role)
	assert.Equal(t, inserted.Status, found.Status)
}

func TestUserRepository_FindByID_NotFound(t *testing.T) {
	db, teardown := setupTestDB(t)
	defer teardown()

	repo := NewUserRepository(db)

	_, err := repo.FindByID(99999)
	assert.ErrorIs(t, err, sql.ErrNoRows, "should return sql.ErrNoRows for non-existent ID")
}

func TestUserRepository_UpdateStatus(t *testing.T) {
	db, teardown := setupTestDB(t)
	defer teardown()

	repo := NewUserRepository(db)
	inserted := insertTestUser(t, db, "dave@test.com", "hash_dave", "Dave", model.UserRoleUser, model.UserStatusPending)

	// Update to approved
	err := repo.UpdateStatus(inserted.ID, model.UserStatusApproved)
	require.NoError(t, err, "UpdateStatus should not error")

	// Verify
	updated, err := repo.FindByID(inserted.ID)
	require.NoError(t, err)
	assert.Equal(t, string(model.UserStatusApproved), updated.Status, "status should be updated")
	assert.True(t, updated.UpdatedAt.After(inserted.UpdatedAt) || updated.UpdatedAt.Equal(inserted.UpdatedAt),
		"UpdatedAt should be refreshed")
}

func TestUserRepository_UpdateStatus_NotFound(t *testing.T) {
	db, teardown := setupTestDB(t)
	defer teardown()

	repo := NewUserRepository(db)

	err := repo.UpdateStatus(99999, model.UserStatusApproved)
	assert.ErrorIs(t, err, sql.ErrNoRows, "should return sql.ErrNoRows for non-existent ID")
}

func TestUserRepository_ListPending(t *testing.T) {
	db, teardown := setupTestDB(t)
	defer teardown()

	repo := NewUserRepository(db)

	// Insert users with different statuses
	insertTestUser(t, db, "eve@test.com", "hash_eve", "Eve", model.UserRoleUser, model.UserStatusPending)
	insertTestUser(t, db, "frank@test.com", "hash_frank", "Frank", model.UserRoleUser, model.UserStatusApproved)
	insertTestUser(t, db, "grace@test.com", "hash_grace", "Grace", model.UserRoleUser, model.UserStatusPending)
	insertTestUser(t, db, "heidi@test.com", "hash_heidi", "Heidi", model.UserRoleUser, model.UserStatusRejected)

	pendingUsers, err := repo.ListPending()
	require.NoError(t, err, "ListPending should not error")
	assert.Len(t, pendingUsers, 2, "should return exactly 2 pending users")

	for _, u := range pendingUsers {
		assert.Equal(t, string(model.UserStatusPending), u.Status,
			"all returned users should have pending status")
	}
}

func TestUserRepository_ListPending_Empty(t *testing.T) {
	db, teardown := setupTestDB(t)
	defer teardown()

	repo := NewUserRepository(db)

	// No users at all
	users, err := repo.ListPending()
	require.NoError(t, err, "ListPending should not error on empty table")
	assert.Empty(t, users, "should return empty slice when no pending users")
}

func TestUserRepository_CountByStatus(t *testing.T) {
	db, teardown := setupTestDB(t)
	defer teardown()

	repo := NewUserRepository(db)

	// Insert users with different statuses
	insertTestUser(t, db, "ivan@test.com", "hash_ivan", "Ivan", model.UserRoleUser, model.UserStatusPending)
	insertTestUser(t, db, "judy@test.com", "hash_judy", "Judy", model.UserRoleUser, model.UserStatusApproved)
	insertTestUser(t, db, "karl@test.com", "hash_karl", "Karl", model.UserRoleUser, model.UserStatusApproved)
	insertTestUser(t, db, "leo@test.com", "hash_leo", "Leo", model.UserRoleUser, model.UserStatusPending)

	pendingCount, err := repo.CountByStatus(model.UserStatusPending)
	require.NoError(t, err)
	assert.Equal(t, 2, pendingCount, "should count 2 pending users")

	approvedCount, err := repo.CountByStatus(model.UserStatusApproved)
	require.NoError(t, err)
	assert.Equal(t, 2, approvedCount, "should count 2 approved users")

	rejectedCount, err := repo.CountByStatus(model.UserStatusRejected)
	require.NoError(t, err)
	assert.Equal(t, 0, rejectedCount, "should count 0 rejected users")
}

func TestUserRepository_UpdateUser(t *testing.T) {
	db, teardown := setupTestDB(t)
	defer teardown()

	repo := NewUserRepository(db)
	inserted := insertTestUser(t, db, "mallory@test.com", "hash_mallory", "Mallory", model.UserRoleUser, model.UserStatusPending)

	// Update nickname and role
	inserted.Nickname = "MalloryUpdated"
	inserted.Role = model.UserRoleAdmin
	inserted.Status = model.UserStatusApproved

	err := repo.UpdateUser(inserted)
	require.NoError(t, err, "UpdateUser should not error")

	// Verify
	updated, err := repo.FindByID(inserted.ID)
	require.NoError(t, err)
	assert.Equal(t, "MalloryUpdated", updated.Nickname)
	assert.Equal(t, string(model.UserRoleAdmin), updated.Role)
	assert.Equal(t, string(model.UserStatusApproved), updated.Status)
	assert.True(t, updated.UpdatedAt.After(inserted.UpdatedAt) || updated.UpdatedAt.Equal(inserted.UpdatedAt),
		"UpdatedAt should be refreshed")
}
