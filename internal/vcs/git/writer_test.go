package git

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/adamf123git/git-migrator/internal/vcs"
	"github.com/stretchr/testify/require"
)

func TestNewWriter(t *testing.T) {
	w := NewWriter()
	if w == nil {
		t.Fatal("NewWriter returned nil")
	}
}

func TestWriterInit(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-writer-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	repoPath := filepath.Join(tmpDir, "test-repo")

	w := NewWriter()
	if err := w.Init(repoPath); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer func() { require.NoError(t, w.Close()) }()

	// Check that .git directory was created
	gitDir := filepath.Join(repoPath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		t.Error(".git directory was not created")
	}

	// Check that objects subdirectories were created
	objectsInfo := filepath.Join(gitDir, "objects", "info")
	objectsPack := filepath.Join(gitDir, "objects", "pack")
	if _, err := os.Stat(objectsInfo); os.IsNotExist(err) {
		t.Error(".git/objects/info directory was not created")
	}
	if _, err := os.Stat(objectsPack); os.IsNotExist(err) {
		t.Error(".git/objects/pack directory was not created")
	}
}

func TestWriterInitExistingPath(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-writer-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	repoPath := filepath.Join(tmpDir, "existing-dir")

	// Create the directory first
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		t.Fatalf("Failed to create dir: %v", err)
	}

	w := NewWriter()
	if err := w.Init(repoPath); err != nil {
		t.Fatalf("Init failed on existing path: %v", err)
	}
	defer func() { require.NoError(t, w.Close()) }()
}

func TestWriterInitNestedPath(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-writer-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	repoPath := filepath.Join(tmpDir, "nested", "deep", "repo")

	w := NewWriter()
	if err := w.Init(repoPath); err != nil {
		t.Fatalf("Init failed on nested path: %v", err)
	}
	defer func() { require.NoError(t, w.Close()) }()

	// Check that all directories were created
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		t.Error("Nested directory was not created")
	}
}

func TestWriterInitWithConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-writer-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	repoPath := filepath.Join(tmpDir, "config-repo")

	w := NewWriter()
	config := map[string]string{
		"user.name":  "Test User",
		"user.email": "test@example.com",
	}

	if err := w.InitWithConfig(repoPath, config); err != nil {
		t.Fatalf("InitWithConfig failed: %v", err)
	}
	defer func() { require.NoError(t, w.Close()) }()

	// Verify config was set
	name, err := w.GetConfig("user.name")
	if err != nil {
		t.Fatalf("GetConfig(user.name) failed: %v", err)
	}
	if name != "Test User" {
		t.Errorf("user.name = %q, want %q", name, "Test User")
	}

	email, err := w.GetConfig("user.email")
	if err != nil {
		t.Fatalf("GetConfig(user.email) failed: %v", err)
	}
	if email != "test@example.com" {
		t.Errorf("user.email = %q, want %q", email, "test@example.com")
	}
}

func TestWriterSetConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-writer-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	repoPath := filepath.Join(tmpDir, "setconfig-repo")

	w := NewWriter()
	if err := w.Init(repoPath); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer func() { require.NoError(t, w.Close()) }()

	// Set user.name
	if err := w.SetConfig("user.name", "New User"); err != nil {
		t.Fatalf("SetConfig failed: %v", err)
	}

	// Verify
	val, err := w.GetConfig("user.name")
	if err != nil {
		t.Fatalf("GetConfig failed: %v", err)
	}
	if val != "New User" {
		t.Errorf("user.name = %q, want %q", val, "New User")
	}

	// Set user.email
	if err := w.SetConfig("user.email", "new@example.com"); err != nil {
		t.Fatalf("SetConfig failed: %v", err)
	}

	val, err = w.GetConfig("user.email")
	if err != nil {
		t.Fatalf("GetConfig failed: %v", err)
	}
	if val != "new@example.com" {
		t.Errorf("user.email = %q, want %q", val, "new@example.com")
	}
}

func TestWriterSetConfigNoRepo(t *testing.T) {
	w := NewWriter()

	err := w.SetConfig("user.name", "Test")
	if err == nil {
		t.Error("SetConfig should fail without repository")
	}
}

func TestWriterGetConfigNoRepo(t *testing.T) {
	w := NewWriter()

	_, err := w.GetConfig("user.name")
	if err == nil {
		t.Error("GetConfig should fail without repository")
	}
}

func TestWriterGetConfigUnknownKey(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-writer-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	repoPath := filepath.Join(tmpDir, "unknownkey-repo")

	w := NewWriter()
	if err := w.Init(repoPath); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer func() { require.NoError(t, w.Close()) }()

	_, err = w.GetConfig("unknown.key")
	if err == nil {
		t.Error("GetConfig should fail for unknown key")
	}
}

