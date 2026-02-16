package core

import (
	"fmt"
	"testing"
	"time"

	"github.com/adamf123git/git-migrator/internal/vcs"
	"github.com/stretchr/testify/require"
)

// simple iterator that yields given commits
type sliceIter struct {
	commits []*vcs.Commit
	idx     int
}

func (s *sliceIter) Next() bool {
	s.idx++
	return s.idx <= len(s.commits)
}
func (s *sliceIter) Commit() *vcs.Commit {
	if s.idx < 1 || s.idx > len(s.commits) {
		return nil
	}
	return s.commits[s.idx-1]
}
func (s *sliceIter) Err() error { return nil }

type mockReaderWithCommits struct{ commits []*vcs.Commit }

func (m *mockReaderWithCommits) Validate() error { return nil }
func (m *mockReaderWithCommits) GetCommits() (vcs.CommitIterator, error) {
	return &sliceIter{commits: m.commits}, nil
}
func (m *mockReaderWithCommits) GetBranches() ([]string, error)      { return []string{}, nil }
func (m *mockReaderWithCommits) GetTags() (map[string]string, error) { return map[string]string{}, nil }
func (m *mockReaderWithCommits) Close() error                        { return nil }

func TestRun_DryRunProcessesCommits(t *testing.T) {
	commits := []*vcs.Commit{
		{Revision: "r1", Author: "a1", Date: time.Now(), Message: "m1"},
		{Revision: "r2", Author: "a2", Date: time.Now(), Message: "m2"},
	}

	cfg := &MigrationConfig{SourceType: "cvs", SourcePath: "/src", TargetPath: "/t", DryRun: true}
	m := NewMigrator(cfg)
	m.source = &mockReaderWithCommits{commits: commits}
	// target nil because dry run

	// Should complete without error
	require.NoError(t, m.Run())
}

func TestRun_InterruptAtStops(t *testing.T) {
	commits := []*vcs.Commit{
		{Revision: "r1", Author: "a1", Date: time.Now(), Message: "m1"},
		{Revision: "r2", Author: "a2", Date: time.Now(), Message: "m2"},
	}
	cfg := &MigrationConfig{SourceType: "cvs", SourcePath: "/src", TargetPath: "/t", DryRun: true, InterruptAt: 1}
	m := NewMigrator(cfg)
	m.source = &mockReaderWithCommits{commits: commits}

	err := m.Run()
	require.Error(t, err)
}

// Mock reader that returns validation error
type mockReaderValidateError struct{}

func (m *mockReaderValidateError) Validate() error { return fmt.Errorf("validation failed") }
func (m *mockReaderValidateError) GetCommits() (vcs.CommitIterator, error) {
	return &sliceIter{}, nil
}
func (m *mockReaderValidateError) GetBranches() ([]string, error) { return []string{}, nil }
func (m *mockReaderValidateError) GetTags() (map[string]string, error) {
	return map[string]string{}, nil
}
func (m *mockReaderValidateError) Close() error { return nil }

func TestRun_ValidateError(t *testing.T) {
	cfg := &MigrationConfig{SourceType: "cvs", SourcePath: "/src", TargetPath: "/t", DryRun: true}
	m := NewMigrator(cfg)
	m.source = &mockReaderValidateError{}

	err := m.Run()
	require.Error(t, err)
	require.Contains(t, err.Error(), "validation failed")
}

// Mock reader that returns GetCommits error
type mockReaderGetCommitsError struct{}

func (m *mockReaderGetCommitsError) Validate() error { return nil }
func (m *mockReaderGetCommitsError) GetCommits() (vcs.CommitIterator, error) {
	return nil, fmt.Errorf("get commits failed")
}
func (m *mockReaderGetCommitsError) GetBranches() ([]string, error) { return []string{}, nil }
func (m *mockReaderGetCommitsError) GetTags() (map[string]string, error) {
	return map[string]string{}, nil
}
func (m *mockReaderGetCommitsError) Close() error { return nil }

func TestRun_GetCommitsError(t *testing.T) {
	cfg := &MigrationConfig{SourceType: "cvs", SourcePath: "/src", TargetPath: "/t", DryRun: true}
	m := NewMigrator(cfg)
	m.source = &mockReaderGetCommitsError{}

	err := m.Run()
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get commits")
}

// Mock iterator that returns error
type errorIter struct {
	idx int
}

func (e *errorIter) Next() bool {
	e.idx++
	return e.idx <= 2
}
func (e *errorIter) Commit() *vcs.Commit {
	if e.idx == 1 {
		return &vcs.Commit{Revision: "r1", Author: "a1", Date: time.Now(), Message: "m1"}
	}
	return nil
}
func (e *errorIter) Err() error {
	if e.idx > 1 {
		return fmt.Errorf("iterator error")
	}
	return nil
}

type mockReaderIterError struct{}

func (m *mockReaderIterError) Validate() error { return nil }
func (m *mockReaderIterError) GetCommits() (vcs.CommitIterator, error) {
	return &errorIter{}, nil
}
func (m *mockReaderIterError) GetBranches() ([]string, error)      { return []string{}, nil }
func (m *mockReaderIterError) GetTags() (map[string]string, error) { return map[string]string{}, nil }
func (m *mockReaderIterError) Close() error                        { return nil }

