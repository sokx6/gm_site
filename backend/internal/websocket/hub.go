// Package websocket provides WebSocket hub and client management.
package websocket

import (
	"sync"
)

// Hub maintains the set of active clients and broadcasts messages to them.
type Hub struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
	mu         sync.RWMutex
}

// NewHub creates and returns a new Hub.
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte),
	}
}

// Run starts the hub event loop. It should be run as a goroutine.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					// Client's send buffer is full — drop it
					h.mu.RUnlock()
					h.mu.Lock()
					delete(h.clients, client)
					close(client.send)
					h.mu.Unlock()
					h.mu.RLock()
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Register adds a client to the hub. Safe for concurrent use.
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister removes a client from the hub and closes its send channel.
// Safe for concurrent use.
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// Broadcast sends a message to all connected clients.
// Safe for concurrent use.
func (h *Hub) Broadcast(message []byte) {
	h.broadcast <- message
}

// ClientCount returns the number of currently connected clients.
// Safe for concurrent use.
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