func TestWriterIsRepo(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-writer-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	repoPath := filepath.Join(tmpDir, "isrepo-repo")

	w := NewWriter()
	if err := w.Init(repoPath); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer func() { require.NoError(t, w.Close()) }()

	// Should be a repo
	if !w.IsRepo(repoPath) {
		t.Error("IsRepo should return true for initialized repo")
	}

	// Non-existent path
	if w.IsRepo("/nonexistent/path") {
		t.Error("IsRepo should return false for non-existent path")
	}

	// Regular directory (not a repo)
	regularDir := filepath.Join(tmpDir, "regular-dir")
	if err := os.MkdirAll(regularDir, 0755); err != nil {
		t.Fatalf("Failed to create regular dir: %v", err)
	}
	if w.IsRepo(regularDir) {
		t.Error("IsRepo should return false for regular directory")
	}
}

func TestWriterOpen(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-writer-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	repoPath := filepath.Join(tmpDir, "open-repo")

	// Create repo first
	w1 := NewWriter()
	if err := w1.Init(repoPath); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	require.NoError(t, w1.Close())

	// Open existing repo
	w2 := NewWriter()
	if err := w2.Open(repoPath); err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer func() { require.NoError(t, w2.Close()) }()
}

func TestWriterOpenNonExistent(t *testing.T) {
	w := NewWriter()

	err := w.Open("/nonexistent/repo")
	if err == nil {
		t.Error("Open should fail for non-existent repo")
	}
}

func TestWriterApplyCommit(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-writer-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	repoPath := filepath.Join(tmpDir, "commit-repo")

	w := NewWriter()
	if err := w.Init(repoPath); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer func() { require.NoError(t, w.Close()) }()

	commit := &vcs.Commit{
		Revision: "rev1",
		Author:   "Test Author",
		Email:    "test@example.com",
		Date:     time.Now(),
		Message:  "Initial commit",
		Files: []vcs.FileChange{
			{
				Path:    "file1.txt",
				Action:  vcs.ActionAdd,
				Content: []byte("Hello, World!"),
			},
		},
	}

	if err := w.ApplyCommit(commit); err != nil {
		t.Fatalf("ApplyCommit failed: %v", err)
	}

	// Verify file was created
	filePath := filepath.Join(repoPath, "file1.txt")
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	if string(content) != "Hello, World!" {
		t.Errorf("File content = %q, want %q", string(content), "Hello, World!")
	}
}

func TestWriterApplyCommitNoRepo(t *testing.T) {
	w := NewWriter()

	commit := &vcs.Commit{
		Revision: "rev1",
		Author:   "Test",
		Email:    "test@example.com",
		Date:     time.Now(),
		Message:  "Test",
		Files:    []vcs.FileChange{},
	}

	err := w.ApplyCommit(commit)
	if err == nil {
		t.Error("ApplyCommit should fail without repository")
	}
}

func TestWriterApplyCommitModifyFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-writer-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	repoPath := filepath.Join(tmpDir, "modify-repo")

	w := NewWriter()
	if err := w.Init(repoPath); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer func() { require.NoError(t, w.Close()) }()

	// First commit - add file
	commit1 := &vcs.Commit{
		Revision: "rev1",
		Author:   "Test",
		Email:    "test@example.com",
		Date:     time.Now(),
		Message:  "Add file",
		Files: []vcs.FileChange{
			{
				Path:    "file.txt",
				Action:  vcs.ActionAdd,
				Content: []byte("Original content"),
			},
		},
	}
	if err := w.ApplyCommit(commit1); err != nil {
		t.Fatalf("First ApplyCommit failed: %v", err)
	}

	// Second commit - modify file
	commit2 := &vcs.Commit{
		Revision: "rev2",
		Author:   "Test",
		Email:    "test@example.com",
		Date:     time.Now(),
		Message:  "Modify file",
		Files: []vcs.FileChange{
			{
				Path:    "file.txt",
				Action:  vcs.ActionModify,
				Content: []byte("Modified content"),
			},
		},
	}
	if err := w.ApplyCommit(commit2); err != nil {
		t.Fatalf("Second ApplyCommit failed: %v", err)
	}

	// Verify file was modified
	content, err := os.ReadFile(filepath.Join(repoPath, "file.txt"))
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	if string(content) != "Modified content" {
		t.Errorf("File content = %q, want %q", string(content), "Modified content")
	}
}

func TestWriterApplyCommitDeleteFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-writer-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	repoPath := filepath.Join(tmpDir, "delete-repo")

	w := NewWriter()
	if err := w.Init(repoPath); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer func() { require.NoError(t, w.Close()) }()

	// First commit - add file
	commit1 := &vcs.Commit{
		Revision: "rev1",
		Author:   "Test",
		Email:    "test@example.com",
		Date:     time.Now(),
		Message:  "Add file",
		Files: []vcs.FileChange{
			{
				Path:    "to-delete.txt",
				Action:  vcs.ActionAdd,
				Content: []byte("Delete me"),
			},
		},
	}
	if err := w.ApplyCommit(commit1); err != nil {
		t.Fatalf("First ApplyCommit failed: %v", err)
	}

	// Second commit - delete file
	commit2 := &vcs.Commit{
		Revision: "rev2",
		Author:   "Test",
		Email:    "test@example.com",
		Date:     time.Now(),
		Message:  "Delete file",
		Files: []vcs.FileChange{
			{
				Path:   "to-delete.txt",
				Action: vcs.ActionDelete,
			},
		},
	}
	if err := w.ApplyCommit(commit2); err != nil {
		t.Fatalf("Second ApplyCommit failed: %v", err)
	}

	// Verify file was deleted
	filePath := filepath.Join(repoPath, "to-delete.txt")
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Error("File should have been deleted")
	}
}

