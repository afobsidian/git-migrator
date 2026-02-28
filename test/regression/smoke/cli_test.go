package smoke

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
	// Get project root directory
	_, filename, _, _ := runtime.Caller(0)
	testDir := filepath.Dir(filename)
	projectRoot := filepath.Dir(filepath.Dir(filepath.Dir(testDir)))

	binaryName := "git-migrator-smoke-test"
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

// TestCLIBinaryBuilds tests that the CLI binary builds successfully
func TestCLIBinaryBuilds(t *testing.T) {
	// Binary was built in TestMain, just verify it exists
	_, err := exec.LookPath(binaryPath)
	assert.NoError(t, err, "Binary should exist")
}

// TestCLIRuns tests that the CLI runs without crashing
func TestCLIRuns(t *testing.T) {
	cmd := exec.Command(binaryPath, "--help")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	assert.NoError(t, err, "CLI should run with --help")
	assert.Contains(t, stdout.String(), "git-migrator", "Output should contain program name")
}

// TestCLIVersionCommand tests the version command works
func TestCLIVersionCommand(t *testing.T) {
	cmd := exec.Command(binaryPath, "version")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	err := cmd.Run()
	assert.NoError(t, err, "version command should succeed")
	assert.NotEmpty(t, stdout.String(), "version command should produce output")
}
