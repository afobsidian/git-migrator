package storage

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStateDB_History_Delete_Close(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "state.db")
	sdb, err := NewStateDB(dbPath)
	require.NoError(t, err)

	// Save two states
	st1 := &MigrationState{MigrationID: "m1", LastCommit: "r1", Processed: 1, Total: 2, SourcePath: "/s", TargetPath: "/t", Status: "in_progress"}
	st2 := &MigrationState{MigrationID: "m2", LastCommit: "r2", Processed: 2, Total: 3, SourcePath: "/s", TargetPath: "/t", Status: "in_progress"}
	require.NoError(t, sdb.Save(st1))
	require.NoError(t, sdb.Save(st2))

	hist, err := sdb.History()
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(hist), 2)

	// Delete one
	require.NoError(t, sdb.Delete("m1"))

	// After delete, loading m1 should return an error
	_, err = sdb.Load("m1")
	require.Error(t, err)

	// Close DB
	require.NoError(t, sdb.Close())
}
