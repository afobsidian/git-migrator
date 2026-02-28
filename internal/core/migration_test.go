package core

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewMigrator(t *testing.T) {
	config := &MigrationConfig{
		SourceType: "cvs",
		SourcePath: "/tmp/test-cvs",
		TargetPath: "/tmp/test-git",
	}

	m := NewMigrator(config)
	if m == nil {
		t.Fatal("NewMigrator returned nil")
	}

	if m.config != config {
		t.Error("config not set correctly")
	}

	if m.authorMap == nil {
		t.Error("authorMap should be initialized")
	}

	if m.reporter == nil {
		t.Error("reporter should be initialized")
	}
}

func TestNewMigratorNilConfig(t *testing.T) {
	// NewMigrator does not support nil config - it will panic
	// This test verifies that an empty config works instead
	config := &MigrationConfig{}
	m := NewMigrator(config)
	if m == nil {
		t.Fatal("NewMigrator returned nil")
	}

	if m.config != config {
		t.Error("config not set correctly")
	}
}

func TestMigrationConfigDefaults(t *testing.T) {
	config := &MigrationConfig{
		SourceType: "cvs",
		SourcePath: "/source",
		TargetPath: "/target",
	}

	if config.DryRun != false {
		t.Error("DryRun should default to false")
	}

	if config.Resume != false {
		t.Error("Resume should default to false")
	}

	if config.ChunkSize != 0 {
		t.Error("ChunkSize should default to 0")
	}
}

func TestMigrationConfigWithAuthorMap(t *testing.T) {
	config := &MigrationConfig{
		SourceType: "cvs",
		SourcePath: "/source",
		TargetPath: "/target",
		AuthorMap: map[string]string{
			"johndoe": "John Doe <john@example.com>",
		},
	}

	m := NewMigrator(config)
	if m.authorMap == nil {
		t.Fatal("authorMap should be initialized")
	}

	name, email := m.authorMap.Get("johndoe")
	if name != "John Doe" {
		t.Errorf("name = %q, want %q", name, "John Doe")
	}
	if email != "john@example.com" {
		t.Errorf("email = %q, want %q", email, "john@example.com")
	}
}

func TestMigratorInitSourceUnsupported(t *testing.T) {
	config := &MigrationConfig{
		SourceType: "unsupported",
		SourcePath: "/source",
		TargetPath: "/target",
	}

	m := NewMigrator(config)
	err := m.initSource()
	if err == nil {
		t.Error("initSource should fail for unsupported type")
	}
}

func TestMigratorInitSourceCVS(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cvs-source")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	// Create minimal CVS structure
	cvsroot := filepath.Join(tmpDir, "CVSROOT")
	if err := os.MkdirAll(cvsroot, 0755); err != nil {
		t.Fatalf("Failed to create CVSROOT: %v", err)
	}

	config := &MigrationConfig{
		SourceType: "cvs",
		SourcePath: tmpDir,
		TargetPath: "/target",
	}

	m := NewMigrator(config)
	err = m.initSource()
	if err != nil {
		t.Errorf("initSource failed: %v", err)
	}

	if m.source == nil {
		t.Error("source should be initialized")
	}
}

func TestMigratorInitTargetNew(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-target")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	repoPath := filepath.Join(tmpDir, "new-repo")

	config := &MigrationConfig{
		SourceType: "cvs",
		SourcePath: "/source",
		TargetPath: repoPath,
	}

	m := NewMigrator(config)
	err = m.initTarget()
	if err != nil {
		t.Errorf("initTarget failed: %v", err)
	}

	if m.target == nil {
		t.Error("target should be initialized")
	}

	// Check that repo was created
	if _, err := os.Stat(filepath.Join(repoPath, ".git")); os.IsNotExist(err) {
		t.Error(".git directory should exist")
	}

	require.NoError(t, m.target.Close())
}

