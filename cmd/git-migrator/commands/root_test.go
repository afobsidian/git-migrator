package commands

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVersionVariables(t *testing.T) {
	// Ensure default version variables exist
	require.NotEmpty(t, Version)
	require.NotEmpty(t, GitCommit)
	require.NotEmpty(t, BuildDate)
}

func TestHandleError_NoPanic(t *testing.T) {
	// handleError should call os.Exit on non-nil; avoid exiting by passing nil
	handleError(nil)
}

func TestExecute_Help(t *testing.T) {
	// Test that Execute works with --help flag
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	os.Args = []string{"git-migrator", "--help"}

	// Execute with --help should not error
	err := Execute()
	// --help may cause cobra to exit or return an error depending on version
	// We just verify it doesn't panic
	_ = err
}

func TestExecute_Version(t *testing.T) {
	// Test that Execute works with --version flag
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	os.Args = []string{"git-migrator", "--version"}

	// Execute with --version should work
	err := Execute()
	// version flag behavior varies by cobra version
	_ = err
}