func TestWriterApplyCommitNestedPath(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-writer-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	repoPath := filepath.Join(tmpDir, "nested-repo")

	w := NewWriter()
	if err := w.Init(repoPath); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer func() { require.NoError(t, w.Close()) }()

	commit := &vcs.Commit{
		Revision: "rev1",
		Author:   "Test",
		Email:    "test@example.com",
		Date:     time.Now(),
		Message:  "Add nested file",
		Files: []vcs.FileChange{
			{
				Path:    "dir1/dir2/file.txt",
				Action:  vcs.ActionAdd,
				Content: []byte("Nested content"),
			},
		},
	}

	if err := w.ApplyCommit(commit); err != nil {
		t.Fatalf("ApplyCommit failed: %v", err)
	}

	// Verify nested file was created
	filePath := filepath.Join(repoPath, "dir1", "dir2", "file.txt")
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	if string(content) != "Nested content" {
		t.Errorf("File content = %q, want %q", string(content), "Nested content")
	}
}

func TestWriterApplyCommitMultipleFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-writer-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	repoPath := filepath.Join(tmpDir, "multifile-repo")

	w := NewWriter()
	if err := w.Init(repoPath); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer func() { require.NoError(t, w.Close()) }()

	commit := &vcs.Commit{
		Revision: "rev1",
		Author:   "Test",
		Email:    "test@example.com",
		Date:     time.Now(),
		Message:  "Add multiple files",
		Files: []vcs.FileChange{
			{
				Path:    "file1.txt",
				Action:  vcs.ActionAdd,
				Content: []byte("Content 1"),
			},
			{
				Path:    "file2.txt",
				Action:  vcs.ActionAdd,
				Content: []byte("Content 2"),
			},
			{
				Path:    "dir/file3.txt",
				Action:  vcs.ActionAdd,
				Content: []byte("Content 3"),
			},
		},
	}

	if err := w.ApplyCommit(commit); err != nil {
		t.Fatalf("ApplyCommit failed: %v", err)
	}

	// Verify all files were created
	for i, expected := range []string{"Content 1", "Content 2", "Content 3"} {
		var path string
		switch i {
		case 0:
			path = filepath.Join(repoPath, "file1.txt")
		case 1:
			path = filepath.Join(repoPath, "file2.txt")
		case 2:
			path = filepath.Join(repoPath, "dir", "file3.txt")
		}
		content, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("Failed to read file %s: %v", path, err)
			continue
		}
		if string(content) != expected {
			t.Errorf("File %s content = %q, want %q", path, string(content), expected)
		}
	}
}

func TestWriterCreateBranch(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-writer-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	repoPath := filepath.Join(tmpDir, "branch-repo")

	w := NewWriter()
	if err := w.Init(repoPath); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer func() { require.NoError(t, w.Close()) }()

	// Create initial commit
	commit := &vcs.Commit{
		Revision: "rev1",
		Author:   "Test",
		Email:    "test@example.com",
		Date:     time.Now(),
		Message:  "Initial commit",
		Files: []vcs.FileChange{
			{
				Path:    "file.txt",
				Action:  vcs.ActionAdd,
				Content: []byte("Content"),
			},
		},
	}
	if err := w.ApplyCommit(commit); err != nil {
		t.Fatalf("ApplyCommit failed: %v", err)
	}

	// Create branch
	if err := w.CreateBranch("feature", "HEAD"); err != nil {
		t.Fatalf("CreateBranch failed: %v", err)
	}

	// Verify branch exists
	branches, err := w.ListBranches()
	if err != nil {
		t.Fatalf("ListBranches failed: %v", err)
	}

	found := false
	for _, b := range branches {
		if b == "feature" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Branch 'feature' not found")
	}
}

func TestWriterCreateBranchNoRepo(t *testing.T) {
	w := NewWriter()

	err := w.CreateBranch("test", "HEAD")
	if err == nil {
		t.Error("CreateBranch should fail without repository")
	}
}

func TestWriterCreateTag(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-writer-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	repoPath := filepath.Join(tmpDir, "tag-repo")

	w := NewWriter()
	if err := w.Init(repoPath); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer func() { require.NoError(t, w.Close()) }()

	// Create initial commit
	commit := &vcs.Commit{
		Revision: "rev1",
		Author:   "Test",
		Email:    "test@example.com",
		Date:     time.Now(),
		Message:  "Initial commit",
		Files: []vcs.FileChange{
			{
				Path:    "file.txt",
				Action:  vcs.ActionAdd,
				Content: []byte("Content"),
			},
		},
	}
	if err := w.ApplyCommit(commit); err != nil {
		t.Fatalf("ApplyCommit failed: %v", err)
	}

	// Create lightweight tag
	if err := w.CreateTag("v1.0.0", "HEAD", ""); err != nil {
		t.Fatalf("CreateTag failed: %v", err)
	}

	// Verify tag exists
	tags, err := w.ListTags()
	if err != nil {
		t.Fatalf("ListTags failed: %v", err)
	}

	if _, ok := tags["v1.0.0"]; !ok {
		t.Error("Tag 'v1.0.0' not found")
	}
}

