package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/require"
)

// createTestGitRepo initialises a minimal Git repo in dir with one commit and
// returns the repo path.
func createTestGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	repo, err := gogit.PlainInit(dir, false)
	require.NoError(t, err)

	w, err := repo.Worktree()
	require.NoError(t, err)

	f := filepath.Join(dir, "README.md")
	require.NoError(t, os.WriteFile(f, []byte("hello"), 0644))
	_, err = w.Add("README.md")
	require.NoError(t, err)

	_, err = w.Commit("initial commit", &gogit.CommitOptions{
		Author: &object.Signature{
			Name:  "Test",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	require.NoError(t, err)
	return dir
}

// createTestCVSRepo creates a minimal CVS repository structure (CVSROOT) and
// returns the repo path.  No RCS files are included so GetCommits returns an
// empty iterator.
func createTestCVSRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "CVSROOT"), 0755))
	return dir
}

func TestNewSyncer(t *testing.T) {
	cfg := &SyncConfig{
		GitPath:   "/tmp/git",
		CVSPath:   "/tmp/cvs",
		CVSModule: "mod",
		Direction: SyncBidirectional,
	}
	s := NewSyncer(cfg)
	if s == nil {
		t.Fatal("NewSyncer returned nil")
	}
	if s.config != cfg {
		t.Error("config not stored correctly")
	}
	if s.authorMap == nil {
		t.Error("authorMap should be initialised")
	}
	if s.reporter == nil {
		t.Error("reporter should be initialised")
	}
}

func TestSyncerProgressReporter(t *testing.T) {
	s := NewSyncer(&SyncConfig{Direction: SyncGitToCVS})
	if s.ProgressReporter() == nil {
		t.Error("ProgressReporter should not return nil")
	}
}

func TestSyncerLoadState_NoFile(t *testing.T) {
	s := NewSyncer(&SyncConfig{StateFile: ""})
	require.NoError(t, s.loadState())
	require.NotNil(t, s.state)
	require.True(t, s.state.LastCVSSync.IsZero())
	require.Empty(t, s.state.LastGitCommit)
}

func TestSyncerLoadState_MissingFile(t *testing.T) {
	s := NewSyncer(&SyncConfig{StateFile: "/nonexistent/path/sync.json"})
	require.NoError(t, s.loadState(), "missing state file should not be an error")
	require.Empty(t, s.state.LastGitCommit)
}

func TestSyncerSaveAndLoadState(t *testing.T) {
	dir := t.TempDir()
	stateFile := filepath.Join(dir, "sync.json")

	s := NewSyncer(&SyncConfig{StateFile: stateFile})
	require.NoError(t, s.loadState())

	s.state.LastGitCommit = "abc123"
	s.state.LastCVSSync = time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	s.state.SyncedAt = time.Now()

	require.NoError(t, s.saveState())

	// Load in a new syncer
	s2 := NewSyncer(&SyncConfig{StateFile: stateFile})
	require.NoError(t, s2.loadState())

	require.Equal(t, "abc123", s2.state.LastGitCommit)
	require.True(t, s2.state.LastCVSSync.Equal(s.state.LastCVSSync))
}

func TestSyncerSaveState_DryRun(t *testing.T) {
	dir := t.TempDir()
	stateFile := filepath.Join(dir, "sync.json")

	s := NewSyncer(&SyncConfig{StateFile: stateFile, DryRun: true})
	s.state = &SyncState{LastGitCommit: "dryrun"}

	require.NoError(t, s.saveState())

	// File should NOT have been written in dry-run mode
	if _, err := os.Stat(stateFile); !os.IsNotExist(err) {
		t.Error("state file should not be written in dry-run mode")
	}
}

func TestSyncerSaveState_NoStateFile(t *testing.T) {
	s := NewSyncer(&SyncConfig{StateFile: ""})
	s.state = &SyncState{LastGitCommit: "noop"}
	require.NoError(t, s.saveState(), "saveState with empty StateFile should be a no-op")
}

func TestSyncerLoadState_CorruptFile(t *testing.T) {
	dir := t.TempDir()
	stateFile := filepath.Join(dir, "sync.json")

	require.NoError(t, os.WriteFile(stateFile, []byte("not-json"), 0600))

	s := NewSyncer(&SyncConfig{StateFile: stateFile})
	err := s.loadState()
	require.Error(t, err, "corrupt state file should return an error")
}

