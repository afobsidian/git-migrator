package requirements

import (
	"testing"

	"github.com/adamf123git/git-migrator/internal/storage"
)

func TestSimpleDBConnection(t *testing.T) {
	db, err := storage.NewStateDB("/tmp/simple_state_test.db")
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("Warning: failed to close database: %v", err)
		}
	}()
}
