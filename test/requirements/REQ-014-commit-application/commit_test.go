package requirements

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/adamf123git/git-migrator/internal/vcs"
	"github.com/adamf123git/git-migrator/internal/vcs/git"
)

// TestApplyCommitAdd tests applying a commit with file addition
func TestApplyCommitAdd(t *testing.T) {
	writer, repoPath := setupTestRepo(t)
	defer func() {
		if err := os.RemoveAll(repoPath); err != nil {
			t.Logf("Warning: failed to remove temp repo: %v", err)
		}
	}()
	defer func() {
		if err := writer.Close(); err != nil {
			t.Logf("Warning: failed to close writer: %v", err)
		}
	}()

	commit := &vcs.Commit{
		Author:  "Test User",
		Email:   "test@example.com",
		Date:    time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		Message: "Initial commit",
		Files: []vcs.FileChange{
			{
				Path:    "README.md",
				Action:  vcs.ActionAdd,
				Content: []byte("# Test Project\n\nThis is a test."),
			},
		},
	}

	err := writer.ApplyCommit(commit)
	if err != nil {
		t.Fatalf("ApplyCommit failed: %v", err)
	}

	// Verify file exists
	content, err := os.ReadFile(filepath.Join(repoPath, "README.md"))
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	if string(content) != "# Test Project\n\nThis is a test." {
		t.Errorf("Unexpected file content: %s", content)
	}
}

// TestApplyCommitModify tests applying a commit with file modification
func TestApplyCommitModify(t *testing.T) {
	writer, repoPath := setupTestRepo(t)
	defer func() {
		if err := os.RemoveAll(repoPath); err != nil {
			t.Logf("Warning: failed to remove temp repo: %v", err)
		}
	}()
	defer func() {
		if err := writer.Close(); err != nil {
			t.Logf("Warning: failed to close writer: %v", err)
		}
	}()

	// First commit
	commit1 := &vcs.Commit{
		Author:  "Test User",
		Email:   "test@example.com",
		Date:    time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		Message: "Initial commit",
		Files: []vcs.FileChange{
			{
				Path:    "file.txt",
				Action:  vcs.ActionAdd,
				Content: []byte("Version 1"),
			},
		},
	}
	if err := writer.ApplyCommit(commit1); err != nil {
		t.Fatalf("ApplyCommit 1 failed: %v", err)
	}

	// Second commit (modification)
	commit2 := &vcs.Commit{
		Author:  "Test User",
		Email:   "test@example.com",
		Date:    time.Date(2024, 1, 16, 10, 30, 0, 0, time.UTC),
		Message: "Update file",
		Files: []vcs.FileChange{
			{
				Path:    "file.txt",
				Action:  vcs.ActionModify,
				Content: []byte("Version 2"),
			},
		},
	}
	if err := writer.ApplyCommit(commit2); err != nil {
		t.Fatalf("ApplyCommit 2 failed: %v", err)
	}

	// Verify file was updated
	content, err := os.ReadFile(filepath.Join(repoPath, "file.txt"))
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	if string(content) != "Version 2" {
		t.Errorf("Expected 'Version 2', got %s", content)
	}
}

// TestApplyCommitDelete tests applying a commit with file deletion
func TestApplyCommitDelete(t *testing.T) {
	writer, repoPath := setupTestRepo(t)
	defer func() {
		if err := os.RemoveAll(repoPath); err != nil {
			t.Logf("Warning: failed to remove temp repo: %v", err)
		}
	}()
	defer func() {
		if err := writer.Close(); err != nil {
			t.Logf("Warning: failed to close writer: %v", err)
		}
	}()

	// First commit - add file
	commit1 := &vcs.Commit{
		Author:  "Test User",
		Email:   "test@example.com",
		Date:    time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		Message: "Add file",
		Files: []vcs.FileChange{
			{
				Path:    "todelete.txt",
				Action:  vcs.ActionAdd,
				Content: []byte("This will be deleted"),
			},
		},
	}
	if err := writer.ApplyCommit(commit1); err != nil {
		t.Fatalf("ApplyCommit 1 failed: %v", err)
	}

	// Second commit - delete file
	commit2 := &vcs.Commit{
		Author:  "Test User",
		Email:   "test@example.com",
		Date:    time.Date(2024, 1, 16, 10, 30, 0, 0, time.UTC),
		Message: "Delete file",
		Files: []vcs.FileChange{
			{
				Path:   "todelete.txt",
				Action: vcs.ActionDelete,
			},
		},
	}
	if err := writer.ApplyCommit(commit2); err != nil {
		t.Fatalf("ApplyCommit 2 failed: %v", err)
	}

	// Verify file was deleted
	_, err := os.Stat(filepath.Join(repoPath, "todelete.txt"))
	if !os.IsNotExist(err) {
		t.Error("Expected file to be deleted")
	}
}