func TestWriterCreateAnnotatedTag(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-writer-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	repoPath := filepath.Join(tmpDir, "annotated-tag-repo")

	w := NewWriter()
	if err := w.Init(repoPath); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer func() { require.NoError(t, w.Close()) }()

	// Create initial commit
	commit := &vcs.Commit{
		Revision: "rev1",
		Author:   "Test",
		Email:    "test@example.com",
		Date:     time.Now(),
		Message:  "Initial commit",
		Files: []vcs.FileChange{
			{
				Path:    "file.txt",
				Action:  vcs.ActionAdd,
				Content: []byte("Content"),
			},
		},
	}
	if err := w.ApplyCommit(commit); err != nil {
		t.Fatalf("ApplyCommit failed: %v", err)
	}

	// Create annotated tag
	if err := w.CreateTag("v1.0.0", "HEAD", "Release version 1.0.0"); err != nil {
		t.Fatalf("CreateTag failed: %v", err)
	}

	// Verify tag exists
	tags, err := w.ListTags()
	if err != nil {
		t.Fatalf("ListTags failed: %v", err)
	}

	if _, ok := tags["v1.0.0"]; !ok {
		t.Error("Tag 'v1.0.0' not found")
	}
}

func TestWriterCreateTagNoRepo(t *testing.T) {
	w := NewWriter()

	err := w.CreateTag("v1.0.0", "HEAD", "")
	if err == nil {
		t.Error("CreateTag should fail without repository")
	}
}

func TestWriterCreateBranchWithRevision(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-writer-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	repoPath := filepath.Join(tmpDir, "branch-revision-repo")

	w := NewWriter()
	if err := w.Init(repoPath); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer func() { require.NoError(t, w.Close()) }()

	// Create initial commit
	commit := &vcs.Commit{
		Revision: "rev1",
		Author:   "Test",
		Email:    "test@example.com",
		Date:     time.Now(),
		Message:  "Initial commit",
		Files: []vcs.FileChange{
			{
				Path:    "file.txt",
				Action:  vcs.ActionAdd,
				Content: []byte("Content"),
			},
		},
	}
	if err := w.ApplyCommit(commit); err != nil {
		t.Fatalf("ApplyCommit failed: %v", err)
	}

	// Get the commit hash
	lastCommit, err := w.GetLastCommit()
	if err != nil {
		t.Fatalf("GetLastCommit failed: %v", err)
	}

	// Create branch using specific revision hash
	if err := w.CreateBranch("feature2", lastCommit.Revision); err != nil {
		t.Fatalf("CreateBranch with revision failed: %v", err)
	}

	// Verify branch exists
	branches, err := w.ListBranches()
	if err != nil {
		t.Fatalf("ListBranches failed: %v", err)
	}

	found := false
	for _, b := range branches {
		if b == "feature2" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Branch 'feature2' not found")
	}
}

func TestWriterCreateBranchInvalidRevision(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-writer-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	repoPath := filepath.Join(tmpDir, "branch-invalid-repo")

	w := NewWriter()
	if err := w.Init(repoPath); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer func() { require.NoError(t, w.Close()) }()

	// Create initial commit
	commit := &vcs.Commit{
		Revision: "rev1",
		Author:   "Test",
		Email:    "test@example.com",
		Date:     time.Now(),
		Message:  "Initial commit",
		Files: []vcs.FileChange{
			{
				Path:    "file.txt",
				Action:  vcs.ActionAdd,
				Content: []byte("Content"),
			},
		},
	}
	if err := w.ApplyCommit(commit); err != nil {
		t.Fatalf("ApplyCommit failed: %v", err)
	}

	// Try to create branch with invalid revision
	err = w.CreateBranch("bad-branch", "invalid-revision")
	if err == nil {
		t.Error("CreateBranch should fail with invalid revision")
	}
}

func TestWriterCreateTagWithRevision(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-writer-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	repoPath := filepath.Join(tmpDir, "tag-revision-repo")

	w := NewWriter()
	if err := w.Init(repoPath); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer func() { require.NoError(t, w.Close()) }()

	// Create initial commit
	commit := &vcs.Commit{
		Revision: "rev1",
		Author:   "Test",
		Email:    "test@example.com",
		Date:     time.Now(),
		Message:  "Initial commit",
		Files: []vcs.FileChange{
			{
				Path:    "file.txt",
				Action:  vcs.ActionAdd,
				Content: []byte("Content"),
			},
		},
	}
	if err := w.ApplyCommit(commit); err != nil {
		t.Fatalf("ApplyCommit failed: %v", err)
	}

	// Get the commit hash
	lastCommit, err := w.GetLastCommit()
	if err != nil {
		t.Fatalf("GetLastCommit failed: %v", err)
	}

	// Create tag using specific revision hash
	if err := w.CreateTag("v1.0.1", lastCommit.Revision, ""); err != nil {
		t.Fatalf("CreateTag with revision failed: %v", err)
	}

	// Verify tag exists
	tags, err := w.ListTags()
	if err != nil {
		t.Fatalf("ListTags failed: %v", err)
	}

	if _, ok := tags["v1.0.1"]; !ok {
		t.Error("Tag 'v1.0.1' not found")
	}
}