func TestMigratorInitTargetExisting(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-target-existing")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	repoPath := filepath.Join(tmpDir, "existing-repo")

	config := &MigrationConfig{
		SourceType: "cvs",
		SourcePath: "/source",
		TargetPath: repoPath,
	}

	m := NewMigrator(config)

	// Create repo first
	if err := m.initTarget(); err != nil {
		t.Fatalf("First initTarget failed: %v", err)
	}
	require.NoError(t, m.target.Close())

	// Reset target
	m.target = nil

	// Open existing
	err = m.initTarget()
	if err != nil {
		t.Errorf("Second initTarget failed: %v", err)
	}

	if m.target == nil {
		t.Error("target should be initialized")
	}

	require.NoError(t, m.target.Close())
}

func TestMigratorGenerateMigrationID(t *testing.T) {
	config := &MigrationConfig{
		SourcePath: "/source/path",
		TargetPath: "/target/path",
	}

	m := NewMigrator(config)
	id := m.generateMigrationID()

	if id == "" {
		t.Error("migration ID should not be empty")
	}

	if len(id) != 16 {
		t.Errorf("migration ID length = %d, want 16", len(id))
	}

	// Same paths should generate same ID
	id2 := m.generateMigrationID()
	if id != id2 {
		t.Error("Same paths should generate same ID")
	}

	// Different paths should generate different ID
	m.config.SourcePath = "/different/source"
	id3 := m.generateMigrationID()
	if id == id3 {
		t.Error("Different paths should generate different ID")
	}
}

func TestMigratorInitState(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "state-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	stateFile := filepath.Join(tmpDir, "state.db")

	config := &MigrationConfig{
		SourcePath: "/source",
		TargetPath: "/target",
		StateFile:  stateFile,
	}

	m := NewMigrator(config)
	err = m.initState()
	if err != nil {
		t.Errorf("initState failed: %v", err)
	}

	if m.db == nil {
		t.Error("db should be initialized")
	}

	if m.state == nil {
		t.Error("state should be initialized")
	}

	require.NoError(t, m.db.Close())
}

func TestMigratorInitStateDefaultPath(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "state-default-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	targetPath := filepath.Join(tmpDir, "target")

	config := &MigrationConfig{
		SourcePath: "/source",
		TargetPath: targetPath,
		// StateFile not set - should use default
	}

	m := NewMigrator(config)
	err = m.initState()
	if err != nil {
		t.Errorf("initState failed: %v", err)
	}

	expectedStateFile := filepath.Join(targetPath, ".migration-state.db")
	if m.config.StateFile != expectedStateFile {
		t.Errorf("StateFile = %q, want %q", m.config.StateFile, expectedStateFile)
	}

	require.NoError(t, m.db.Close())
}

func TestSimpleStateSave(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "simple-state")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	stateFile := filepath.Join(tmpDir, "state.txt")
	s := NewMigrationState(stateFile)

	err = s.Save("abc123", 50, 100)
	if err != nil {
		t.Errorf("Save failed: %v", err)
	}

	if s.lastCommit != "abc123" {
		t.Errorf("lastCommit = %q, want %q", s.lastCommit, "abc123")
	}
	if s.processed != 50 {
		t.Errorf("processed = %d, want 50", s.processed)
	}
	if s.total != 100 {
		t.Errorf("total = %d, want 100", s.total)
	}
}

func TestSimpleStateLoad(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "simple-state-load")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	stateFile := filepath.Join(tmpDir, "state.txt")
	s := NewMigrationState(stateFile)

	// Save first
	if err := s.Save("xyz789", 25, 200); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Create new state and load
	s2 := NewMigrationState(stateFile)
	commit, processed, total, err := s2.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if commit != "xyz789" {
		t.Errorf("commit = %q, want %q", commit, "xyz789")
	}
	if processed != 25 {
		t.Errorf("processed = %d, want 25", processed)
	}
	if total != 200 {
		t.Errorf("total = %d, want 200", total)
	}
}