// TestApplyCommitMultipleFiles tests applying multiple files in one commit
func TestApplyCommitMultipleFiles(t *testing.T) {
	writer, repoPath := setupTestRepo(t)
	defer func() {
		if err := os.RemoveAll(repoPath); err != nil {
			t.Logf("Warning: failed to remove temp repo: %v", err)
		}
	}()
	defer func() {
		if err := writer.Close(); err != nil {
			t.Logf("Warning: failed to close writer: %v", err)
		}
	}()

	commit := &vcs.Commit{
		Author:  "Test User",
		Email:   "test@example.com",
		Date:    time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		Message: "Add multiple files",
		Files: []vcs.FileChange{
			{Path: "file1.txt", Action: vcs.ActionAdd, Content: []byte("File 1")},
			{Path: "file2.txt", Action: vcs.ActionAdd, Content: []byte("File 2")},
			{Path: "dir/file3.txt", Action: vcs.ActionAdd, Content: []byte("File 3")},
		},
	}

	err := writer.ApplyCommit(commit)
	if err != nil {
		t.Fatalf("ApplyCommit failed: %v", err)
	}

	// Verify all files exist
	for _, fc := range commit.Files {
		content, err := os.ReadFile(filepath.Join(repoPath, fc.Path))
		if err != nil {
			t.Errorf("Failed to read %s: %v", fc.Path, err)
			continue
		}
		if string(content) != string(fc.Content) {
			t.Errorf("Unexpected content in %s", fc.Path)
		}
	}
}

// TestCommitMetadata tests that commit metadata is preserved
func TestCommitMetadata(t *testing.T) {
	writer, repoPath := setupTestRepo(t)
	defer func() {
		if err := os.RemoveAll(repoPath); err != nil {
			t.Logf("Warning: failed to remove temp repo: %v", err)
		}
	}()
	defer func() {
		if err := writer.Close(); err != nil {
			t.Logf("Warning: failed to close writer: %v", err)
		}
	}()

	commit := &vcs.Commit{
		Author:  "John Doe",
		Email:   "john@example.com",
		Date:    time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		Message: "Test commit message",
		Files: []vcs.FileChange{
			{Path: "test.txt", Action: vcs.ActionAdd, Content: []byte("test")},
		},
	}

	err := writer.ApplyCommit(commit)
	if err != nil {
		t.Fatalf("ApplyCommit failed: %v", err)
	}

	// Verify commit metadata via log
	log, err := writer.GetLastCommit()
	if err != nil {
		t.Fatalf("GetLastCommit failed: %v", err)
	}

	if log.Author != "John Doe" {
		t.Errorf("Expected author 'John Doe', got %q", log.Author)
	}
	if log.Email != "john@example.com" {
		t.Errorf("Expected email 'john@example.com', got %q", log.Email)
	}
	if log.Message != "Test commit message" {
		t.Errorf("Expected message 'Test commit message', got %q", log.Message)
	}
}

// setupTestRepo creates a test repository and returns it
func setupTestRepo(t *testing.T) (*git.Writer, string) {
	tmpDir, err := os.MkdirTemp("", "git-migrator-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	repoPath := filepath.Join(tmpDir, "test-repo")
	writer := git.NewWriter()

	err = writer.Init(repoPath)
	if err != nil {
		if removeErr := os.RemoveAll(tmpDir); removeErr != nil {
			t.Logf("Warning: failed to remove temp dir: %v", removeErr)
		}
		t.Fatalf("Failed to init repo: %v", err)
	}

	return writer, repoPath
}