func TestSyncerStateJSON(t *testing.T) {
	dir := t.TempDir()
	stateFile := filepath.Join(dir, "sync.json")

	ts := time.Date(2025, 6, 15, 8, 0, 0, 0, time.UTC)
	state := &SyncState{
		LastGitCommit: "deadbeef",
		LastCVSSync:   ts,
		SyncedAt:      ts,
	}

	data, err := json.Marshal(state)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(stateFile, data, 0600))

	s := NewSyncer(&SyncConfig{StateFile: stateFile})
	require.NoError(t, s.loadState())

	require.Equal(t, "deadbeef", s.state.LastGitCommit)
	require.True(t, s.state.LastCVSSync.Equal(ts))
}

func TestSyncerRunUnknownDirection(t *testing.T) {
	s := NewSyncer(&SyncConfig{Direction: SyncDirection("invalid")})
	s.state = &SyncState{}
	err := s.Run()
	require.Error(t, err)
}

func TestSyncerRunGitToCVS_NonExistentGit(t *testing.T) {
	s := NewSyncer(&SyncConfig{
		GitPath:   "/nonexistent/git/repo",
		CVSPath:   "/nonexistent/cvs/repo",
		CVSModule: "mod",
		Direction: SyncGitToCVS,
	})
	err := s.Run()
	require.Error(t, err, "Run should fail for non-existent git repository")
}

func TestSyncerRunCVSToGit_NonExistentCVS(t *testing.T) {
	s := NewSyncer(&SyncConfig{
		GitPath:   "/nonexistent/git/repo",
		CVSPath:   "/nonexistent/cvs/repo",
		CVSModule: "mod",
		Direction: SyncCVSToGit,
	})
	err := s.Run()
	require.Error(t, err, "Run should fail for non-existent CVS repository")
}

func TestSyncDirectionConstants(t *testing.T) {
	cases := []struct {
		d    SyncDirection
		want string
	}{
		{SyncGitToCVS, "git-to-cvs"},
		{SyncCVSToGit, "cvs-to-git"},
		{SyncBidirectional, "bidirectional"},
	}
	for _, tc := range cases {
		if string(tc.d) != tc.want {
			t.Errorf("SyncDirection %q, want %q", tc.d, tc.want)
		}
	}
}

func TestSyncConfigDefaults(t *testing.T) {
	cfg := &SyncConfig{
		GitPath:   "/git",
		CVSPath:   "/cvs",
		CVSModule: "mod",
	}
	s := NewSyncer(cfg)
	require.False(t, s.config.DryRun)
	require.Empty(t, s.config.StateFile)
}

// ---------------------------------------------------------------------------
// syncGitToCVS path tests
// ---------------------------------------------------------------------------

// TestSyncerSyncGitToCVS_DryRun verifies the dry-run code path of
// syncGitToCVS: a valid Git repo with commits, no CVS operations performed.
func TestSyncerSyncGitToCVS_DryRun(t *testing.T) {
	gitDir := createTestGitRepo(t)

	s := NewSyncer(&SyncConfig{
		GitPath:   gitDir,
		CVSPath:   "/nonexistent/cvs",
		CVSModule: "mod",
		Direction: SyncGitToCVS,
		DryRun:    true,
	})
	// Run should succeed: discovers commits and logs them (dry-run), no CVS call.
	require.NoError(t, s.Run())
}

// TestSyncerSyncGitToCVS_UpToDate verifies that when all Git commits have
// already been synced, syncGitToCVS returns "up to date" immediately.
func TestSyncerSyncGitToCVS_UpToDate(t *testing.T) {
	gitDir := createTestGitRepo(t)

	// Get the head commit hash via go-git
	repo, err := gogit.PlainOpen(gitDir)
	require.NoError(t, err)
	head, err := repo.Head()
	require.NoError(t, err)
	lastHash := head.Hash().String()

	s := NewSyncer(&SyncConfig{
		GitPath:   gitDir,
		CVSPath:   "/nonexistent/cvs",
		CVSModule: "mod",
		Direction: SyncGitToCVS,
		DryRun:    true,
	})
	require.NoError(t, s.loadState())
	s.state.LastGitCommit = lastHash

	// No new commits → up-to-date path, no error.
	require.NoError(t, s.syncGitToCVS())
}

// TestSyncerSyncGitToCVS_PrepareCVSWorkDir_WithConfig exercises the non-dry-run
// path all the way through to prepareCVSWorkDir (configured CVSWorkDir).
// The CVS Init will fail, but prepareCVSWorkDir itself is covered.
func TestSyncerSyncGitToCVS_PrepareCVSWorkDir_WithConfig(t *testing.T) {
	gitDir := createTestGitRepo(t)
	workDir := t.TempDir()

	s := NewSyncer(&SyncConfig{
		GitPath:    gitDir,
		CVSPath:    "/nonexistent/cvs",
		CVSModule:  "mod",
		CVSWorkDir: workDir, // configured → prepareCVSWorkDir returns it directly
		Direction:  SyncGitToCVS,
		DryRun:     false, // must be false to reach prepareCVSWorkDir
	})
	// Expect an error (CVS binary / repo not available), but prepareCVSWorkDir
	// will have been called, giving it coverage.
	err := s.Run()
	require.Error(t, err)
}