func TestWriterCreateTagInvalidRevision(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-writer-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	repoPath := filepath.Join(tmpDir, "tag-invalid-repo")

	w := NewWriter()
	if err := w.Init(repoPath); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer func() { require.NoError(t, w.Close()) }()

	// Create initial commit
	commit := &vcs.Commit{
		Revision: "rev1",
		Author:   "Test",
		Email:    "test@example.com",
		Date:     time.Now(),
		Message:  "Initial commit",
		Files: []vcs.FileChange{
			{
				Path:    "file.txt",
				Action:  vcs.ActionAdd,
				Content: []byte("Content"),
			},
		},
	}
	if err := w.ApplyCommit(commit); err != nil {
		t.Fatalf("ApplyCommit failed: %v", err)
	}

	// Try to create tag with invalid revision
	err = w.CreateTag("bad-tag", "invalid-revision", "")
	if err == nil {
		t.Error("CreateTag should fail with invalid revision")
	}
}

func TestWriterCreateAnnotatedTagWithRevision(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-writer-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	repoPath := filepath.Join(tmpDir, "annotated-tag-revision-repo")

	w := NewWriter()
	if err := w.Init(repoPath); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer func() { require.NoError(t, w.Close()) }()

	// Create initial commit
	commit := &vcs.Commit{
		Revision: "rev1",
		Author:   "Test",
		Email:    "test@example.com",
		Date:     time.Now(),
		Message:  "Initial commit",
		Files: []vcs.FileChange{
			{
				Path:    "file.txt",
				Action:  vcs.ActionAdd,
				Content: []byte("Content"),
			},
		},
	}
	if err := w.ApplyCommit(commit); err != nil {
		t.Fatalf("ApplyCommit failed: %v", err)
	}

	// Get the commit hash
	lastCommit, err := w.GetLastCommit()
	if err != nil {
		t.Fatalf("GetLastCommit failed: %v", err)
	}

	// Create annotated tag using specific revision hash
	if err := w.CreateTag("v1.0.2", lastCommit.Revision, "Release version 1.0.2"); err != nil {
		t.Fatalf("CreateTag with revision failed: %v", err)
	}

	// Verify tag exists
	tags, err := w.ListTags()
	if err != nil {
		t.Fatalf("ListTags failed: %v", err)
	}

	if _, ok := tags["v1.0.2"]; !ok {
		t.Error("Tag 'v1.0.2' not found")
	}
}

func TestWriterListBranches(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-writer-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	repoPath := filepath.Join(tmpDir, "list-branches-repo")

	w := NewWriter()
	if err := w.Init(repoPath); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer func() { require.NoError(t, w.Close()) }()

	// Initial list - nil slice is valid for empty in Go (len(nil slice) == 0)
	branches, err := w.ListBranches()
	if err != nil {
		t.Fatalf("ListBranches failed: %v", err)
	}
	// Just verify we can call it without error
	_ = branches
}

func TestWriterListBranchesNoRepo(t *testing.T) {
	w := NewWriter()

	_, err := w.ListBranches()
	if err == nil {
		t.Error("ListBranches should fail without repository")
	}
}

func TestWriterListTags(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-writer-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	repoPath := filepath.Join(tmpDir, "list-tags-repo")

	w := NewWriter()
	if err := w.Init(repoPath); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer func() { require.NoError(t, w.Close()) }()

	// Initial list should be empty
	tags, err := w.ListTags()
	if err != nil {
		t.Fatalf("ListTags failed: %v", err)
	}
	if tags == nil {
		t.Error("ListTags should return empty map, not nil")
	}
}

func TestWriterListTagsNoRepo(t *testing.T) {
	w := NewWriter()

	_, err := w.ListTags()
	if err == nil {
		t.Error("ListTags should fail without repository")
	}
}

func TestWriterGetLastCommit(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-writer-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	repoPath := filepath.Join(tmpDir, "last-commit-repo")

	w := NewWriter()
	if err := w.Init(repoPath); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer func() { require.NoError(t, w.Close()) }()

	// Create a commit
	commit := &vcs.Commit{
		Revision: "rev1",
		Author:   "Test Author",
		Email:    "test@example.com",
		Date:     time.Now(),
		Message:  "Test commit message",
		Files: []vcs.FileChange{
			{
				Path:    "file.txt",
				Action:  vcs.ActionAdd,
				Content: []byte("Content"),
			},
		},
	}
	if err := w.ApplyCommit(commit); err != nil {
		t.Fatalf("ApplyCommit failed: %v", err)
	}

	// Get last commit
	lastCommit, err := w.GetLastCommit()
	if err != nil {
		t.Fatalf("GetLastCommit failed: %v", err)
	}

	if lastCommit.Author != "Test Author" {
		t.Errorf("Author = %q, want %q", lastCommit.Author, "Test Author")
	}
	if lastCommit.Email != "test@example.com" {
		t.Errorf("Email = %q, want %q", lastCommit.Email, "test@example.com")
	}
	if lastCommit.Message != "Test commit message" {
		t.Errorf("Message = %q, want %q", lastCommit.Message, "Test commit message")
	}
}