func TestRun_IteratorError(t *testing.T) {
	cfg := &MigrationConfig{SourceType: "cvs", SourcePath: "/src", TargetPath: "/t", DryRun: true}
	m := NewMigrator(cfg)
	m.source = &mockReaderIterError{}

	err := m.Run()
	require.Error(t, err)
	require.Contains(t, err.Error(), "iterator error")
}

// Mock reader with branches and tags
type mockReaderWithBranchesAndTags struct {
	commits []*vcs.Commit
}

func (m *mockReaderWithBranchesAndTags) Validate() error { return nil }
func (m *mockReaderWithBranchesAndTags) GetCommits() (vcs.CommitIterator, error) {
	return &sliceIter{commits: m.commits}, nil
}
func (m *mockReaderWithBranchesAndTags) GetBranches() ([]string, error) {
	return []string{"branch1", "branch2"}, nil
}
func (m *mockReaderWithBranchesAndTags) GetTags() (map[string]string, error) {
	return map[string]string{"tag1": "rev1", "tag2": "rev2"}, nil
}
func (m *mockReaderWithBranchesAndTags) Close() error { return nil }

func TestRun_WithBranchesAndTags(t *testing.T) {
	commits := []*vcs.Commit{
		{Revision: "r1", Author: "a1", Date: time.Now(), Message: "m1"},
	}
	cfg := &MigrationConfig{SourceType: "cvs", SourcePath: "/src", TargetPath: "/t", DryRun: true}
	m := NewMigrator(cfg)
	m.source = &mockReaderWithBranchesAndTags{commits: commits}

	err := m.Run()
	require.NoError(t, err)
}

func TestRun_WithChunkSize(t *testing.T) {
	commits := []*vcs.Commit{
		{Revision: "r1", Author: "a1", Date: time.Now(), Message: "m1"},
		{Revision: "r2", Author: "a2", Date: time.Now(), Message: "m2"},
		{Revision: "r3", Author: "a3", Date: time.Now(), Message: "m3"},
	}
	cfg := &MigrationConfig{SourceType: "cvs", SourcePath: "/src", TargetPath: "/t", DryRun: true, ChunkSize: 2}
	m := NewMigrator(cfg)
	m.source = &mockReaderWithCommits{commits: commits}

	err := m.Run()
	require.NoError(t, err)
}

// Mock reader with GetBranches error
type mockReaderBranchesError struct {
	commits []*vcs.Commit
}

func (m *mockReaderBranchesError) Validate() error { return nil }
func (m *mockReaderBranchesError) GetCommits() (vcs.CommitIterator, error) {
	return &sliceIter{commits: m.commits}, nil
}
func (m *mockReaderBranchesError) GetBranches() ([]string, error) {
	return nil, fmt.Errorf("branches error")
}
func (m *mockReaderBranchesError) GetTags() (map[string]string, error) {
	return map[string]string{}, nil
}
func (m *mockReaderBranchesError) Close() error { return nil }

func TestRun_GetBranchesError(t *testing.T) {
	commits := []*vcs.Commit{
		{
			Revision: "r1",
			Author:   "a1",
			Date:     time.Now(),
			Message:  "m1",
			Files: []vcs.FileChange{
				{
					Path:    "file.txt",
					Action:  vcs.ActionAdd,
					Content: []byte("content"),
				},
			},
		},
	}
	// Use a valid temp directory for target path
	tmpDir := t.TempDir()
	repoPath := tmpDir + "/repo"
	cfg := &MigrationConfig{SourceType: "cvs", SourcePath: "/src", TargetPath: repoPath, DryRun: false}
	m := NewMigrator(cfg)
	m.source = &mockReaderBranchesError{commits: commits}

	err := m.Run()
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to create branches")
}

// Mock reader with GetTags error
type mockReaderTagsError struct {
	commits []*vcs.Commit
}

func (m *mockReaderTagsError) Validate() error { return nil }
func (m *mockReaderTagsError) GetCommits() (vcs.CommitIterator, error) {
	return &sliceIter{commits: m.commits}, nil
}
func (m *mockReaderTagsError) GetBranches() ([]string, error) { return []string{}, nil }
func (m *mockReaderTagsError) GetTags() (map[string]string, error) {
	return nil, fmt.Errorf("tags error")
}
func (m *mockReaderTagsError) Close() error { return nil }

func TestRun_GetTagsError(t *testing.T) {
	commits := []*vcs.Commit{
		{
			Revision: "r1",
			Author:   "a1",
			Date:     time.Now(),
			Message:  "m1",
			Files: []vcs.FileChange{
				{
					Path:    "file.txt",
					Action:  vcs.ActionAdd,
					Content: []byte("content"),
				},
			},
		},
	}
	// Use a valid temp directory for target path
	tmpDir := t.TempDir()
	repoPath := tmpDir + "/repo"
	cfg := &MigrationConfig{SourceType: "cvs", SourcePath: "/src", TargetPath: repoPath, DryRun: false}
	m := NewMigrator(cfg)
	m.source = &mockReaderTagsError{commits: commits}

	err := m.Run()
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to create tags")
}

func TestRun_InitSourceError(t *testing.T) {
	cfg := &MigrationConfig{SourceType: "unsupported", SourcePath: "/src", TargetPath: "/t", DryRun: true}
	m := NewMigrator(cfg)
	// Don't set m.source, so it will try to init

	err := m.Run()
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to init source")
}