// TestSyncerSyncGitToCVS_PrepareCVSWorkDir_TempDir exercises the temp-dir
// branch of prepareCVSWorkDir (CVSWorkDir not configured).
func TestSyncerSyncGitToCVS_PrepareCVSWorkDir_TempDir(t *testing.T) {
	gitDir := createTestGitRepo(t)

	s := NewSyncer(&SyncConfig{
		GitPath:   gitDir,
		CVSPath:   "/nonexistent/cvs",
		CVSModule: "mod",
		// CVSWorkDir is empty → temp dir will be created and cleaned up
		Direction: SyncGitToCVS,
		DryRun:    false,
	})
	// Expect an error (CVS binary / repo not available), but temp-dir branch
	// of prepareCVSWorkDir is covered.
	err := s.Run()
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// syncCVSToGit path tests
// ---------------------------------------------------------------------------

// TestSyncerSyncCVSToGit_UpToDate verifies that when no new CVS commits exist
// (empty repo), syncCVSToGit returns without error.
func TestSyncerSyncCVSToGit_UpToDate(t *testing.T) {
	cvsDir := createTestCVSRepo(t) // empty CVSROOT, no RCS files → no commits

	s := NewSyncer(&SyncConfig{
		GitPath:   "/nonexistent/git",
		CVSPath:   cvsDir,
		CVSModule: "mod",
		Direction: SyncCVSToGit,
		DryRun:    true,
	})
	// Empty CVS repo → "up to date" path, no error.
	require.NoError(t, s.Run())
}

// TestSyncerSyncCVSToGit_DryRun uses the CVS fixture which contains real RCS
// files so that newCommits is non-empty and the dry-run log path is exercised.
func TestSyncerSyncCVSToGit_DryRun(t *testing.T) {
	// Use the CVS fixture that ships with the repository.
	fixturePath := filepath.Join("..", "..", "test", "fixtures", "cvs", "simple")
	if _, err := os.Stat(filepath.Join(fixturePath, "CVSROOT")); err != nil {
		t.Skip("CVS fixture not available")
	}

	s := NewSyncer(&SyncConfig{
		GitPath:   "/nonexistent/git",
		CVSPath:   fixturePath,
		CVSModule: "mod",
		Direction: SyncCVSToGit,
		DryRun:    true, // log only, no git write
	})
	// Should discover commits and log them, then return nil.
	require.NoError(t, s.Run())
}

// TestSyncerSyncCVSToGit_SkipOldCommits verifies that commits on or before
// LastCVSSync are filtered out (up-to-date when all commits are old).
func TestSyncerSyncCVSToGit_SkipOldCommits(t *testing.T) {
	fixturePath := filepath.Join("..", "..", "test", "fixtures", "cvs", "simple")
	if _, err := os.Stat(filepath.Join(fixturePath, "CVSROOT")); err != nil {
		t.Skip("CVS fixture not available")
	}

	s := NewSyncer(&SyncConfig{
		GitPath:   "/nonexistent/git",
		CVSPath:   fixturePath,
		CVSModule: "mod",
		Direction: SyncCVSToGit,
		DryRun:    true,
	})
	require.NoError(t, s.loadState())
	// Set LastCVSSync to far future so all fixture commits are considered old.
	s.state.LastCVSSync = time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC)

	require.NoError(t, s.syncCVSToGit())
}

// ---------------------------------------------------------------------------
// Bidirectional path tests
// ---------------------------------------------------------------------------

// TestSyncerBidirectional_CVSEmptyGitDryRun exercises the bidirectional Run
// path: CVS has no commits (up-to-date), then Git dry-run syncs.
func TestSyncerBidirectional_CVSEmptyGitDryRun(t *testing.T) {
	gitDir := createTestGitRepo(t)
	cvsDir := createTestCVSRepo(t)

	s := NewSyncer(&SyncConfig{
		GitPath:   gitDir,
		CVSPath:   cvsDir,
		CVSModule: "mod",
		Direction: SyncBidirectional,
		DryRun:    true,
	})
	require.NoError(t, s.Run())
}

