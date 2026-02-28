package requirements

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

var binaryPath string

// TestMain builds the binary once for all tests
func TestMain(m *testing.M) {
	// Get project root by finding this test file's location and going up 3 levels
	_, filename, _, _ := runtime.Caller(0)
	testDir := filepath.Dir(filename)
	projectRoot := filepath.Dir(filepath.Dir(filepath.Dir(testDir)))

	binaryName := "git-migrator-cli-test"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}
	binaryPath = filepath.Join(os.TempDir(), binaryName)

	// Build the binary
	cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/git-migrator")
	cmd.Dir = projectRoot
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		panic("Failed to build binary: " + stderr.String())
	}

	// Run tests
	code := m.Run()

	// Cleanup
	if err := os.Remove(binaryPath); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to remove binary %s: %v\n", binaryPath, err)
	}
	os.Exit(code)
}

// TestCLICommands tests CLI commands
func TestCLICommands(t *testing.T) {
	tests := []struct {
		name    string
		command string
	}{
		{"version exists", "version"},
		{"help exists", "help"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, tt.command)
			assert.NoError(t, cmd.Run(), "Command failed")
		})
	}
}

// TestCLIVersion tests `version` command
func TestCLIVersion(t *testing.T) {
	cmd := exec.Command(binaryPath, "version")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	assert.NoError(t, cmd.Run(), "Version command failed")
	assert.NotEmpty(t, stdout.String(), "Version output should not be empty")
}

// TestCLIHelp tests `help` command
func TestCLIHelp(t *testing.T) {
	cmd := exec.Command(binaryPath, "help")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	assert.NoError(t, cmd.Run(), "Help command failed")
	assert.NotEmpty(t, stdout.String(), "Help output should not be empty")
}
