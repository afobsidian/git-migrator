package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewStateDB(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "statedb-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := NewStateDB(dbPath)
	if err != nil {
		t.Fatalf("NewStateDB failed: %v", err)
	}
	defer func() { require.NoError(t, db.Close()) }()

	// Check that the database file was created
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("Database file was not created")
	}
}

func TestNewStateDBCreatesParentDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "statedb-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	// Use a nested directory that doesn't exist
	dbPath := filepath.Join(tmpDir, "nested", "deep", "test.db")
	db, err := NewStateDB(dbPath)
	if err != nil {
		t.Fatalf("NewStateDB failed: %v", err)
	}
	defer func() { require.NoError(t, db.Close()) }()

	// Check that the parent directory was created
	parentDir := filepath.Dir(dbPath)
	if _, err := os.Stat(parentDir); os.IsNotExist(err) {
		t.Error("Parent directory was not created")
	}
}

func TestNewStateDBInvalidPath(t *testing.T) {
	// Try to create a database in a path that can't be created
	// For example, trying to create under /proc which should fail
	_, err := NewStateDB("/proc/nonexistent/test.db")
	if err == nil {
		t.Error("NewStateDB should fail with invalid path")
	}
}

func TestStateDBSave(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "statedb-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := NewStateDB(dbPath)
	if err != nil {
		t.Fatalf("NewStateDB failed: %v", err)
	}
	defer func() { require.NoError(t, db.Close()) }()

	state := &MigrationState{
		MigrationID: "test-migration-1",
		LastCommit:  "abc123",
		Processed:   50,
		Total:       100,
		SourcePath:  "/source/path",
		TargetPath:  "/target/path",
		Status:      "in_progress",
	}

	if err := db.Save(state); err != nil {
		t.Fatalf("Save failed: %v", err)
	}
}

func TestStateDBSaveAndUpdate(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "statedb-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := NewStateDB(dbPath)
	if err != nil {
		t.Fatalf("NewStateDB failed: %v", err)
	}
	defer func() { require.NoError(t, db.Close()) }()

	// Save initial state
	state := &MigrationState{
		MigrationID: "test-migration-1",
		LastCommit:  "abc123",
		Processed:   50,
		Total:       100,
		SourcePath:  "/source/path",
		TargetPath:  "/target/path",
		Status:      "in_progress",
	}
	if err := db.Save(state); err != nil {
		t.Fatalf("First save failed: %v", err)
	}

	// Update and save again (should use INSERT OR REPLACE)
	state.Processed = 75
	state.LastCommit = "def456"
	if err := db.Save(state); err != nil {
		t.Fatalf("Second save failed: %v", err)
	}

	// Load and verify the updated state
	loaded, err := db.Load("test-migration-1")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.Processed != 75 {
		t.Errorf("Processed = %d, want 75", loaded.Processed)
	}
	if loaded.LastCommit != "def456" {
		t.Errorf("LastCommit = %q, want %q", loaded.LastCommit, "def456")
	}
}

func TestStateDBLoad(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "statedb-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := NewStateDB(dbPath)
	if err != nil {
		t.Fatalf("NewStateDB failed: %v", err)
	}
	defer func() { require.NoError(t, db.Close()) }()

	// Save a state first
	state := &MigrationState{
		MigrationID: "test-migration-2",
		LastCommit:  "xyz789",
		Processed:   25,
		Total:       200,
		SourcePath:  "/cvs/repo",
		TargetPath:  "/git/repo",
		Status:      "in_progress",
	}
	if err := db.Save(state); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Load the state
	loaded, err := db.Load("test-migration-2")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.MigrationID != "test-migration-2" {
		t.Errorf("MigrationID = %q, want %q", loaded.MigrationID, "test-migration-2")
	}
	if loaded.LastCommit != "xyz789" {
		t.Errorf("LastCommit = %q, want %q", loaded.LastCommit, "xyz789")
	}
	if loaded.Processed != 25 {
		t.Errorf("Processed = %d, want 25", loaded.Processed)
	}
	if loaded.Total != 200 {
		t.Errorf("Total = %d, want 200", loaded.Total)
	}
	if loaded.SourcePath != "/cvs/repo" {
		t.Errorf("SourcePath = %q, want %q", loaded.SourcePath, "/cvs/repo")
	}
	if loaded.TargetPath != "/git/repo" {
		t.Errorf("TargetPath = %q, want %q", loaded.TargetPath, "/git/repo")
	}
	if loaded.Status != "in_progress" {
		t.Errorf("Status = %q, want %q", loaded.Status, "in_progress")
	}
	if loaded.LastUpdated.IsZero() {
		t.Error("LastUpdated should not be zero")
	}
}

