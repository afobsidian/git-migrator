// Package req020 contains requirement validation tests for REQ-020: Git ↔ CVS Bidirectional Sync.
package req020

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/adamf123git/git-migrator/internal/core"
	"github.com/stretchr/testify/require"
)

// TestREQ020_SyncDirectionConstants verifies the string values of SyncDirection constants.
func TestREQ020_SyncDirectionConstants(t *testing.T) {
	require.Equal(t, "git-to-cvs", string(core.SyncGitToCVS))
	require.Equal(t, "cvs-to-git", string(core.SyncCVSToGit))
	require.Equal(t, "bidirectional", string(core.SyncBidirectional))
}

// TestREQ020_NewSyncer ensures NewSyncer initialises all required fields.
func TestREQ020_NewSyncer(t *testing.T) {
	cfg := &core.SyncConfig{
		GitPath:   "/git",
		CVSPath:   "/cvs",
		CVSModule: "mod",
		Direction: core.SyncBidirectional,
	}
	s := core.NewSyncer(cfg)
	require.NotNil(t, s)
	require.NotNil(t, s.ProgressReporter())
}

// TestREQ020_SyncStateJSONRoundTrip verifies that SyncState serialises and deserialises correctly.
func TestREQ020_SyncStateJSONRoundTrip(t *testing.T) {
	dir := t.TempDir()
	stateFile := filepath.Join(dir, "sync.json")

	ts := time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC)
	state := &core.SyncState{
		LastGitCommit: "deadbeef",
		LastCVSSync:   ts,
		SyncedAt:      ts,
	}

	data, err := json.Marshal(state)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(stateFile, data, 0600))

	var loaded core.SyncState
	raw, err := os.ReadFile(stateFile)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(raw, &loaded))

	require.Equal(t, "deadbeef", loaded.LastGitCommit)
	require.True(t, loaded.LastCVSSync.Equal(ts))
}

// TestREQ020_UnknownDirection verifies that an unrecognised direction returns an error.
func TestREQ020_UnknownDirection(t *testing.T) {
	s := core.NewSyncer(&core.SyncConfig{Direction: core.SyncDirection("unknown")})
	err := s.Run()
	require.Error(t, err)
}

// TestREQ020_GitToCVSFailsForMissingRepo verifies graceful failure for missing Git repo.
func TestREQ020_GitToCVSFailsForMissingRepo(t *testing.T) {
	s := core.NewSyncer(&core.SyncConfig{
		GitPath:   "/nonexistent/git",
		CVSPath:   "/nonexistent/cvs",
		CVSModule: "mod",
		Direction: core.SyncGitToCVS,
	})
	require.Error(t, s.Run())
}

// TestREQ020_CVSToGitFailsForMissingRepo verifies graceful failure for missing CVS repo.
func TestREQ020_CVSToGitFailsForMissingRepo(t *testing.T) {
	s := core.NewSyncer(&core.SyncConfig{
		GitPath:   "/nonexistent/git",
		CVSPath:   "/nonexistent/cvs",
		CVSModule: "mod",
		Direction: core.SyncCVSToGit,
	})
	require.Error(t, s.Run())
}

// TestREQ020_SyncConfigFields ensures all required SyncConfig fields are present.
func TestREQ020_SyncConfigFields(t *testing.T) {
	cfg := &core.SyncConfig{
		GitPath:    "/git",
		CVSPath:    "/cvs",
		CVSModule:  "mod",
		CVSWorkDir: "/work",
		Direction:  core.SyncBidirectional,
		AuthorMap:  map[string]string{"alice": "Alice <alice@example.com>"},
		StateFile:  "/tmp/state.json",
		DryRun:     true,
	}
	s := core.NewSyncer(cfg)
	require.NotNil(t, s)
}