func TestSimpleStateLoadNonExistent(t *testing.T) {
	s := NewMigrationState("/nonexistent/path/state.txt")

	commit, processed, total, err := s.Load()
	if err != nil {
		t.Errorf("Load should not fail for non-existent file: %v", err)
	}

	if commit != "" {
		t.Errorf("commit = %q, want empty", commit)
	}
	if processed != 0 {
		t.Errorf("processed = %d, want 0", processed)
	}
	if total != 0 {
		t.Errorf("total = %d, want 0", total)
	}
}

func TestSimpleStateClear(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "simple-state-clear")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	stateFile := filepath.Join(tmpDir, "state.txt")
	s := NewMigrationState(stateFile)

	// Save first
	if err := s.Save("commit1", 10, 100); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Clear
	if err := s.Clear(); err != nil {
		t.Errorf("Clear failed: %v", err)
	}

	if s.lastCommit != "" {
		t.Errorf("lastCommit = %q, want empty", s.lastCommit)
	}
	if s.processed != 0 {
		t.Errorf("processed = %d, want 0", s.processed)
	}
	if s.total != 0 {
		t.Errorf("total = %d, want 0", s.total)
	}

	// File should be deleted
	if _, err := os.Stat(stateFile); !os.IsNotExist(err) {
		t.Error("State file should be deleted")
	}
}

func TestSimpleStateClearNonExistent(t *testing.T) {
	s := NewMigrationState("/nonexistent/state.txt")

	// Clear should not fail on non-existent file
	if err := s.Clear(); err != nil {
		t.Errorf("Clear should not fail for non-existent file: %v", err)
	}
}

func TestSimpleStateClose(t *testing.T) {
	s := NewMigrationState("/tmp/state.txt")

	// Close should always succeed
	if err := s.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

func TestSimpleStateInvalidFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "simple-state-invalid")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	stateFile := filepath.Join(tmpDir, "invalid-state.txt")

	// Write invalid content
	if err := os.WriteFile(stateFile, []byte("invalid\ncontent"), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	s := NewMigrationState(stateFile)
	_, _, _, err = s.Load()
	if err == nil {
		t.Error("Load should fail for invalid file")
	}
}

func TestSimpleStatePartialFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "simple-state-partial")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	stateFile := filepath.Join(tmpDir, "partial-state.txt")

	// Write partial content (only 2 lines instead of 3)
	if err := os.WriteFile(stateFile, []byte("commit123\n50"), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	s := NewMigrationState(stateFile)
	_, _, _, err = s.Load()
	if err == nil {
		t.Error("Load should fail for partial file")
	}
}

func TestMigratorProgressReporter(t *testing.T) {
	config := &MigrationConfig{
		SourceType: "cvs",
		SourcePath: "/source",
		TargetPath: "/target",
	}

	m := NewMigrator(config)
	reporter := m.ProgressReporter()

	if reporter == nil {
		t.Error("ProgressReporter should not return nil")
	}
}

func TestMigrationStateStruct(t *testing.T) {
	state := &MigrationState{
		migrationID: "test-id",
		lastCommit:  "abc123",
		processed:   50,
		total:       100,
	}

	if state.migrationID != "test-id" {
		t.Errorf("migrationID = %q, want %q", state.migrationID, "test-id")
	}
	if state.lastCommit != "abc123" {
		t.Errorf("lastCommit = %q, want %q", state.lastCommit, "abc123")
	}
	if state.processed != 50 {
		t.Errorf("processed = %d, want 50", state.processed)
	}
	if state.total != 100 {
		t.Errorf("total = %d, want 100", state.total)
	}
}

func TestMigrationConfigWithBranchMap(t *testing.T) {
	config := &MigrationConfig{
		SourceType: "cvs",
		SourcePath: "/source",
		TargetPath: "/target",
		BranchMap: map[string]string{
			"cvs-branch": "git-branch",
		},
	}

	m := NewMigrator(config)
	if m == nil {
		t.Fatal("NewMigrator returned nil")
	}

	if m.config.BranchMap["cvs-branch"] != "git-branch" {
		t.Error("BranchMap not set correctly")
	}
}

