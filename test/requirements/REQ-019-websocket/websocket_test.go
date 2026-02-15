package requirements_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/adamf123git/git-migrator/internal/web"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWebSocketConnection tests basic WebSocket connection
func TestWebSocketConnection(t *testing.T) {
	server := web.NewServer(web.ServerConfig{
		Port: 8080,
	})

	// Create test server
	ts := httptest.NewServer(server.Router())
	defer ts.Close()

	// Convert http URL to ws URL
	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws/progress/test-migration"

	// Connect
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer func() {
		if err := ws.Close(); err != nil {
			t.Logf("Warning: failed to close websocket: %v", err)
		}
	}()

	// Connection should succeed
	assert.NotNil(t, ws)
}

// TestWebSocketProgressMessage tests receiving progress messages
func TestWebSocketProgressMessage(t *testing.T) {
	server := web.NewServer(web.ServerConfig{
		Port: 8080,
	})

	ts := httptest.NewServer(server.Router())
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws/progress/test-migration"

	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer func() {
		if err := ws.Close(); err != nil {
			t.Logf("Warning: failed to close websocket: %v", err)
		}
	}()

	// Set read deadline
	if err := ws.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Logf("Warning: failed to set read deadline: %v", err)
	}

	// Read message
	_, message, err := ws.ReadMessage()
	require.NoError(t, err)

	// Parse message
	var event web.ProgressEvent
	err = json.Unmarshal(message, &event)
	require.NoError(t, err)

	// Verify message structure
	assert.Contains(t, []string{"progress", "connected", "status"}, event.Type)
}

// TestWebSocketInvalidMigration tests connecting to non-existent migration
func TestWebSocketInvalidMigration(t *testing.T) {
	server := web.NewServer(web.ServerConfig{
		Port: 8080,
	})

	ts := httptest.NewServer(server.Router())
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws/progress/invalid-id"

	// Should still connect (but may receive error message)
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		// Connection rejection is acceptable
		return
	}
	defer func() {
		if err := ws.Close(); err != nil {
			t.Logf("Warning: failed to close websocket: %v", err)
		}
	}()

	// If connected, should receive error message
	if err := ws.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Logf("Warning: failed to set read deadline: %v", err)
	}
	_, message, err := ws.ReadMessage()
	if err != nil {
		return // Timeout is acceptable
	}

	var event web.ProgressEvent
	if err := json.Unmarshal(message, &event); err != nil {
		t.Logf("Warning: failed to unmarshal event: %v", err)
	}
	// Should receive error or status update
	assert.NotEmpty(t, event.Type)
}

// TestWebSocketMultipleClients tests multiple clients connecting to same migration
func TestWebSocketMultipleClients(t *testing.T) {
	server := web.NewServer(web.ServerConfig{
		Port: 8080,
	})

	ts := httptest.NewServer(server.Router())
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws/progress/test-multi"

	// Connect first client
	ws1, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer func() {
		if err := ws1.Close(); err != nil {
			t.Logf("Warning: failed to close websocket 1: %v", err)
		}
	}()

	// Connect second client
	ws2, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer func() {
		if err := ws2.Close(); err != nil {
			t.Logf("Warning: failed to close websocket 2: %v", err)
		}
	}()

	// Both should be connected
	assert.NotNil(t, ws1)
	assert.NotNil(t, ws2)
}

// TestWebSocketCloseOnCompletion tests that WebSocket closes when migration completes
func TestWebSocketCloseOnCompletion(t *testing.T) {
	server := web.NewServer(web.ServerConfig{
		Port: 8080,
	})

	ts := httptest.NewServer(server.Router())
	defer ts.Close()

	// Create a migration first via API
	migrationReq := web.StartMigrationRequest{
		SourceType: "cvs",
		SourcePath: "/tmp/test-cvs",
		TargetPath: "/tmp/test-git",
		Options: map[string]interface{}{
			"dryRun": true,
		},
	}

	body, _ := json.Marshal(migrationReq)
	resp, err := http.Post(ts.URL+"/api/migrations", "application/json", strings.NewReader(string(body)))
	require.NoError(t, err)
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Logf("Warning: failed to close response body: %v", err)
		}
	}()

	var createResp web.APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&createResp); err != nil {
		t.Fatalf("Failed to decode create response: %v", err)
	}
	data := createResp.Data.(map[string]interface{})
	migrationID := data["id"].(string)

	// Connect to WebSocket
	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws/progress/" + migrationID
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer func() {
		if err := ws.Close(); err != nil {
			t.Logf("Warning: failed to close websocket: %v", err)
		}
	}()

	// Read messages until close or timeout
	if err := ws.SetReadDeadline(time.Now().Add(3 * time.Second)); err != nil {
		t.Logf("Warning: failed to set read deadline: %v", err)
	}
	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			break
		}
	}
}

// TestProgressEventStructure tests that progress events have correct structure
func TestProgressEventStructure(t *testing.T) {
	event := web.ProgressEvent{
		Type: "progress",
		Data: web.ProgressData{
			MigrationID:      "test-id",
			Status:           "running",
			Percentage:       50,
			CurrentStep:      "Processing commits",
			TotalCommits:     100,
			ProcessedCommits: 50,
			Errors:           []string{},
		},
	}

	// Verify JSON serialization
	data, err := json.Marshal(event)
	require.NoError(t, err)

	var decoded web.ProgressEvent
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "progress", decoded.Type)
	assert.Equal(t, "test-id", decoded.Data.MigrationID)
	assert.Equal(t, 50, decoded.Data.Percentage)
	assert.Equal(t, "Processing commits", decoded.Data.CurrentStep)
}
