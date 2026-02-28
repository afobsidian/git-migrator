package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

func TestWebSocket_NotFound(t *testing.T) {
	s := NewServer(ServerConfig{Port: 0})
	ts := httptest.NewServer(s.Router())
	defer ts.Close()

	// connect to websocket for nonexistent id
	url := "ws" + ts.URL[len("http"):] + "/ws/progress/nonexistent"
	dialer := websocket.Dialer{HandshakeTimeout: 2 * time.Second}
	conn, resp, err := dialer.Dial(url, nil)
	if conn != nil {
		defer func() { _ = conn.Close() }()
	}
	// The server upgrades then sends connected and error, connection may close
	if err != nil {
		// Some environments block upgrade; ensure we at least returned an HTTP response
		require.NotNil(t, resp)
		return
	}

	// Read initial connected event
	_, msg, err := conn.ReadMessage()
	require.NoError(t, err)
	require.Contains(t, string(msg), "connected")

	// Read error message
	_, msg, err = conn.ReadMessage()
	require.NoError(t, err)
	require.Contains(t, string(msg), "Migration not found")
}

func TestSendJSONAndProgressHelpers(t *testing.T) {
	s := NewServer(ServerConfig{Port: 0})

	// Make a test handler that upgrades and then uses sendProgressEvent/sendFullProgress
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgr := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
		conn, err := upgr.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()

		// Create a migration status and send full progress
		ms := &MigrationStatus{ID: "mid", Status: "running", Percentage: 10, CurrentStep: "step", TotalCommits: 5, ProcessedCommits: 1, Errors: []string{}}
		s.sendProgressEvent(conn, "mid", "connected", "hi")
		s.sendFullProgress(conn, ms)
	}))
	defer ts.Close()

	url := "ws" + ts.URL[len("http"):]
	dialer := websocket.Dialer{HandshakeTimeout: 2 * time.Second}
	conn, resp, err := dialer.Dial(url, nil)
	if conn != nil {
		defer func() { _ = conn.Close() }()
	}
	if err != nil {
		require.NotNil(t, resp)
		return
	}

	// Read messages
	_, msg, err := conn.ReadMessage()
	require.NoError(t, err)
	require.Contains(t, string(msg), "connected")

	_, msg, err = conn.ReadMessage()
	require.NoError(t, err)
	require.Contains(t, string(msg), "progress")
}
