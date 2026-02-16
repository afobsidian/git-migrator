package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

// type that cannot be marshaled to JSON easily
type badType struct{ C chan int }

func TestSendJSON_MarshalError(t *testing.T) {
	s := NewServer(ServerConfig{Port: 0})
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgr := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
		conn, err := upgr.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// Attempt to send a bad type
		s.sendJSON(conn, badType{})
	}))
	defer ts.Close()

	url := "ws" + ts.URL[len("http"):]
	conn, resp, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		require.NotNil(t, resp)
		return
	}
	defer conn.Close()

	// server will attempt to write but since marshal fails, no message expected; ensure no panic
}