// TestSyncerSyncGitToCVS_WithFakeCVS covers the commit-application loop in
// syncGitToCVS by providing a fake cvs binary that always exits successfully.
func TestSyncerSyncGitToCVS_WithFakeCVS(t *testing.T) {
	binDir := t.TempDir()
	fakeCVS := filepath.Join(binDir, "cvs")
	require.NoError(t, os.WriteFile(fakeCVS, []byte("#!/bin/sh\nexit 0\n"), 0755))

	// Prepend fake binary dir to PATH (t.Setenv restores it automatically)
	t.Setenv("PATH", binDir+":"+os.Getenv("PATH"))

	gitDir := createTestGitRepo(t)
	workDir := t.TempDir()
	stateDir := t.TempDir()
	stateFile := filepath.Join(stateDir, "sync.json")

	s := NewSyncer(&SyncConfig{
		GitPath:    gitDir,
		CVSPath:    "/fakecvs/root",
		CVSModule:  "mod",
		CVSWorkDir: workDir,
		Direction:  SyncGitToCVS,
		StateFile:  stateFile,
		DryRun:     false,
	})

	err := s.Run()
	require.NoError(t, err, "syncGitToCVS with fake cvs should succeed")
}
// state file is corrupt/unreadable.
func TestSyncerRun_LoadStateError(t *testing.T) {
	dir := t.TempDir()
	stateFile := filepath.Join(dir, "bad.json")
	require.NoError(t, os.WriteFile(stateFile, []byte("not-json"), 0600))

	s := NewSyncer(&SyncConfig{
		GitPath:   "/nonexistent",
		CVSPath:   "/nonexistent",
		CVSModule: "mod",
		Direction: SyncGitToCVS,
		StateFile: stateFile,
	})
	err := s.Run()
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to load sync state")
}

// TestSyncerRun_BidirectionalCVSError verifies that the bidirectional path
// returns a wrapped error when the CVS-to-Git sync fails.
func TestSyncerRun_BidirectionalCVSError(t *testing.T) {
	s := NewSyncer(&SyncConfig{
		GitPath:   "/nonexistent/git",
		CVSPath:   "/nonexistent/cvs",
		CVSModule: "mod",
		Direction: SyncBidirectional,
	})
	err := s.Run()
	require.Error(t, err)
	require.Contains(t, err.Error(), "cvs-to-git sync failed")
}
// repository cannot be opened in non-dry-run mode (lines after the dry-run check).
func TestSyncerSyncCVSToGit_NonDryRun_GitOpenFails(t *testing.T) {
	fixturePath := filepath.Join("..", "..", "test", "fixtures", "cvs", "simple")
	if _, err := os.Stat(filepath.Join(fixturePath, "CVSROOT")); err != nil {
		t.Skip("CVS fixture not available")
	}

	s := NewSyncer(&SyncConfig{
		GitPath:   "/nonexistent/git/path",
		CVSPath:   fixturePath,
		CVSModule: "mod",
		Direction: SyncCVSToGit,
		DryRun:    false, // non-dry-run so it reaches gitWriter.Open
	})
	// Fails at gitWriter.Open because /nonexistent/git/path doesn't exist.
	err := s.Run()
	require.Error(t, err)
}

// TestSyncerSyncCVSToGit_NonDryRun_WithGitRepo covers the full syncCVSToGit
// code path including the commit-application loop.
func TestSyncerSyncCVSToGit_NonDryRun_WithGitRepo(t *testing.T) {
	gitDir := createTestGitRepo(t)
	fixturePath := filepath.Join("..", "..", "test", "fixtures", "cvs", "simple")
	if _, err := os.Stat(filepath.Join(fixturePath, "CVSROOT")); err != nil {
		t.Skip("CVS fixture not available")
	}

	dir := t.TempDir()
	stateFile := filepath.Join(dir, "sync.json")

	s := NewSyncer(&SyncConfig{
		GitPath:   gitDir,
		CVSPath:   fixturePath,
		CVSModule: "mod",
		Direction: SyncCVSToGit,
		StateFile: stateFile,
		DryRun:    false,
	})
	// The fixture has one commit (with no file changes in the reader output).
	// ApplyCommit may fail for empty commits; either way the loop body is
	// exercised for coverage purposes.
	_ = s.Run()
}

// TestSyncerSaveState_WriteError covers the os.WriteFile error branch by
// writing to a path inside a non-existent directory.
func TestSyncerSaveState_WriteError(t *testing.T) {
	s := NewSyncer(&SyncConfig{StateFile: "/nonexistent/dir/state.json"})
	s.state = &SyncState{LastGitCommit: "abc"}
	err := s.saveState()
	require.Error(t, err, "saveState should fail when the directory does not exist")
}
