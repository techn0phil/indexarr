package api

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"indexarr/internal/config"
	"indexarr/internal/models"
	"indexarr/internal/repository"
	"indexarr/internal/services"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

func TestHandleWebSocket_InitialStatusMessageMapping(t *testing.T) {
	logger := zerolog.Nop()
	config.GlobalLogger = &logger

	tests := []struct {
		name        string
		status      models.ScanStatus
		expectedMsg string
	}{
		{
			name:        "idle status maps to scan_idle",
			status:      models.ScanStatus{Status: "idle", FilesFound: 0, FilesProcessed: 0},
			expectedMsg: "scan_idle",
		},
		{
			name: "running status maps to scan_progress",
			status: models.ScanStatus{
				Status:         "running",
				FilesFound:     10,
				FilesProcessed: 3,
				StartedAt:      "2026-08-25T10:00:00Z",
			},
			expectedMsg: "scan_progress",
		},
		{
			name: "completed status maps to scan_complete",
			status: models.ScanStatus{
				Status:         "completed",
				FilesFound:     8,
				FilesProcessed: 8,
				CompletedAt:    "2026-08-25T10:15:00Z",
			},
			expectedMsg: "scan_complete",
		},
		{
			name: "error status maps to scan_error",
			status: models.ScanStatus{
				Status:         "error",
				FilesFound:     8,
				FilesProcessed: 2,
				ErrorMessage:   "boom",
			},
			expectedMsg: "scan_error",
		},
		{
			name: "stopped status maps to scan_stopped",
			status: models.ScanStatus{
				Status:         "stopped",
				FilesFound:     8,
				FilesProcessed: 4,
			},
			expectedMsg: "scan_stopped",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupAPITestDBWithMigrations(t)

			if _, err := repository.GetScanStatus(db); err != nil {
				t.Fatalf("failed to initialize scan status: %v", err)
			}
			if err := repository.UpdateScanStatus(db, &tt.status); err != nil {
				t.Fatalf("failed to set scan status: %v", err)
			}

			broadcaster := services.NewBroadcaster()
			go broadcaster.Run()

			server := httptest.NewServer(HandleWebSocket(db, broadcaster))
			defer server.Close()

			wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
			conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			if err != nil {
				t.Fatalf("failed to connect websocket: %v", err)
			}
			defer conn.Close()

			if err := conn.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
				t.Fatalf("failed to set read deadline: %v", err)
			}

			_, data, err := conn.ReadMessage()
			if err != nil {
				t.Fatalf("failed to read initial websocket message: %v", err)
			}

			var msg services.WSMessage
			if err := json.Unmarshal(data, &msg); err != nil {
				t.Fatalf("failed to decode websocket message: %v", err)
			}
			if msg.Type != tt.expectedMsg {
				t.Fatalf("unexpected message type: got %q want %q", msg.Type, tt.expectedMsg)
			}
		})
	}
}
