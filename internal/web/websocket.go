package web

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

// handleWebSocket handles WebSocket connections for progress updates
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	migrationID := chi.URLParam(r, "id")

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("Warning: failed to close WebSocket connection: %v", err)
		}
	}()

	// Send initial connection message
	s.sendProgressEvent(conn, migrationID, "connected", "Connected to migration progress")

	// Check if migration exists
	s.mu.RLock()
	migration, exists := s.migrations[migrationID]
	s.mu.RUnlock()

	if !exists {
		// Migration not found, send error and close
		s.sendProgressEvent(conn, migrationID, "error", "Migration not found")
		return
	}

	// Send current status
	s.sendFullProgress(conn, migration)

	// Keep connection alive and send updates
	for {
		// Check if connection is still open
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}

		// Send periodic status updates
		s.mu.RLock()
		currentMigration, stillExists := s.migrations[migrationID]
		s.mu.RUnlock()

		if !stillExists {
			s.sendProgressEvent(conn, migrationID, "error", "Migration no longer exists")
			break
		}

		// Send update if status changed
		s.sendFullProgress(conn, currentMigration)

		// If migration is complete, close connection
		if currentMigration.Status == "completed" || currentMigration.Status == "failed" || currentMigration.Status == "stopped" {
			s.sendProgressEvent(conn, migrationID, currentMigration.Status, "Migration "+currentMigration.Status)
			break
		}
	}
}

// sendProgressEvent sends a progress event to the WebSocket client
func (s *Server) sendProgressEvent(conn *websocket.Conn, migrationID, eventType, message string) {
	event := ProgressEvent{
		Type: eventType,
		Data: ProgressData{
			MigrationID:      migrationID,
			Status:           eventType,
			CurrentStep:      message,
			Percentage:       0,
			TotalCommits:     0,
			ProcessedCommits: 0,
			Errors:           []string{},
		},
	}
	s.sendJSON(conn, event)
}

// sendFullProgress sends a full progress update to the WebSocket client
func (s *Server) sendFullProgress(conn *websocket.Conn, migration *MigrationStatus) {
	event := ProgressEvent{
		Type: "progress",
		Data: ProgressData{
			MigrationID:      migration.ID,
			Status:           migration.Status,
			Percentage:       migration.Percentage,
			CurrentStep:      migration.CurrentStep,
			TotalCommits:     migration.TotalCommits,
			ProcessedCommits: migration.ProcessedCommits,
			Errors:           migration.Errors,
		},
	}
	s.sendJSON(conn, event)
}

// sendJSON sends a JSON message to the WebSocket client
func (s *Server) sendJSON(conn *websocket.Conn, v interface{}) {
	data, err := json.Marshal(v)
	if err != nil {
		log.Printf("JSON marshal error: %v", err)
		return
	}

	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		log.Printf("WebSocket write error: %v", err)
	}
}
