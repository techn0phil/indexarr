package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"indexarr/internal/repository"
	"indexarr/internal/services"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from localhost (development and production)
		origin := r.Header.Get("Origin")
		return origin == "http://localhost:3000" || // Vite dev server
			origin == "http://localhost:5173" || // Vite dev server (alt port)
			origin == "http://localhost:8787" || // Production nginx port
			origin == "" // Allow non-browser clients
	},
}

// HandleWebSocket upgrades HTTP connection to WebSocket and manages client lifecycle
func HandleWebSocket(db *sql.DB, broadcaster *services.Broadcaster) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("Failed to upgrade to WebSocket: %v", err)
			return
		}

		client := services.NewClient(conn)

		// Register client with broadcaster
		broadcaster.Register(client)

		// Send initial scan status to the client
		go func() {
			status, err := repository.GetScanStatus(db)
			if err != nil {
				log.Printf("Failed to get initial scan status: %v", err)
			} else {
				// Send current status as initial message
				var msgType string
				switch status.Status {
				case "running":
					msgType = "scan_progress"
				case "completed":
					msgType = "scan_complete"
				case "error":
					msgType = "scan_error"
				case "stopped":
					msgType = "scan_stopped"
				default:
					msgType = "scan_idle"
				}

				msg := services.WSMessage{
					Type:           msgType,
					FilesFound:     status.FilesFound,
					FilesProcessed: status.FilesProcessed,
					StartedAt:      status.StartedAt,
					ErrorMessage:   status.ErrorMessage,
				}

				data, err := json.Marshal(msg)
				if err == nil {
					select {
					case client.GetSendChannel() <- data:
					case <-time.After(time.Second):
						log.Printf("Timeout sending initial status to client")
					}
				}
			}
		}()

		// Start goroutines for reading and writing
		go writePump(client, broadcaster)
		go readPump(client, broadcaster)
	}
}

// readPump handles incoming messages from the WebSocket client
func readPump(client *services.Client, broadcaster *services.Broadcaster) {
	defer func() {
		broadcaster.Unregister(client)
		client.GetConn().Close()
	}()

	conn := client.GetConn()
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}
		// We don't expect messages from clients, but read to detect disconnects
	}
}

// writePump sends messages from the broadcaster to the WebSocket client
func writePump(client *services.Client, broadcaster *services.Broadcaster) {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		client.GetConn().Close()
	}()

	conn := client.GetConn()
	sendChan := client.GetSendChannel()

	for {
		select {
		case message, ok := <-sendChan:
			conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// Broadcaster closed the channel
				conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