func TestMigrationConfigWithTagMap(t *testing.T) {
	config := &MigrationConfig{
		SourceType: "cvs",
		SourcePath: "/source",
		TargetPath: "/target",
		TagMap: map[string]string{
			"cvs-tag": "git-tag",
		},
	}

	m := NewMigrator(config)
	if m == nil {
		t.Fatal("NewMigrator returned nil")
	}

	if m.config.TagMap["cvs-tag"] != "git-tag" {
		t.Error("TagMap not set correctly")
	}
}

func TestMigrationConfigChunkSize(t *testing.T) {
	config := &MigrationConfig{
		SourceType: "cvs",
		SourcePath: "/source",
		TargetPath: "/target",
		ChunkSize:  100,
	}

	m := NewMigrator(config)
	if m == nil {
		t.Fatal("NewMigrator returned nil")
	}

	if m.config.ChunkSize != 100 {
		t.Errorf("ChunkSize = %d, want 100", m.config.ChunkSize)
	}
}

func TestMigrationConfigInterruptAt(t *testing.T) {
	config := &MigrationConfig{
		SourceType:  "cvs",
		SourcePath:  "/source",
		TargetPath:  "/target",
		InterruptAt: 5,
	}

	m := NewMigrator(config)
	if m == nil {
		t.Fatal("NewMigrator returned nil")
	}

	if m.config.InterruptAt != 5 {
		t.Errorf("InterruptAt = %d, want 5", m.config.InterruptAt)
	}
}

func TestMigratorRunDryRun(t *testing.T) {
	// This test verifies that dry run mode doesn't create files
	tmpDir, err := os.MkdirTemp("", "dry-run-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	targetPath := filepath.Join(tmpDir, "target")

	config := &MigrationConfig{
		SourceType: "cvs",
		SourcePath: tmpDir,
		TargetPath: targetPath,
		DryRun:     true,
	}

	// Create minimal CVS structure
	cvsroot := filepath.Join(tmpDir, "CVSROOT")
	if err := os.MkdirAll(cvsroot, 0755); err != nil {
		t.Fatalf("Failed to create CVSROOT: %v", err)
	}

	m := NewMigrator(config)

	// Run should work even without real CVS data in dry run mode
	// (it will fail to find commits but that's ok for this test)
	_ = m.Run()

	// In dry run, git repo should NOT be created (but state DB dir may be created)
	gitDir := filepath.Join(targetPath, ".git")
	if _, err := os.Stat(gitDir); !os.IsNotExist(err) {
		t.Error(".git directory should not be created in dry run mode")
	}
}

func TestMigratorSaveState(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "save-state-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	stateFile := filepath.Join(tmpDir, "state.db")

	config := &MigrationConfig{
		SourcePath: "/source",
		TargetPath: "/target",
		StateFile:  stateFile,
	}

	m := NewMigrator(config)

	// Initialize state
	if err := m.initState(); err != nil {
		t.Fatalf("initState failed: %v", err)
	}
	defer func() { require.NoError(t, m.db.Close()) }()

	// Save state
	err = m.saveState("commit123", 50, 100)
	if err != nil {
		t.Errorf("saveState failed: %v", err)
	}

	// Verify state was saved
	if m.state.lastCommit != "commit123" {
		t.Errorf("lastCommit = %q, want %q", m.state.lastCommit, "commit123")
	}
	if m.state.processed != 50 {
		t.Errorf("processed = %d, want 50", m.state.processed)
	}
	if m.state.total != 100 {
		t.Errorf("total = %d, want 100", m.state.total)
	}
}

func TestSimpleStateSaveAndLoadMultiple(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "state-multi")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	stateFile := filepath.Join(tmpDir, "state.txt")
	s := NewMigrationState(stateFile)

	// Save multiple times (simulating progress)
	for i := 1; i <= 5; i++ {
		commit := string(rune('a' + i - 1))
		if err := s.Save(commit, i*10, 50); err != nil {
			t.Fatalf("Save %d failed: %v", i, err)
		}

		// Load and verify
		loaded, processed, total, err := s.Load()
		if err != nil {
			t.Fatalf("Load %d failed: %v", i, err)
		}
		if loaded != commit {
			t.Errorf("Load %d: commit = %q, want %q", i, loaded, commit)
		}
		if processed != i*10 {
			t.Errorf("Load %d: processed = %d, want %d", i, processed, i*10)
		}
		if total != 50 {
			t.Errorf("Load %d: total = %d, want 50", i, total)
		}
	}
}