func TestStateDBLoadNotFound(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "statedb-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := NewStateDB(dbPath)
	if err != nil {
		t.Fatalf("NewStateDB failed: %v", err)
	}
	defer func() { require.NoError(t, db.Close()) }()

	// Try to load a non-existent migration
	_, err = db.Load("nonexistent")
	if err == nil {
		t.Error("Load should fail for non-existent migration")
	}
}

func TestStateDBComplete(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "statedb-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := NewStateDB(dbPath)
	if err != nil {
		t.Fatalf("NewStateDB failed: %v", err)
	}
	defer func() { require.NoError(t, db.Close()) }()

	// Save a state first
	state := &MigrationState{
		MigrationID: "test-migration-3",
		LastCommit:  "final123",
		Processed:   100,
		Total:       100,
		SourcePath:  "/source",
		TargetPath:  "/target",
		Status:      "in_progress",
	}
	if err := db.Save(state); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Mark as complete
	if err := db.Complete("test-migration-3"); err != nil {
		t.Fatalf("Complete failed: %v", err)
	}

	// Verify status is updated
	loaded, err := db.Load("test-migration-3")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.Status != "completed" {
		t.Errorf("Status = %q, want %q", loaded.Status, "completed")
	}
}

func TestStateDBCompleteNonExistent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "statedb-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := NewStateDB(dbPath)
	if err != nil {
		t.Fatalf("NewStateDB failed: %v", err)
	}
	defer func() { require.NoError(t, db.Close()) }()

	// Complete on non-existent migration should not fail (UPDATE just affects 0 rows)
	if err := db.Complete("nonexistent"); err != nil {
		t.Errorf("Complete on non-existent should not fail: %v", err)
	}
}

func TestStateDBDelete(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "statedb-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := NewStateDB(dbPath)
	if err != nil {
		t.Fatalf("NewStateDB failed: %v", err)
	}
	defer func() { require.NoError(t, db.Close()) }()

	// Save a state first
	state := &MigrationState{
		MigrationID: "test-migration-4",
		LastCommit:  "delete123",
		Processed:   50,
		Total:       100,
		SourcePath:  "/source",
		TargetPath:  "/target",
		Status:      "in_progress",
	}
	if err := db.Save(state); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Delete it
	if err := db.Delete("test-migration-4"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify it's gone
	_, err = db.Load("test-migration-4")
	if err == nil {
		t.Error("Load should fail after Delete")
	}
}

func TestStateDBDeleteNonExistent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "statedb-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := NewStateDB(dbPath)
	if err != nil {
		t.Fatalf("NewStateDB failed: %v", err)
	}
	defer func() { require.NoError(t, db.Close()) }()

	// Delete on non-existent should not fail
	if err := db.Delete("nonexistent"); err != nil {
		t.Errorf("Delete on non-existent should not fail: %v", err)
	}
}

