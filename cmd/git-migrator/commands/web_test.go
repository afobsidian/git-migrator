package commands

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWebCommandFlagDefault(t *testing.T) {
	// Default should be 8080 as set in init
	require.Equal(t, 8080, webPort)

	// Changing and restoring
	old := webPort
	webPort = 9090
	defer func() { webPort = old }()
	require.Equal(t, 9090, webPort)
}

func TestRunWeb_ErrorOnInvalidPort(t *testing.T) {
	old := webPort
	webPort = -1
	defer func() { webPort = old }()

	err := runWeb(nil, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to start web server")
}
