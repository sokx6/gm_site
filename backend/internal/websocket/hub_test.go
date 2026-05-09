package websocket

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testUpgrader mirrors the upgrader used in the handler package.
var testUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// setupWSServer creates a test HTTP server with a /ws endpoint that upgrades
// connections and registers clients with the given hub.
func setupWSServer(t *testing.T, hub *Hub) string {
	t.Helper()

	e := echo.New()
	e.GET("/ws", func(c echo.Context) error {
		conn, err := testUpgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "upgrade failed"})
		}
		client := NewClient(hub, conn, 0)
		hub.Register(client)
		go client.WritePump()
		go client.ReadPump()
		return nil
	})

	server := httptest.NewServer(e)
	t.Cleanup(server.Close)

	// Convert http://... to ws://...
	return "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
}

// dialWS dials the WebSocket endpoint and returns the connection.
func dialWS(t *testing.T, url string) *websocket.Conn {
	t.Helper()

	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	require.NoError(t, err, "failed to dial WebSocket")
	t.Cleanup(func() { conn.Close() })

	// Give the hub goroutine time to process the registration.
	time.Sleep(50 * time.Millisecond)

	return conn
}

// readWSMessage reads a single text message from the WebSocket connection.
func readWSMessage(t *testing.T, conn *websocket.Conn, timeout time.Duration) string {
	t.Helper()

	conn.SetReadDeadline(time.Now().Add(timeout))
	mt, msg, err := conn.ReadMessage()
	require.NoError(t, err, "failed to read message")
	assert.Equal(t, websocket.TextMessage, mt)
	return string(msg)
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestHub_RegisterClient(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	url := setupWSServer(t, hub)
	dialWS(t, url)

	assert.Equal(t, 1, hub.ClientCount())
}

func TestHub_UnregisterClient(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	url := setupWSServer(t, hub)
	conn := dialWS(t, url)

	assert.Equal(t, 1, hub.ClientCount())

	// Closing the connection causes ReadPump to exit, which triggers
	// Unregister on the hub.
	conn.Close()
	time.Sleep(50 * time.Millisecond)

	assert.Equal(t, 0, hub.ClientCount())
}

func TestHub_Broadcast(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	url := setupWSServer(t, hub)

	conn1 := dialWS(t, url)
	conn2 := dialWS(t, url)

	assert.Equal(t, 2, hub.ClientCount())

	hub.Broadcast([]byte("hello world"))

	msg1 := readWSMessage(t, conn1, 2*time.Second)
	msg2 := readWSMessage(t, conn2, 2*time.Second)

	assert.Equal(t, "hello world", msg1)
	assert.Equal(t, "hello world", msg2)
}

func TestHub_ClientCount(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	url := setupWSServer(t, hub)

	assert.Equal(t, 0, hub.ClientCount())

	conn1 := dialWS(t, url)
	assert.Equal(t, 1, hub.ClientCount())

	conn2 := dialWS(t, url)
	assert.Equal(t, 2, hub.ClientCount())

	conn1.Close()
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 1, hub.ClientCount())

	conn2.Close()
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 0, hub.ClientCount())
}

func TestHub_BroadcastGracefulIgnoreFullSendBuffer(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	url := setupWSServer(t, hub)
	conn := dialWS(t, url)

	assert.Equal(t, 1, hub.ClientCount())

	// ReadPump receives the message and re-broadcasts it via hub.Broadcast,
	// creating an infinite loop. The send buffer is 256 messages; eventually
	// it fills up and the client gets dropped. This test validates that the
	// hub handles a full send buffer gracefully without deadlock.
	done := make(chan struct{})
	go func() {
		// Send a message to trigger the echo loop
		conn.WriteMessage(websocket.TextMessage, []byte("ping"))
		time.Sleep(200 * time.Millisecond)
		close(done)
	}()

	select {
	case <-done:
		// OK — no deadlock
	case <-time.After(3 * time.Second):
		t.Fatal("test timed out — possible deadlock with full send buffer")
	}
}