func TestStateDBHistory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "statedb-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := NewStateDB(dbPath)
	if err != nil {
		t.Fatalf("NewStateDB failed: %v", err)
	}
	defer func() { require.NoError(t, db.Close()) }()

	// Save multiple states
	states := []*MigrationState{
		{
			MigrationID: "migration-1",
			LastCommit:  "commit1",
			Processed:   100,
			Total:       100,
			SourcePath:  "/src1",
			TargetPath:  "/tgt1",
			Status:      "completed",
		},
		{
			MigrationID: "migration-2",
			LastCommit:  "commit2",
			Processed:   50,
			Total:       100,
			SourcePath:  "/src2",
			TargetPath:  "/tgt2",
			Status:      "in_progress",
		},
		{
			MigrationID: "migration-3",
			LastCommit:  "commit3",
			Processed:   0,
			Total:       200,
			SourcePath:  "/src3",
			TargetPath:  "/tgt3",
			Status:      "pending",
		},
	}

	// Small delay to ensure different timestamps
	for i, s := range states {
		if i > 0 {
			time.Sleep(10 * time.Millisecond)
		}
		if err := db.Save(s); err != nil {
			t.Fatalf("Save failed: %v", err)
		}
	}

	// Get history
	history, err := db.History()
	if err != nil {
		t.Fatalf("History failed: %v", err)
	}

	if len(history) != 3 {
		t.Errorf("History returned %d entries, want 3", len(history))
	}

	// History should be ordered by last_updated DESC (newest first)
	if len(history) >= 2 {
		if history[0].LastUpdated.Before(history[1].LastUpdated) {
			t.Error("History should be ordered by last_updated DESC")
		}
	}
}

func TestStateDBHistoryEmpty(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "statedb-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := NewStateDB(dbPath)
	if err != nil {
		t.Fatalf("NewStateDB failed: %v", err)
	}
	defer func() { require.NoError(t, db.Close()) }()

	// Get history from empty database
	history, err := db.History()
	if err != nil {
		t.Fatalf("History failed: %v", err)
	}

	// nil slice is valid for empty in Go (len(nil slice) == 0)
	if len(history) != 0 {
		t.Errorf("History returned %d entries, want 0", len(history))
	}
}

func TestStateDBClose(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "statedb-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := NewStateDB(dbPath)
	if err != nil {
		t.Fatalf("NewStateDB failed: %v", err)
	}

	// Close should not fail
	if err := db.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}

	// Double close should be safe (but might fail on the underlying DB)
	_ = db.Close()
}

func TestStateDBSaveSetsTimestamp(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "statedb-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := NewStateDB(dbPath)
	if err != nil {
		t.Fatalf("NewStateDB failed: %v", err)
	}
	defer func() { require.NoError(t, db.Close()) }()

	before := time.Now()
	time.Sleep(1 * time.Millisecond)

	state := &MigrationState{
		MigrationID: "timestamp-test",
		LastCommit:  "ts123",
		Processed:   1,
		Total:       10,
		Status:      "in_progress",
	}
	if err := db.Save(state); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	time.Sleep(1 * time.Millisecond)
	after := time.Now()

	loaded, err := db.Load("timestamp-test")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.LastUpdated.Before(before) {
		t.Error("LastUpdated should be after Save was called")
	}
	if loaded.LastUpdated.After(after) {
		t.Error("LastUpdated should be before Save returned")
	}
}

