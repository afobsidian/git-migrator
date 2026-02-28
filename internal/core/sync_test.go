package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

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
