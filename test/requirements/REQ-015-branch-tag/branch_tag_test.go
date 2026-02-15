package requirements

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/adamf123git/git-migrator/internal/vcs"
	"github.com/adamf123git/git-migrator/internal/vcs/git"
)

// TestCreateBranch tests creating a branch
func TestCreateBranch(t *testing.T) {
	writer, repoPath := setupTestRepoWithCommit(t)
	defer func() {
		if err := os.RemoveAll(filepath.Dir(filepath.Dir(repoPath))); err != nil {
			t.Logf("Warning: failed to remove temp dir: %v", err)
		}
	}()
	defer func() {
		if err := writer.Close(); err != nil {
			t.Logf("Warning: failed to close writer: %v", err)
		}
	}()

	// Create a branch
	err := writer.CreateBranch("feature-branch", "HEAD")
	if err != nil {
		t.Fatalf("CreateBranch failed: %v", err)
	}

	// Verify branch exists
	branches, err := writer.ListBranches()
	if err != nil {
		t.Fatalf("ListBranches failed: %v", err)
	}

	found := false
	for _, b := range branches {
		if b == "feature-branch" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected feature-branch to exist")
	}
}

// TestCreateBranchFromRevision tests creating a branch from specific revision
func TestCreateBranchFromRevision(t *testing.T) {
	writer, repoPath := setupTestRepoWithCommits(t, 2)
	defer func() {
		if err := os.RemoveAll(filepath.Dir(filepath.Dir(repoPath))); err != nil {
			t.Logf("Warning: failed to remove temp dir: %v", err)
		}
	}()
	defer func() {
		if err := writer.Close(); err != nil {
			t.Logf("Warning: failed to close writer: %v", err)
		}
	}()

	// Get first commit hash
	hashes, err := writer.GetCommitHashes()
	if err != nil {
		t.Fatalf("GetCommitHashes failed: %v", err)
	}
	if len(hashes) < 2 {
		t.Fatal("Expected at least 2 commits")
	}
	firstCommit := hashes[len(hashes)-1] // Oldest first

	// Create branch from first commit
	err = writer.CreateBranch("from-first", firstCommit)
	if err != nil {
		t.Fatalf("CreateBranch failed: %v", err)
	}

	// Verify branch exists
	branches, err := writer.ListBranches()
	if err != nil {
		t.Fatalf("ListBranches failed: %v", err)
	}

	found := false
	for _, b := range branches {
		if b == "from-first" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected from-first branch to exist")
	}
}

// TestCreateTag tests creating a lightweight tag
func TestCreateTag(t *testing.T) {
	writer, repoPath := setupTestRepoWithCommit(t)
	defer func() {
		if err := os.RemoveAll(filepath.Dir(filepath.Dir(repoPath))); err != nil {
			t.Logf("Warning: failed to remove temp dir: %v", err)
		}
	}()
	defer func() {
		if err := writer.Close(); err != nil {
			t.Logf("Warning: failed to close writer: %v", err)
		}
	}()

	// Create a tag
	err := writer.CreateTag("v1.0.0", "HEAD", "")
	if err != nil {
		t.Fatalf("CreateTag failed: %v", err)
	}

	// Verify tag exists
	tags, err := writer.ListTags()
	if err != nil {
		t.Fatalf("ListTags failed: %v", err)
	}

	if _, ok := tags["v1.0.0"]; !ok {
		t.Error("Expected v1.0.0 tag to exist")
	}
}

// TestCreateAnnotatedTag tests creating an annotated tag
func TestCreateAnnotatedTag(t *testing.T) {
	writer, repoPath := setupTestRepoWithCommit(t)
	defer func() {
		if err := os.RemoveAll(filepath.Dir(filepath.Dir(repoPath))); err != nil {
			t.Logf("Warning: failed to remove temp dir: %v", err)
		}
	}()
	defer func() {
		if err := writer.Close(); err != nil {
			t.Logf("Warning: failed to close writer: %v", err)
		}
	}()

	// Create an annotated tag
	err := writer.CreateTag("v1.0.0", "HEAD", "Release version 1.0.0")
	if err != nil {
		t.Fatalf("CreateTag failed: %v", err)
	}

	// Verify tag exists and is annotated
	tags, err := writer.ListTags()
	if err != nil {
		t.Fatalf("ListTags failed: %v", err)
	}

	if _, ok := tags["v1.0.0"]; !ok {
		t.Error("Expected v1.0.0 tag to exist")
	}
}

