package service

import (
	"database/sql"
	"encoding/json"
	"path/filepath"
	"sync"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// MockHub — implements HubBroadcaster for testing
// ---------------------------------------------------------------------------

// MockHub is a test double for HubBroadcaster.
type MockHub struct {
	mu       sync.Mutex
	clients  int
	messages [][]byte
}

// NewMockHub creates a new MockHub with zero clients and no messages.
func NewMockHub() *MockHub {
	return &MockHub{}
}

// ClientCount returns the simulated number of connected clients.
func (m *MockHub) ClientCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.clients
}

// Broadcast records the message for later inspection.
func (m *MockHub) Broadcast(message []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	msg := make([]byte, len(message))
	copy(msg, message)
	m.messages = append(m.messages, msg)
}

// SetClientCount sets the simulated client count.
func (m *MockHub) SetClientCount(n int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.clients = n
}

// LastMessage returns the most recently broadcast message, or nil.
func (m *MockHub) LastMessage() []byte {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.messages) == 0 {
		return nil
	}
	return m.messages[len(m.messages)-1]
}

// Messages returns all broadcast messages.
func (m *MockHub) Messages() [][]byte {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([][]byte, len(m.messages))
	for i, msg := range m.messages {
		result[i] = make([]byte, len(msg))
		copy(result[i], msg)
	}
	return result
}

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

// setupStatsTestDB creates a temporary SQLite database with visitors and users
// tables and returns the db handle and a teardown function.
func setupStatsTestDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "stats_test.db")
	dsn := "file:" + dbPath + "?cache=shared&_journal_mode=WAL"

	db, err := sql.Open("sqlite", dsn)
	require.NoError(t, err, "failed to open test database")

	_, err = db.Exec(`
		CREATE TABLE visitors (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			visited_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)
	require.NoError(t, err, "failed to create visitors table")

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
	`)
	require.NoError(t, err, "failed to create users table")

	teardown := func() {
		db.Close()
	}

	return db, teardown
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

// TestStatsService_GetOnlineCount verifies that GetOnlineCount returns the
// current client count from the hub.
func TestStatsService_GetOnlineCount(t *testing.T) {
	mockHub := NewMockHub()
	mockHub.SetClientCount(5)

	svc := NewStatsService(mockHub, nil, time.Now())

	assert.Equal(t, 5, svc.GetOnlineCount())

	// Change count and verify again
	mockHub.SetClientCount(0)
	assert.Equal(t, 0, svc.GetOnlineCount())

	mockHub.SetClientCount(42)
	assert.Equal(t, 42, svc.GetOnlineCount())
}

// TestStatsService_GetVisitorCount verifies visitor counting from the database.
func TestStatsService_GetVisitorCount(t *testing.T) {
	db, teardown := setupStatsTestDB(t)
	defer teardown()

	mockHub := NewMockHub()
	svc := NewStatsService(mockHub, db, time.Now())

	// No visitors yet
	count, err := svc.GetVisitorCount()
	require.NoError(t, err)
	assert.Equal(t, 0, count)

	// Insert a few visitors
	for i := 0; i < 3; i++ {
		_, err := db.Exec("INSERT INTO visitors DEFAULT VALUES")
		require.NoError(t, err)
	}

	count, err = svc.GetVisitorCount()
	require.NoError(t, err)
	assert.Equal(t, 3, count)
}

// TestStatsService_GetSafeDays verifies that GetSafeDays returns the correct
// number of days since the start date.
func TestStatsService_GetSafeDays(t *testing.T) {
	now := time.Date(2026, 5, 9, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		startDate time.Time
		expected  int
	}{
		{
			name:      "same day",
			startDate: now,
			expected:  0,
		},
		{
			name:      "one day ago",
			startDate: now.AddDate(0, 0, -1),
			expected:  1,
		},
		{
			name:      "one week ago",
			startDate: now.AddDate(0, 0, -7),
			expected:  7,
		},
		{
			name:      "one month ago",
			startDate: now.AddDate(0, -1, 0),
			expected:  30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockHub := NewMockHub()
			svc := NewStatsService(mockHub, nil, tt.startDate)
			days := svc.GetSafeDays()
			// Allow a tolerance of 1 day to account for test execution time
			assert.InDelta(t, float64(tt.expected), float64(days), 1.0,
				"expected ~%d days, got %d", tt.expected, days)
		})
	}
}

