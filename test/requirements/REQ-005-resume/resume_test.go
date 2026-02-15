package requirements

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/adamf123git/git-migrator/internal/core"
)

// TestStateSave tests saving migration state
func TestStateSave(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "migration.db")

	state := core.NewMigrationState(stateFile)
	defer state.Close()

	// Save state
	err := state.Save("testcommit123", 5, 10)
	if err != nil {
		t.Fatalf("Failed to save state: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(stateFile); os.IsNotExist(err) {
		t.Error("Expected state file to be created")
	}
}

// TestStateLoad tests loading migration state
func TestStateLoad(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "migration.db")

	state := core.NewMigrationState(stateFile)
	defer state.Close()

	// Save state
	if err := state.Save("testcommit456", 3, 10); err != nil {
		t.Fatalf("Failed to save state: %v", err)
	}

	// Close and reopen
	state.Close()
	state = core.NewMigrationState(stateFile)
	defer state.Close()

	// Load state
	commit, processed, total, err := state.Load()
	if err != nil {
		t.Fatalf("Failed to load state: %v", err)
	}

	if commit != "testcommit456" {
		t.Errorf("Expected commit 'testcommit456', got %q", commit)
	}
	if processed != 3 {
		t.Errorf("Expected processed=3, got %d", processed)
	}
	if total != 10 {
		t.Errorf("Expected total=10, got %d", total)
	}
}

// TestStateNoState tests loading when no state exists
func TestStateNoState(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "nonexistent.db")

	state := core.NewMigrationState(stateFile)
	defer state.Close()

	// Should return empty state
	commit, processed, total, err := state.Load()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if commit != "" {
		t.Errorf("Expected empty commit, got %q", commit)
	}
	if processed != 0 {
		t.Errorf("Expected processed=0, got %d", processed)
	}
	if total != 0 {
		t.Errorf("Expected total=0, got %d", total)
	}
}

// TestStateClear tests clearing migration state
func TestStateClear(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "migration.db")

	state := core.NewMigrationState(stateFile)

	// Save state
	if err := state.Save("testcommit789", 7, 10); err != nil {
		t.Fatalf("Failed to save state: %v", err)
	}

	// Clear state
	if err := state.Clear(); err != nil {
		t.Fatalf("Failed to clear state: %v", err)
	}

	// Load should return empty
	commit, _, _, err := state.Load()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if commit != "" {
		t.Errorf("Expected empty state after clear, got commit %q", commit)
	}

	state.Close()
}

// TestResumeMigration tests resuming a migration
func TestResumeMigration(t *testing.T) {
	t.Skip("Integration test - requires CVS fixtures")
}

// TestResumeNoDuplicates tests that resume doesn't create duplicates
func TestResumeNoDuplicates(t *testing.T) {
	t.Skip("Integration test - requires CVS fixtures")
}
