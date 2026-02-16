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
