package handler

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"

	"gm_site/internal/middleware"
	ws "gm_site/internal/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // dev mode — allow all origins
	},
}

// ServeWS returns an Echo handler that upgrades an HTTP connection to
// WebSocket and registers the client with the given hub.
//
// If the request has been processed by the OptionalAuth middleware, the
// authenticated user's ID is attached to the WebSocket client.
func ServeWS(hub *ws.Hub) echo.HandlerFunc {
	return func(c echo.Context) error {
		conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "websocket upgrade failed",
			})
		}

		userID := int64(0)
		if id, ok := c.Get(middleware.UserIDKey).(float64); ok {
			userID = int64(id)
		} else if id, ok := c.Get(middleware.UserIDKey).(int64); ok {
			userID = id
		}

		client := ws.NewClient(hub, conn, userID)
		hub.Register(client)

		go client.WritePump()
		go client.ReadPump()

		return nil
	}
}