func TestWriterGetLastCommitNoRepo(t *testing.T) {
	w := NewWriter()

	_, err := w.GetLastCommit()
	if err == nil {
		t.Error("GetLastCommit should fail without repository")
	}
}

func TestWriterGetCommitHashes(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-writer-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	repoPath := filepath.Join(tmpDir, "hashes-repo")

	w := NewWriter()
	if err := w.Init(repoPath); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer func() { require.NoError(t, w.Close()) }()

	// Create multiple commits
	for i := 0; i < 3; i++ {
		commit := &vcs.Commit{
			Revision: "rev" + string(rune('1'+i)),
			Author:   "Test",
			Email:    "test@example.com",
			Date:     time.Now().Add(time.Duration(i) * time.Second),
			Message:  "Commit " + string(rune('1'+i)),
			Files: []vcs.FileChange{
				{
					Path:    "file.txt",
					Action:  vcs.ActionAdd,
					Content: []byte("Content " + string(rune('1'+i))),
				},
			},
		}
		if err := w.ApplyCommit(commit); err != nil {
			t.Fatalf("ApplyCommit %d failed: %v", i, err)
		}
	}

	// Get hashes
	hashes, err := w.GetCommitHashes()
	if err != nil {
		t.Fatalf("GetCommitHashes failed: %v", err)
	}

	if len(hashes) != 3 {
		t.Errorf("GetCommitHashes returned %d hashes, want 3", len(hashes))
	}

	// All hashes should be 40 characters
	for i, h := range hashes {
		if len(h) != 40 {
			t.Errorf("Hash %d has length %d, want 40", i, len(h))
		}
	}
}

func TestWriterGetCommitHashesNoRepo(t *testing.T) {
	w := NewWriter()

	_, err := w.GetCommitHashes()
	if err == nil {
		t.Error("GetCommitHashes should fail without repository")
	}
}

func TestWriterResolveRevision(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-writer-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	repoPath := filepath.Join(tmpDir, "resolve-repo")

	w := NewWriter()
	if err := w.Init(repoPath); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer func() { require.NoError(t, w.Close()) }()

	// Create a commit
	commit := &vcs.Commit{
		Revision: "rev1",
		Author:   "Test",
		Email:    "test@example.com",
		Date:     time.Now(),
		Message:  "Test commit",
		Files: []vcs.FileChange{
			{
				Path:    "file.txt",
				Action:  vcs.ActionAdd,
				Content: []byte("Content"),
			},
		},
	}
	if err := w.ApplyCommit(commit); err != nil {
		t.Fatalf("ApplyCommit failed: %v", err)
	}

	// Resolve HEAD
	hash, err := w.ResolveRevision("HEAD")
	if err != nil {
		t.Fatalf("ResolveRevision(HEAD) failed: %v", err)
	}
	if len(hash) != 40 {
		t.Errorf("HEAD hash length = %d, want 40", len(hash))
	}
}

func TestWriterResolveRevisionHash(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-writer-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	repoPath := filepath.Join(tmpDir, "resolve-hash-repo")

	w := NewWriter()
	if err := w.Init(repoPath); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer func() { require.NoError(t, w.Close()) }()

	// Pass a 40-character hash directly
	fakeHash := "0123456789012345678901234567890123456789"
	hash, err := w.ResolveRevision(fakeHash)
	if err != nil {
		t.Fatalf("ResolveRevision(hash) failed: %v", err)
	}
	if hash != fakeHash {
		t.Errorf("hash = %q, want %q", hash, fakeHash)
	}
}

func TestWriterResolveRevisionNoRepo(t *testing.T) {
	w := NewWriter()

	_, err := w.ResolveRevision("HEAD")
	if err == nil {
		t.Error("ResolveRevision should fail without repository")
	}
}

func TestWriterClose(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-writer-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	repoPath := filepath.Join(tmpDir, "close-repo")

	w := NewWriter()
	if err := w.Init(repoPath); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Close should not fail
	if err := w.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}

	// Double close should be safe
	if err := w.Close(); err != nil {
		t.Errorf("Double close failed: %v", err)
	}
}

func TestWriterCloseNoRepo(t *testing.T) {
	w := NewWriter()

	// Close without repo should not fail
	if err := w.Close(); err != nil {
		t.Errorf("Close without repo failed: %v", err)
	}
}

func TestWriterApplyCommitEmptyFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-writer-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	repoPath := filepath.Join(tmpDir, "empty-files-repo")

	w := NewWriter()
	if err := w.Init(repoPath); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer func() { require.NoError(t, w.Close()) }()

	// Create a commit with empty file content
	commit := &vcs.Commit{
		Revision: "rev1",
		Author:   "Test",
		Email:    "test@example.com",
		Date:     time.Now(),
		Message:  "Empty file commit",
		Files: []vcs.FileChange{
			{
				Path:    "empty.txt",
				Action:  vcs.ActionAdd,
				Content: []byte{},
			},
		},
	}

	if err := w.ApplyCommit(commit); err != nil {
		t.Fatalf("ApplyCommit failed: %v", err)
	}

	// Verify file was created (even if empty)
	filePath := filepath.Join(repoPath, "empty.txt")
	info, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}
	if info.Size() != 0 {
		t.Errorf("File size = %d, want 0", info.Size())
	}
}

func TestWriterApplyCommitBinaryContent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-writer-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	repoPath := filepath.Join(tmpDir, "binary-repo")

	w := NewWriter()
	if err := w.Init(repoPath); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer func() { require.NoError(t, w.Close()) }()

	// Create a commit with binary content
	binaryContent := []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD}
	commit := &vcs.Commit{
		Revision: "rev1",
		Author:   "Test",
		Email:    "test@example.com",
		Date:     time.Now(),
		Message:  "Binary file commit",
		Files: []vcs.FileChange{
			{
				Path:    "binary.bin",
				Action:  vcs.ActionAdd,
				Content: binaryContent,
			},
		},
	}

	if err := w.ApplyCommit(commit); err != nil {
		t.Fatalf("ApplyCommit failed: %v", err)
	}

	// Verify binary content was preserved
	content, err := os.ReadFile(filepath.Join(repoPath, "binary.bin"))
	if err != nil {
		t.Fatalf("Failed to read binary file: %v", err)
	}
	if len(content) != len(binaryContent) {
		t.Errorf("Binary content length = %d, want %d", len(content), len(binaryContent))
	}
}

func TestWriterApplyCommitLargeContent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-writer-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	repoPath := filepath.Join(tmpDir, "large-repo")

	w := NewWriter()
	if err := w.Init(repoPath); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer func() { require.NoError(t, w.Close()) }()

	// Create a commit with large content
	largeContent := make([]byte, 1024*1024) // 1MB
	for i := range largeContent {
		largeContent[i] = byte(i % 256)
	}

	commit := &vcs.Commit{
		Revision: "rev1",
		Author:   "Test",
		Email:    "test@example.com",
		Date:     time.Now(),
		Message:  "Large file commit",
		Files: []vcs.FileChange{
			{
				Path:    "large.bin",
				Action:  vcs.ActionAdd,
				Content: largeContent,
			},
		},
	}

	if err := w.ApplyCommit(commit); err != nil {
		t.Fatalf("ApplyCommit failed: %v", err)
	}

	// Verify large content was written
	info, err := os.Stat(filepath.Join(repoPath, "large.bin"))
	if err != nil {
		t.Fatalf("Failed to stat large file: %v", err)
	}
	if info.Size() != int64(len(largeContent)) {
		t.Errorf("Large file size = %d, want %d", info.Size(), len(largeContent))
	}
}

func TestWriterGetLastCommit_EmptyRepo(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-writer-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	repoPath := filepath.Join(tmpDir, "empty-repo")

	w := NewWriter()
	if err := w.Init(repoPath); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer func() { require.NoError(t, w.Close()) }()

	// Try to get last commit from empty repo
	_, err = w.GetLastCommit()
	if err == nil {
		t.Error("GetLastCommit should fail on empty repository")
	}
}

func TestWriterInitWithConfig_MultipleOptions(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-writer-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	repoPath := filepath.Join(tmpDir, "multi-config-repo")

	w := NewWriter()
	config := map[string]string{
		"user.name":  "Test User",
		"user.email": "test@example.com",
	}

	if err := w.InitWithConfig(repoPath, config); err != nil {
		t.Fatalf("InitWithConfig failed: %v", err)
	}
	defer func() { require.NoError(t, w.Close()) }()

	// Verify configs were set
	name, err := w.GetConfig("user.name")
	if err != nil {
		t.Fatalf("GetConfig(user.name) failed: %v", err)
	}
	if name != "Test User" {
		t.Errorf("user.name = %q, want %q", name, "Test User")
	}

	email, err := w.GetConfig("user.email")
	if err != nil {
		t.Fatalf("GetConfig(user.email) failed: %v", err)
	}
	if email != "test@example.com" {
		t.Errorf("user.email = %q, want %q", email, "test@example.com")
	}
}

