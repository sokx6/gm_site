package service

import (
	"database/sql"
	"encoding/json"
	"log"
	"time"
)

// HubBroadcaster defines the interface for broadcasting stats to connected clients.
// *websocket.Hub satisfies this interface directly.
type HubBroadcaster interface {
	ClientCount() int
	Broadcast(message []byte)
}

// StatsMessage is the JSON structure broadcast to WebSocket clients.
type StatsMessage struct {
	Type string    `json:"type"`
	Data StatsData `json:"data"`
}

// StatsData contains the actual statistics values.
type StatsData struct {
	Online     int `json:"online"`
	Visitors   int `json:"visitors"`
	NewMembers int `json:"newMembers"`
	SafeDays   int `json:"safeDays"`
}

// StatsService handles online statistics tracking and periodic broadcasting.
type StatsService struct {
	hub       HubBroadcaster
	db        *sql.DB
	startDate time.Time
}

// NewStatsService creates a new StatsService with the given hub, database, and start date.
func NewStatsService(hub HubBroadcaster, db *sql.DB, startDate time.Time) *StatsService {
	return &StatsService{
		hub:       hub,
		db:        db,
		startDate: startDate,
	}
}

// IncrementVisitor records a new visitor visit in the database.
func (s *StatsService) IncrementVisitor() error {
	_, err := s.db.Exec("INSERT INTO visitors DEFAULT VALUES")
	return err
}

// GetOnlineCount returns the number of currently connected WebSocket clients.
func (s *StatsService) GetOnlineCount() int {
	return s.hub.ClientCount()
}

// GetVisitorCount returns the total number of visitors recorded.
func (s *StatsService) GetVisitorCount() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM visitors").Scan(&count)
	return count, err
}

// GetNewMembersCount returns the count of approved users registered in the last 7 days.
func (s *StatsService) GetNewMembersCount() (int, error) {
	var count int
	err := s.db.QueryRow(
		"SELECT COUNT(*) FROM users WHERE status = ? AND created_at > datetime('now', '-7 days')",
		"approved",
	).Scan(&count)
	return count, err
}

// GetSafeDays returns the number of full days since the service start date.
func (s *StatsService) GetSafeDays() int {
	return int(time.Since(s.startDate).Hours() / 24)
}

// BroadcastStats collects current statistics and broadcasts them to all connected clients.
// Message format: {"type":"stats","data":{"online":N,"visitors":N,"newMembers":N,"safeDays":N}}
func (s *StatsService) BroadcastStats() {
	online := s.GetOnlineCount()

	visitors, err := s.GetVisitorCount()
	if err != nil {
		log.Printf("[StatsService] GetVisitorCount error: %v", err)
		visitors = 0
	}

	newMembers, err := s.GetNewMembersCount()
	if err != nil {
		log.Printf("[StatsService] GetNewMembersCount error: %v", err)
		newMembers = 0
	}

	safeDays := s.GetSafeDays()

	msg := StatsMessage{
		Type: "stats",
		Data: StatsData{
			Online:     online,
			Visitors:   visitors,
			NewMembers: newMembers,
			SafeDays:   safeDays,
		},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("[StatsService] json.Marshal error: %v", err)
		return
	}

	s.hub.Broadcast(data)
}

// StartBroadcastLoop starts a background goroutine that broadcasts stats at the given interval.
func (s *StatsService) StartBroadcastLoop(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			s.BroadcastStats()
		}
	}()
}