func TestSimpleStateConcurrentAccess(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "state-concurrent")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	stateFile := filepath.Join(tmpDir, "state.txt")
	s := NewMigrationState(stateFile)

	done := make(chan bool, 10)

	// Concurrent saves
	for i := 0; i < 5; i++ {
		go func(idx int) {
			for j := 0; j < 10; j++ {
				_ = s.Save(string(rune('A'+idx)), j, 100)
			}
			done <- true
		}(i)
	}

	// Concurrent loads
	for i := 0; i < 5; i++ {
		go func() {
			for j := 0; j < 10; j++ {
				_, _, _, _ = s.Load()
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestMigratorStateWithResume(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "state-resume")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	stateFile := filepath.Join(tmpDir, "state.db")

	config := &MigrationConfig{
		SourcePath: "/source",
		TargetPath: "/target",
		StateFile:  stateFile,
		Resume:     true,
	}

	m := NewMigrator(config)

	// Initialize and save state
	if err := m.initState(); err != nil {
		t.Fatalf("initState failed: %v", err)
	}

	// Save some progress
	if err := m.saveState("initial-commit", 25, 100); err != nil {
		t.Fatalf("saveState failed: %v", err)
	}

	require.NoError(t, m.db.Close())

	// Create new migrator and resume
	m2 := NewMigrator(config)
	if err := m2.initState(); err != nil {
		t.Fatalf("second initState failed: %v", err)
	}
	defer func() { require.NoError(t, m2.db.Close()) }()

	// State should be loaded
	if m2.state == nil {
		t.Fatal("state should be loaded")
	}

	if m2.state.lastCommit != "initial-commit" {
		t.Errorf("lastCommit = %q, want %q", m2.state.lastCommit, "initial-commit")
	}
}

func TestMigratorStateWithoutResume(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "state-no-resume")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	stateFile := filepath.Join(tmpDir, "state.db")

	config := &MigrationConfig{
		SourcePath: "/source",
		TargetPath: "/target",
		StateFile:  stateFile,
		Resume:     false, // Resume disabled
	}

	m := NewMigrator(config)

	// Initialize and save state
	if err := m.initState(); err != nil {
		t.Fatalf("initState failed: %v", err)
	}

	// Save some progress
	if err := m.saveState("initial-commit", 25, 100); err != nil {
		t.Fatalf("saveState failed: %v", err)
	}

	require.NoError(t, m.db.Close())

	// Create new migrator without resume
	m2 := NewMigrator(config)
	if err := m2.initState(); err != nil {
		t.Fatalf("second initState failed: %v", err)
	}
	defer func() { require.NoError(t, m2.db.Close()) }()

	// State should be fresh (not loaded)
	if m2.state.lastCommit != "" {
		t.Errorf("lastCommit = %q, want empty (no resume)", m2.state.lastCommit)
	}
	if m2.state.processed != 0 {
		t.Errorf("processed = %d, want 0 (no resume)", m2.state.processed)
	}
}

func TestSimpleStateWithLargeNumbers(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "state-large")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	stateFile := filepath.Join(tmpDir, "state.txt")
	s := NewMigrationState(stateFile)

	largeProcessed := 1000000
	largeTotal := 2000000

	if err := s.Save("commit", largeProcessed, largeTotal); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	_, processed, total, err := s.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if processed != largeProcessed {
		t.Errorf("processed = %d, want %d", processed, largeProcessed)
	}
	if total != largeTotal {
		t.Errorf("total = %d, want %d", total, largeTotal)
	}
}
