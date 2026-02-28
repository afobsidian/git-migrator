package git

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/adamf123git/git-migrator/internal/vcs"
	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/require"
)

// createTestRepo initialises a minimal Git repository with the provided
// commits and returns the repository path.
func createTestRepo(t *testing.T, commits []struct {
	file    string
	content string
	message string
}) string {
	t.Helper()
	dir := t.TempDir()

	repo, err := gogit.PlainInit(dir, false)
	require.NoError(t, err)

	w, err := repo.Worktree()
	require.NoError(t, err)

	for _, c := range commits {
		full := filepath.Join(dir, c.file)
		require.NoError(t, os.MkdirAll(filepath.Dir(full), 0755))
		require.NoError(t, os.WriteFile(full, []byte(c.content), 0644))
		_, err = w.Add(c.file)
		require.NoError(t, err)
		_, err = w.Commit(c.message, &gogit.CommitOptions{
			Author: &object.Signature{
				Name:  "Test Author",
				Email: "test@example.com",
				When:  time.Now(),
			},
		})
		require.NoError(t, err)
	}

	return dir
}

func TestNewGitReader(t *testing.T) {
	r := NewReader("/some/path")
	if r == nil {
		t.Fatal("NewReader returned nil")
	}
	if r.path != "/some/path" {
		t.Errorf("path = %q, want %q", r.path, "/some/path")
	}
}

func TestGitReaderValidate_Valid(t *testing.T) {
	dir := createTestRepo(t, []struct {
		file    string
		content string
		message string
	}{
		{"README.md", "hello", "initial commit"},
	})

	r := NewReader(dir)
	if err := r.Validate(); err != nil {
		t.Errorf("Validate() error = %v", err)
	}
}

func TestGitReaderValidate_Invalid(t *testing.T) {
	r := NewReader("/nonexistent/path/12345")
	if err := r.Validate(); err == nil {
		t.Error("Validate() should fail on non-existent path")
	}
}

func TestGitReaderGetCommits(t *testing.T) {
	dir := createTestRepo(t, []struct {
		file    string
		content string
		message string
	}{
		{"a.txt", "a", "first commit"},
		{"b.txt", "b", "second commit"},
		{"c.txt", "c", "third commit"},
	})

	r := NewReader(dir)
	iter, err := r.GetCommits()
	require.NoError(t, err)

	var commits []*vcs.Commit
	for iter.Next() {
		commits = append(commits, iter.Commit())
	}
	require.NoError(t, iter.Err())

	if len(commits) != 3 {
		t.Fatalf("got %d commits, want 3", len(commits))
	}

	// Oldest first
	if commits[0].Message != "first commit\n" && commits[0].Message != "first commit" {
		t.Errorf("first commit message = %q", commits[0].Message)
	}
}

func TestGitReaderGetCommits_Empty(t *testing.T) {
	r := NewReader("/nonexistent/path/12345")
	_, err := r.GetCommits()
	if err == nil {
		t.Error("GetCommits() should fail on non-existent path")
	}
}

func TestGitReaderGetCommitsSince(t *testing.T) {
	dir := createTestRepo(t, []struct {
		file    string
		content string
		message string
	}{
		{"a.txt", "a", "commit A"},
		{"b.txt", "b", "commit B"},
		{"c.txt", "c", "commit C"},
	})

	r := NewReader(dir)

	// Get all commits first to obtain a revision hash
	allIter, err := r.GetCommits()
	require.NoError(t, err)

	var all []*vcs.Commit
	for allIter.Next() {
		all = append(all, allIter.Commit())
	}
	require.NoError(t, allIter.Err())
	require.Len(t, all, 3)

	// Commits since the first one – should return the last two
	sinceIter, err := r.GetCommitsSince(all[0].Revision)
	require.NoError(t, err)

	var since []*vcs.Commit
	for sinceIter.Next() {
		since = append(since, sinceIter.Commit())
	}
	require.NoError(t, sinceIter.Err())

	if len(since) != 2 {
		t.Fatalf("GetCommitsSince returned %d commits, want 2", len(since))
	}
}

func TestGitReaderGetCommitsSince_Empty(t *testing.T) {
	dir := createTestRepo(t, []struct {
		file    string
		content string
		message string
	}{
		{"a.txt", "a", "only commit"},
	})

	r := NewReader(dir)
	allIter, err := r.GetCommits()
	require.NoError(t, err)

	var all []*vcs.Commit
	for allIter.Next() {
		all = append(all, allIter.Commit())
	}
	require.Len(t, all, 1)

	// Commits since the only commit → should return zero
	sinceIter, err := r.GetCommitsSince(all[0].Revision)
	require.NoError(t, err)

	var since []*vcs.Commit
	for sinceIter.Next() {
		since = append(since, sinceIter.Commit())
	}
	if len(since) != 0 {
		t.Errorf("GetCommitsSince after last commit returned %d commits, want 0", len(since))
	}
}

func TestGitReaderGetCommitsSince_UnknownRevision(t *testing.T) {
	dir := createTestRepo(t, []struct {
		file    string
		content string
		message string
	}{
		{"a.txt", "a", "commit A"},
		{"b.txt", "b", "commit B"},
	})

	r := NewReader(dir)

	// Unknown revision → should return all commits
	sinceIter, err := r.GetCommitsSince("0000000000000000000000000000000000000000")
	require.NoError(t, err)

	var since []*vcs.Commit
	for sinceIter.Next() {
		since = append(since, sinceIter.Commit())
	}
	if len(since) != 2 {
		t.Errorf("GetCommitsSince unknown rev returned %d commits, want 2", len(since))
	}
}

func TestGitReaderGetBranches(t *testing.T) {
	dir := createTestRepo(t, []struct {
		file    string
		content string
		message string
	}{
		{"a.txt", "a", "initial"},
	})

	r := NewReader(dir)
	branches, err := r.GetBranches()
	require.NoError(t, err)

	if len(branches) == 0 {
		t.Error("expected at least one branch")
	}
}

func TestGitReaderGetTags(t *testing.T) {
	dir := createTestRepo(t, []struct {
		file    string
		content string
		message string
	}{
		{"a.txt", "a", "initial"},
	})

	r := NewReader(dir)
	tags, err := r.GetTags()
	require.NoError(t, err)

	// New repo has no tags – map should be empty, not nil
	if tags == nil {
		t.Error("GetTags() returned nil map")
	}
}

func TestGitReaderClose(t *testing.T) {
	r := NewReader("/tmp")
	if err := r.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

func TestGitReaderImplementsVCSReader(t *testing.T) {
	var _ interface {
		Validate() error
		GetCommits() (vcs.CommitIterator, error)
		GetBranches() ([]string, error)
		GetTags() (map[string]string, error)
		Close() error
	} = (*Reader)(nil)
}
