package requirements

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/adamf123git/git-migrator/internal/storage"
)

// TestStateDatabase tests SQLite state database
func TestStateDatabase(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "state.db")

	db, err := storage.NewStateDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create state DB: %v", err)
	}
	defer db.Close()

	// Verify database file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("Expected database file to be created")
	}
}

// TestStateSaveLoad tests saving and loading state
func TestStateSaveLoad(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "state.db")

	db, err := storage.NewStateDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create state DB: %v", err)
	}
	defer db.Close()

	// Save state
	state := &storage.MigrationState{
		MigrationID:   "test-migration-1",
		LastCommit:    "abc123",
		Processed:     5,
		Total:         10,
		SourcePath:    "/path/to/cvs",
		TargetPath:    "/path/to/git",
		LastUpdated:   time.Now(),
		Status:        "in_progress",
	}

	if err := db.Save(state); err != nil {
		t.Fatalf("Failed to save state: %v", err)
	}

	// Load state
	loaded, err := db.Load("test-migration-1")
	if err != nil {
		t.Fatalf("Failed to load state: %v", err)
	}

	if loaded.LastCommit != "abc123" {
		t.Errorf("Expected LastCommit=abc123, got %s", loaded.LastCommit)
	}
	if loaded.Processed != 5 {
		t.Errorf("Expected Processed=5, got %d", loaded.Processed)
	}
	if loaded.Total != 10 {
		t.Errorf("Expected Total=10, got %d", loaded.Total)
	}
}

// TestStateUpdate tests updating existing state
func TestStateUpdate(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "state.db")

	db, err := storage.NewStateDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create state DB: %v", err)
	}
	defer db.Close()

	// Initial save
	state := &storage.MigrationState{
		MigrationID: "test-migration-2",
		LastCommit:  "commit1",
		Processed:   1,
		Total:       10,
		LastUpdated: time.Now(),
		Status:      "in_progress",
	}
	if err := db.Save(state); err != nil {
		t.Fatalf("Failed to save state: %v", err)
	}

	// Update state
	state.LastCommit = "commit2"
	state.Processed = 2
	if err := db.Save(state); err != nil {
		t.Fatalf("Failed to update state: %v", err)
	}

	// Verify update
	loaded, err := db.Load("test-migration-2")
	if err != nil {
		t.Fatalf("Failed to load state: %v", err)
	}
	if loaded.LastCommit != "commit2" {
		t.Errorf("Expected LastCommit=commit2, got %s", loaded.LastCommit)
	}
	if loaded.Processed != 2 {
		t.Errorf("Expected Processed=2, got %d", loaded.Processed)
	}
}

// TestStateComplete tests marking migration complete
func TestStateComplete(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "state.db")

	db, err := storage.NewStateDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create state DB: %v", err)
	}
	defer db.Close()

	// Save state
	state := &storage.MigrationState{
		MigrationID: "test-migration-3",
		LastCommit:  "final",
		Processed:   10,
		Total:       10,
		LastUpdated: time.Now(),
		Status:      "in_progress",
	}
	if err := db.Save(state); err != nil {
		t.Fatalf("Failed to save state: %v", err)
	}

	// Mark complete
	if err := db.Complete("test-migration-3"); err != nil {
		t.Fatalf("Failed to complete migration: %v", err)
	}

	// Verify status
	loaded, err := db.Load("test-migration-3")
	if err != nil {
		t.Fatalf("Failed to load state: %v", err)
	}
	if loaded.Status != "completed" {
		t.Errorf("Expected status=completed, got %s", loaded.Status)
	}
}

// TestStateHistory tests migration history
func TestStateHistory(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "state.db")

	db, err := storage.NewStateDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create state DB: %v", err)
	}
	defer db.Close()

	// Create multiple migrations
	for i := 1; i <= 3; i++ {
		state := &storage.MigrationState{
			MigrationID:   string(rune('a' + i)),
			LastCommit:    "commit",
			Processed:     i,
			Total:         10,
			LastUpdated:   time.Now(),
			Status:        "completed",
		}
		if err := db.Save(state); err != nil {
			t.Fatalf("Failed to save state %d: %v", i, err)
		}
	}

	// Query history
	history, err := db.History()
	if err != nil {
		t.Fatalf("Failed to get history: %v", err)
	}

	if len(history) < 3 {
		t.Errorf("Expected at least 3 history entries, got %d", len(history))
	}
}

// TestStateDelete tests deleting state
func TestStateDelete(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "state.db")

	db, err := storage.NewStateDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create state DB: %v", err)
	}
	defer db.Close()

	// Save state
	state := &storage.MigrationState{
		MigrationID: "to-delete",
		LastCommit:  "commit",
		Processed:   1,
		Total:       10,
		LastUpdated: time.Now(),
		Status:      "in_progress",
	}
	if err := db.Save(state); err != nil {
		t.Fatalf("Failed to save state: %v", err)
	}

	// Delete state
	if err := db.Delete("to-delete"); err != nil {
		t.Fatalf("Failed to delete state: %v", err)
	}

	// Verify deleted
	_, err = db.Load("to-delete")
	if err == nil {
		t.Error("Expected error loading deleted state")
	}
}
