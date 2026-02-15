package requirements

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/adamf123git/git-migrator/internal/vcs/git"
)

// TestGitWriterInit tests repository initialization
func TestGitWriterInit(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "git-migrator-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("Warning: failed to remove temp dir: %v", err)
		}
	}()

	repoPath := filepath.Join(tmpDir, "test-repo")
	writer := git.NewWriter()

	// Initialize repository
	err = writer.Init(repoPath)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Check .git directory exists
	gitDir := filepath.Join(repoPath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		t.Error("Expected .git directory to exist")
	}
}

// TestGitWriterInitWithConfig tests initialization with config
func TestGitWriterInitWithConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-migrator-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("Warning: failed to remove temp dir: %v", err)
		}
	}()

	repoPath := filepath.Join(tmpDir, "test-repo")
	writer := git.NewWriter()

	config := map[string]string{
		"user.name":  "Test User",
		"user.email": "test@example.com",
	}

	err = writer.InitWithConfig(repoPath, config)
	if err != nil {
		t.Fatalf("InitWithConfig failed: %v", err)
	}

	// Verify config was set
	name, err := writer.GetConfig("user.name")
	if err != nil {
		t.Fatalf("GetConfig failed: %v", err)
	}
	if name != "Test User" {
		t.Errorf("Expected user.name='Test User', got %q", name)
	}
}

// TestGitWriterIsRepo tests repository detection
func TestGitWriterIsRepo(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-migrator-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("Warning: failed to remove temp dir: %v", err)
		}
	}()

	repoPath := filepath.Join(tmpDir, "test-repo")
	writer := git.NewWriter()

	// Not a repo yet
	if writer.IsRepo(repoPath) {
		t.Error("Expected IsRepo to return false for non-existent repo")
	}

	// Initialize
	err = writer.Init(repoPath)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Now should be a repo
	if !writer.IsRepo(repoPath) {
		t.Error("Expected IsRepo to return true for initialized repo")
	}
}

// TestGitWriterClose tests cleanup
func TestGitWriterClose(t *testing.T) {
	writer := git.NewWriter()
	err := writer.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
}
