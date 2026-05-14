package services

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

// WSMessage represents a WebSocket message sent to clients
type WSMessage struct {
	Type           string `json:"type"`
	FilesFound     int    `json:"filesFound,omitempty"`
	FilesProcessed int    `json:"filesProcessed,omitempty"`
	StartedAt      string `json:"startedAt,omitempty"`
	ErrorMessage   string `json:"error,omitempty"`
	MoviesAdded    int    `json:"moviesAdded,omitempty"`
	EpisodesAdded  int    `json:"episodesAdded,omitempty"`
}

// Client represents a WebSocket client connection
type Client struct {
	conn *websocket.Conn
	send chan []byte
}

// NewClient creates a new WebSocket client
func NewClient(conn *websocket.Conn) *Client {
	return &Client{
		conn: conn,
		send: make(chan []byte, 256),
	}
}

// GetConn returns the client's WebSocket connection
func (c *Client) GetConn() *websocket.Conn {
	return c.conn
}

// GetSendChannel returns the client's send channel
func (c *Client) GetSendChannel() chan []byte {
	return c.send
}

// Broadcaster manages WebSocket connections and broadcasts messages
type Broadcaster struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
	mu         sync.RWMutex
}

// NewBroadcaster creates a new broadcaster instance
func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte, 256),
	}
}

// Run starts the broadcaster's main loop
func (b *Broadcaster) Run() {
	for {
		select {
		case client := <-b.register:
			b.mu.Lock()
			b.clients[client] = true
			b.mu.Unlock()
			log.Printf("WebSocket client connected (total: %d)", len(b.clients))

		case client := <-b.unregister:
			b.mu.Lock()
			if _, ok := b.clients[client]; ok {
				delete(b.clients, client)
				close(client.send)
				log.Printf("WebSocket client disconnected (total: %d)", len(b.clients))
			}
			b.mu.Unlock()

		case message := <-b.broadcast:
			b.mu.RLock()
			for client := range b.clients {
				select {
				case client.send <- message:
				default:
					// Client's send buffer is full, disconnect it
					close(client.send)
					delete(b.clients, client)
				}
			}
			b.mu.RUnlock()
		}
	}
}

// Register adds a new WebSocket client
func (b *Broadcaster) Register(client *Client) {
	b.register <- client
}

// Unregister removes a WebSocket client
func (b *Broadcaster) Unregister(client *Client) {
	b.unregister <- client
}

// BroadcastScanStart sends scan start event to all clients
func (b *Broadcaster) BroadcastScanStart(filesFound int, startedAt string) {
	msg := WSMessage{
		Type:       "scan_start",
		FilesFound: filesFound,
		StartedAt:  startedAt,
	}
	b.broadcastMessage(msg)
}

// BroadcastScanProgress sends scan progress update to all clients
func (b *Broadcaster) BroadcastScanProgress(filesProcessed, filesFound int) {
	msg := WSMessage{
		Type:           "scan_progress",
		FilesProcessed: filesProcessed,
		FilesFound:     filesFound,
	}
	b.broadcastMessage(msg)
}

// BroadcastScanComplete sends scan completion event to all clients
func (b *Broadcaster) BroadcastScanComplete(filesProcessed, moviesAdded, episodesAdded int) {
	msg := WSMessage{
		Type:           "scan_complete",
		FilesProcessed: filesProcessed,
		MoviesAdded:    moviesAdded,
		EpisodesAdded:  episodesAdded,
	}
	b.broadcastMessage(msg)
}

// BroadcastScanError sends scan error event to all clients
func (b *Broadcaster) BroadcastScanError(errorMessage string) {
	msg := WSMessage{
		Type:         "scan_error",
		ErrorMessage: errorMessage,
	}
	b.broadcastMessage(msg)
}

// BroadcastScanStopped sends scan stopped event to all clients
func (b *Broadcaster) BroadcastScanStopped() {
	msg := WSMessage{
		Type: "scan_stopped",
	}
	b.broadcastMessage(msg)
}

// broadcastMessage serializes and broadcasts a message to all clients
func (b *Broadcaster) broadcastMessage(msg WSMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to marshal WebSocket message: %v", err)
		return
	}
	b.broadcast <- data
}

// ClientCount returns the number of connected clients
func (b *Broadcaster) ClientCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.clients)
}