// TestListBranches tests listing branches
func TestListBranches(t *testing.T) {
	writer, repoPath := setupTestRepoWithCommit(t)
	defer func() {
		if err := os.RemoveAll(filepath.Dir(filepath.Dir(repoPath))); err != nil {
			t.Logf("Warning: failed to remove temp dir: %v", err)
		}
	}()
	defer func() {
		if err := writer.Close(); err != nil {
			t.Logf("Warning: failed to close writer: %v", err)
		}
	}()

	// Create multiple branches
	branches := []string{"feature-a", "feature-b", "bugfix-1"}
	for _, b := range branches {
		err := writer.CreateBranch(b, "HEAD")
		if err != nil {
			t.Fatalf("CreateBranch failed: %v", err)
		}
	}

	// List branches
	list, err := writer.ListBranches()
	if err != nil {
		t.Fatalf("ListBranches failed: %v", err)
	}

	// Verify all branches exist
	for _, expected := range branches {
		found := false
		for _, actual := range list {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected branch %s in list", expected)
		}
	}
}

// TestListTags tests listing tags
func TestListTags(t *testing.T) {
	writer, repoPath := setupTestRepoWithCommit(t)
	defer func() {
		if err := os.RemoveAll(filepath.Dir(filepath.Dir(repoPath))); err != nil {
			t.Logf("Warning: failed to remove temp dir: %v", err)
		}
	}()
	defer func() {
		if err := writer.Close(); err != nil {
			t.Logf("Warning: failed to close writer: %v", err)
		}
	}()

	// Create multiple tags
	tags := []string{"v1.0.0", "v1.0.1", "v2.0.0-beta"}
	for _, tag := range tags {
		err := writer.CreateTag(tag, "HEAD", "")
		if err != nil {
			t.Fatalf("CreateTag failed: %v", err)
		}
	}

	// List tags
	list, err := writer.ListTags()
	if err != nil {
		t.Fatalf("ListTags failed: %v", err)
	}

	// Verify all tags exist
	for _, expected := range tags {
		if _, ok := list[expected]; !ok {
			t.Errorf("Expected tag %s in list", expected)
		}
	}
}

// setupTestRepoWithCommit creates a repo with a single commit
func setupTestRepoWithCommit(t *testing.T) (*git.Writer, string) {
	tmpDir, err := os.MkdirTemp("", "git-migrator-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	repoPath := filepath.Join(tmpDir, "test-repo")
	writer := git.NewWriter()

	err = writer.Init(repoPath)
	if err != nil {
		err = os.RemoveAll(tmpDir)
		if err != nil {
			t.Logf("Warning: failed to remove temp dir: %v", err)
		}
		t.Fatalf("Failed to init repo: %v", err)
	}

	// Add a commit
	commit := &vcs.Commit{
		Author:  "Test User",
		Email:   "test@example.com",
		Date:    time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		Message: "Initial commit",
		Files: []vcs.FileChange{
			{Path: "README.md", Action: vcs.ActionAdd, Content: []byte("# Test")},
		},
	}
	if err := writer.ApplyCommit(commit); err != nil {
		err = os.RemoveAll(tmpDir)
		if err != nil {
			t.Logf("Warning: failed to remove temp dir: %v", err)
		}
		t.Fatalf("Failed to add commit: %v", err)
	}

	return writer, repoPath
}

// setupTestRepoWithCommits creates a repo with multiple commits
func setupTestRepoWithCommits(t *testing.T, count int) (*git.Writer, string) {
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

	// Add commits with different content
	for i := 0; i < count; i++ {
		content := fmt.Sprintf("Content %d", i)
		filename := fmt.Sprintf("file%d.txt", i)
		commit := &vcs.Commit{
			Author:  "Test User",
			Email:   "test@example.com",
			Date:    time.Date(2024, 1, 15+i, 10, 30, 0, 0, time.UTC),
			Message: fmt.Sprintf("Commit %d", i),
			Files: []vcs.FileChange{
				{Path: filename, Action: vcs.ActionAdd, Content: []byte(content)},
			},
		}
		if err := writer.ApplyCommit(commit); err != nil {
			if removeErr := os.RemoveAll(tmpDir); removeErr != nil {
				t.Logf("Warning: failed to remove temp dir: %v", removeErr)
			}
			t.Fatalf("Failed to add commit: %v", err)
		}
	}

	return writer, repoPath
}