func TestStateDBConcurrentAccess(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "statedb-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := NewStateDB(dbPath)
	if err != nil {
		t.Fatalf("NewStateDB failed: %v", err)
	}
	defer func() { require.NoError(t, db.Close()) }()

	// Create initial states
	for i := 0; i < 5; i++ {
		state := &MigrationState{
			MigrationID: string(rune('A' + i)),
			Processed:   0,
			Total:       100,
			Status:      "in_progress",
		}
		if err := db.Save(state); err != nil {
			t.Fatalf("Initial save failed: %v", err)
		}
	}

	// Concurrent updates
	done := make(chan bool)
	for i := 0; i < 5; i++ {
		go func(idx int) {
			id := string(rune('A' + idx))
			for j := 0; j < 10; j++ {
				state := &MigrationState{
					MigrationID: id,
					Processed:   (j + 1) * 10,
					Total:       100,
					Status:      "in_progress",
				}
				_ = db.Save(state)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 5; i++ {
		<-done
	}

	// Verify final states
	for i := 0; i < 5; i++ {
		id := string(rune('A' + i))
		loaded, err := db.Load(id)
		if err != nil {
			t.Errorf("Load failed for %s: %v", id, err)
			continue
		}
		if loaded.Processed != 100 {
			t.Errorf("Migration %s: Processed = %d, want 100", id, loaded.Processed)
		}
	}
}

func TestStateDBEmptyStrings(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "statedb-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := NewStateDB(dbPath)
	if err != nil {
		t.Fatalf("NewStateDB failed: %v", err)
	}
	defer func() { require.NoError(t, db.Close()) }()

	// Save with empty string fields
	state := &MigrationState{
		MigrationID: "empty-strings-test",
		LastCommit:  "",
		Processed:   0,
		Total:       0,
		SourcePath:  "",
		TargetPath:  "",
		Status:      "",
	}
	if err := db.Save(state); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := db.Load("empty-strings-test")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.LastCommit != "" {
		t.Errorf("LastCommit = %q, want empty", loaded.LastCommit)
	}
	if loaded.SourcePath != "" {
		t.Errorf("SourcePath = %q, want empty", loaded.SourcePath)
	}
	if loaded.TargetPath != "" {
		t.Errorf("TargetPath = %q, want empty", loaded.TargetPath)
	}
	if loaded.Status != "" {
		t.Errorf("Status = %q, want empty", loaded.Status)
	}
}

func TestStateDBZeroValues(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "statedb-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := NewStateDB(dbPath)
	if err != nil {
		t.Fatalf("NewStateDB failed: %v", err)
	}
	defer func() { require.NoError(t, db.Close()) }()

	// Save with zero numeric values
	state := &MigrationState{
		MigrationID: "zero-values-test",
		LastCommit:  "zero",
		Processed:   0,
		Total:       0,
		SourcePath:  "/src",
		TargetPath:  "/tgt",
		Status:      "pending",
	}
	if err := db.Save(state); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := db.Load("zero-values-test")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.Processed != 0 {
		t.Errorf("Processed = %d, want 0", loaded.Processed)
	}
	if loaded.Total != 0 {
		t.Errorf("Total = %d, want 0", loaded.Total)
	}
}

func TestStateDBLargeValues(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "statedb-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := NewStateDB(dbPath)
	if err != nil {
		t.Fatalf("NewStateDB failed: %v", err)
	}
	defer func() { require.NoError(t, db.Close()) }()

	// Save with large numeric values
	state := &MigrationState{
		MigrationID: "large-values-test",
		LastCommit:  "large",
		Processed:   1000000,
		Total:       1000000,
		SourcePath:  "/very/long/path/to/source/repository",
		TargetPath:  "/equally/long/path/to/target/repository",
		Status:      "completed",
	}
	if err := db.Save(state); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := db.Load("large-values-test")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.Processed != 1000000 {
		t.Errorf("Processed = %d, want 1000000", loaded.Processed)
	}
	if loaded.Total != 1000000 {
		t.Errorf("Total = %d, want 1000000", loaded.Total)
	}
}

func TestStateDBMigrationStateFields(t *testing.T) {
	// Test the MigrationState struct directly
	now := time.Now()
	state := &MigrationState{
		MigrationID: "test-id",
		LastCommit:  "commit123",
		Processed:   50,
		Total:       100,
		SourcePath:  "/source",
		TargetPath:  "/target",
		LastUpdated: now,
		Status:      "in_progress",
	}

	if state.MigrationID != "test-id" {
		t.Errorf("MigrationID = %q, want %q", state.MigrationID, "test-id")
	}
	if state.LastCommit != "commit123" {
		t.Errorf("LastCommit = %q, want %q", state.LastCommit, "commit123")
	}
	if state.Processed != 50 {
		t.Errorf("Processed = %d, want 50", state.Processed)
	}
	if state.Total != 100 {
		t.Errorf("Total = %d, want 100", state.Total)
	}
	if state.SourcePath != "/source" {
		t.Errorf("SourcePath = %q, want %q", state.SourcePath, "/source")
	}
	if state.TargetPath != "/target" {
		t.Errorf("TargetPath = %q, want %q", state.TargetPath, "/target")
	}
	if !state.LastUpdated.Equal(now) {
		t.Errorf("LastUpdated = %v, want %v", state.LastUpdated, now)
	}
	if state.Status != "in_progress" {
		t.Errorf("Status = %q, want %q", state.Status, "in_progress")
	}
}
