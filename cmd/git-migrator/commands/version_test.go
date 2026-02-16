package commands

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVersionCommand_RunDoesNotPanic(t *testing.T) {
	// Call the Run function to ensure printing works (no panic)
	versionCmd.Run(nil, nil)
	require.True(t, true)
}
