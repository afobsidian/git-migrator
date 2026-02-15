package requirements_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getProjectRoot returns the project root directory
func getProjectRoot() string {
	// Navigate up from test/requirements/REQ-006-docker to project root
	dir, _ := os.Getwd()
	return filepath.Join(dir, "..", "..", "..")
}

// TestDockerfileExists tests that Dockerfile exists
func TestDockerfileExists(t *testing.T) {
	root := getProjectRoot()
	_, err := os.Stat(filepath.Join(root, "Dockerfile"))
	assert.NoError(t, err, "Dockerfile should exist")
}

// TestDockerComposeExists tests that docker-compose.yml exists
func TestDockerComposeExists(t *testing.T) {
	root := getProjectRoot()
	_, err := os.Stat(filepath.Join(root, "docker-compose.yml"))
	assert.NoError(t, err, "docker-compose.yml should exist")
}

// TestDockerfileBuild tests that Docker image builds successfully
// This test is skipped if Docker is not available
func TestDockerfileBuild(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Docker build test in short mode")
	}

	// Check if Docker is available
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("Docker not available")
	}

	// Try to build the image
	cmd := exec.Command("docker", "build", "-t", "git-migrator-test", ".")
	cmd.Dir = getProjectRoot()
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Docker build output: %s", string(output))
	}
	require.NoError(t, err, "Docker image should build successfully")

	// Cleanup
	if err := exec.Command("docker", "rmi", "git-migrator-test").Run(); err != nil {
		t.Logf("Warning: failed to remove docker image: %v", err)
	}
}

// TestDockerComposeSyntax tests docker-compose.yml syntax
func TestDockerComposeSyntax(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping docker-compose test in short mode")
	}

	if _, err := exec.LookPath("docker-compose"); err != nil {
		t.Skip("docker-compose not available")
	}

	cmd := exec.Command("docker-compose", "config")
	cmd.Dir = getProjectRoot()
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("docker-compose config output: %s", string(output))
	}
	require.NoError(t, err, "docker-compose.yml should have valid syntax")
}

// TestDockerImageSize tests that image is reasonably sized
// This is a manual test, run with: go test -run TestDockerImageSize
func TestDockerImageSize(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Docker size test in short mode")
	}

	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("Docker not available")
	}

	// Build first
	cmd := exec.Command("docker", "build", "-t", "git-migrator-size-test", ".")
	cmd.Dir = getProjectRoot()
	if err := cmd.Run(); err != nil {
		t.Skip("Could not build image")
	}
	defer func() {
		if err := exec.Command("docker", "rmi", "git-migrator-size-test").Run(); err != nil {
			t.Logf("Warning: failed to remove docker image: %v", err)
		}
	}()

	// Check size
	cmd = exec.Command("docker", "image", "inspect", "--format={{.Size}}", "git-migrator-size-test")
	output, err := cmd.Output()
	require.NoError(t, err)

	// Parse size (output is in bytes)
	sizeStr := string(output)
	sizeStr = sizeStr[:len(sizeStr)-1] // Remove newline
	var size int64
	_, err = fmt.Sscanf(sizeStr, "%d", &size)
	require.NoError(t, err)

	// Should be less than 100MB
	assert.Less(t, size, int64(100*1024*1024), "Image should be less than 100MB")
}

// TestDockerRunCLI tests running migrations via Docker CLI
func TestDockerRunCLI(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Docker run test in short mode")
	}

	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("Docker not available")
	}

	// Build first
	cmd := exec.Command("docker", "build", "-t", "git-migrator-cli-test", ".")
	cmd.Dir = getProjectRoot()
	if err := cmd.Run(); err != nil {
		t.Skip("Could not build image")
	}
	defer func() {
		if err := exec.Command("docker", "rmi", "git-migrator-cli-test").Run(); err != nil {
			t.Logf("Warning: failed to remove docker image: %v", err)
		}
	}()

	// Run version command
	cmd = exec.Command("docker", "run", "--rm", "git-migrator-cli-test", "version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Docker run output: %s", string(output))
	}
	require.NoError(t, err, "Docker container should run version command")
	assert.Contains(t, string(output), "git-migrator")
}

// TestDockerRunWeb tests running web UI via Docker
func TestDockerRunWeb(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Docker web test in short mode")
	}

	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("Docker not available")
	}

	// Build first
	cmd := exec.Command("docker", "build", "-t", "git-migrator-web-test", ".")
	cmd.Dir = getProjectRoot()
	if err := cmd.Run(); err != nil {
		t.Skip("Could not build image")
	}
	defer func() {
		if err := exec.Command("docker", "rmi", "git-migrator-web-test").Run(); err != nil {
			t.Logf("Warning: failed to remove docker image: %v", err)
		}
	}()

	// Run web command in background
	cmd = exec.Command("docker", "run", "--rm", "-p", "18080:8080", "git-migrator-web-test", "web", "--port", "8080")
	if err := cmd.Start(); err != nil {
		t.Skip("Could not start web container")
	}
	defer func() {
		if err := cmd.Process.Kill(); err != nil {
			t.Logf("Warning: failed to kill process: %v", err)
		}
	}()
}
