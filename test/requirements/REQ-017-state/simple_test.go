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
	defer db.Close()
}