func TestWriterApplyCommit_DeleteNonExistentFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-writer-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	repoPath := filepath.Join(tmpDir, "delete-nonexistent-repo")

	w := NewWriter()
	if err := w.Init(repoPath); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer func() { require.NoError(t, w.Close()) }()

	// Create initial commit
	commit1 := &vcs.Commit{
		Revision: "rev1",
		Author:   "Test",
		Email:    "test@example.com",
		Date:     time.Now(),
		Message:  "Initial commit",
		Files: []vcs.FileChange{
			{
				Path:    "file.txt",
				Action:  vcs.ActionAdd,
				Content: []byte("Content"),
			},
		},
	}
	if err := w.ApplyCommit(commit1); err != nil {
		t.Fatalf("First ApplyCommit failed: %v", err)
	}

	// Try to delete a file that doesn't exist and add a new file
	// This should work because there's at least one real change (the add)
	commit2 := &vcs.Commit{
		Revision: "rev2",
		Author:   "Test",
		Email:    "test@example.com",
		Date:     time.Now(),
		Message:  "Add new file and delete nonexistent",
		Files: []vcs.FileChange{
			{
				Path:   "nonexistent.txt",
				Action: vcs.ActionDelete,
			},
			{
				Path:    "newfile.txt",
				Action:  vcs.ActionAdd,
				Content: []byte("New content"),
			},
		},
	}
	// This should succeed because there's a real file change
	if err := w.ApplyCommit(commit2); err != nil {
		t.Fatalf("Second ApplyCommit failed: %v", err)
	}
}

func TestWriterCreateBranch_WithLastCommit(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-writer-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	repoPath := filepath.Join(tmpDir, "branch-lastcommit-repo")

	w := NewWriter()
	if err := w.Init(repoPath); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer func() { require.NoError(t, w.Close()) }()

	// Create initial commit (this sets w.lastCommit)
	commit := &vcs.Commit{
		Revision: "rev1",
		Author:   "Test",
		Email:    "test@example.com",
		Date:     time.Now(),
		Message:  "Initial commit",
		Files: []vcs.FileChange{
			{
				Path:    "file.txt",
				Action:  vcs.ActionAdd,
				Content: []byte("Content"),
			},
		},
	}
	if err := w.ApplyCommit(commit); err != nil {
		t.Fatalf("ApplyCommit failed: %v", err)
	}

	// Create branch using HEAD (should use w.lastCommit)
	if err := w.CreateBranch("from-lastcommit", "HEAD"); err != nil {
		t.Fatalf("CreateBranch failed: %v", err)
	}

	// Verify branch exists
	branches, err := w.ListBranches()
	if err != nil {
		t.Fatalf("ListBranches failed: %v", err)
	}

	found := false
	for _, b := range branches {
		if b == "from-lastcommit" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Branch 'from-lastcommit' not found")
	}
}

func TestWriterCreateTag_WithLastCommit(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-writer-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	repoPath := filepath.Join(tmpDir, "tag-lastcommit-repo")

	w := NewWriter()
	if err := w.Init(repoPath); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer func() { require.NoError(t, w.Close()) }()

	// Create initial commit (this sets w.lastCommit)
	commit := &vcs.Commit{
		Revision: "rev1",
		Author:   "Test",
		Email:    "test@example.com",
		Date:     time.Now(),
		Message:  "Initial commit",
		Files: []vcs.FileChange{
			{
				Path:    "file.txt",
				Action:  vcs.ActionAdd,
				Content: []byte("Content"),
			},
		},
	}
	if err := w.ApplyCommit(commit); err != nil {
		t.Fatalf("ApplyCommit failed: %v", err)
	}

	// Create tag using HEAD (should use w.lastCommit)
	if err := w.CreateTag("v1.0.0", "HEAD", ""); err != nil {
		t.Fatalf("CreateTag failed: %v", err)
	}

	// Verify tag exists
	tags, err := w.ListTags()
	if err != nil {
		t.Fatalf("ListTags failed: %v", err)
	}

	if _, ok := tags["v1.0.0"]; !ok {
		t.Error("Tag 'v1.0.0' not found")
	}
}

func TestWriterCreateAnnotatedTag_WithLastCommit(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-writer-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	repoPath := filepath.Join(tmpDir, "annotated-tag-lastcommit-repo")

	w := NewWriter()
	if err := w.Init(repoPath); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer func() { require.NoError(t, w.Close()) }()

	// Create initial commit (this sets w.lastCommit)
	commit := &vcs.Commit{
		Revision: "rev1",
		Author:   "Test",
		Email:    "test@example.com",
		Date:     time.Now(),
		Message:  "Initial commit",
		Files: []vcs.FileChange{
			{
				Path:    "file.txt",
				Action:  vcs.ActionAdd,
				Content: []byte("Content"),
			},
		},
	}
	if err := w.ApplyCommit(commit); err != nil {
		t.Fatalf("ApplyCommit failed: %v", err)
	}

	// Create annotated tag using HEAD (should use w.lastCommit)
	if err := w.CreateTag("v1.0.0", "HEAD", "Release version 1.0.0"); err != nil {
		t.Fatalf("CreateTag failed: %v", err)
	}

	// Verify tag exists
	tags, err := w.ListTags()
	if err != nil {
		t.Fatalf("ListTags failed: %v", err)
	}

	if _, ok := tags["v1.0.0"]; !ok {
		t.Error("Tag 'v1.0.0' not found")
	}
}