// TestStatsService_BroadcastStats verifies that BroadcastStats sends a properly
// formatted JSON message via the hub.
func TestStatsService_BroadcastStats(t *testing.T) {
	db, teardown := setupStatsTestDB(t)
	defer teardown()

	mockHub := NewMockHub()
	mockHub.SetClientCount(3)

	startDate := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)

	// Insert some visitors
	for i := 0; i < 10; i++ {
		_, err := db.Exec("INSERT INTO visitors DEFAULT VALUES")
		require.NoError(t, err)
	}

	// Insert an approved user from within the last 7 days
	_, err := db.Exec(`
		INSERT INTO users (email, password_hash, nickname, role, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, datetime('now', '-1 days'), datetime('now'))
	`, "user1@test.com", "hash1", "user1", "user", "approved")
	require.NoError(t, err)

	// Insert an older approved user (outside 7-day window)
	_, err = db.Exec(`
		INSERT INTO users (email, password_hash, nickname, role, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, datetime('now', '-14 days'), datetime('now'))
	`, "user2@test.com", "hash2", "user2", "user", "approved")
	require.NoError(t, err)

	// Insert a pending user (should not be counted)
	_, err = db.Exec(`
		INSERT INTO users (email, password_hash, nickname, role, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, datetime('now', '-1 days'), datetime('now'))
	`, "user3@test.com", "hash3", "user3", "user", "pending")
	require.NoError(t, err)

	svc := NewStatsService(mockHub, db, startDate)
	svc.BroadcastStats()

	// Verify the message
	raw := mockHub.LastMessage()
	require.NotNil(t, raw, "expected at least one broadcast message")

	var msg StatsMessage
	err = json.Unmarshal(raw, &msg)
	require.NoError(t, err, "message should be valid JSON")

	assert.Equal(t, "stats", msg.Type)
	assert.Equal(t, 3, msg.Data.Online, "online should be 3 (mock hub clients)")
	assert.Equal(t, 10, msg.Data.Visitors, "visitors should be 10")
	assert.Equal(t, 1, msg.Data.NewMembers, "newMembers should be 1 (only the recent approved user)")
	assert.Greater(t, msg.Data.SafeDays, 0, "safeDays should be positive")
}

// TestStatsService_BroadcastStats_Error verifies graceful handling when the
// database queries fail (e.g., missing tables).
func TestStatsService_BroadcastStats_Error(t *testing.T) {
	// Create a valid DB connection but without visitors/users tables
	dbPath := filepath.Join(t.TempDir(), "empty_test.db")
	dsn := "file:" + dbPath + "?cache=shared&_journal_mode=WAL"
	db, err := sql.Open("sqlite", dsn)
	require.NoError(t, err, "failed to open test database")
	defer db.Close()

	mockHub := NewMockHub()
	mockHub.SetClientCount(1)

	svc := NewStatsService(mockHub, db, time.Now())
	svc.BroadcastStats()

	raw := mockHub.LastMessage()
	require.NotNil(t, raw)

	var msg StatsMessage
	err = json.Unmarshal(raw, &msg)
	require.NoError(t, err)

	assert.Equal(t, "stats", msg.Type)
	assert.Equal(t, 1, msg.Data.Online)
	assert.Equal(t, 0, msg.Data.Visitors, "visitors should be 0 on DB error")
	assert.Equal(t, 0, msg.Data.NewMembers, "newMembers should be 0 on DB error")
}
