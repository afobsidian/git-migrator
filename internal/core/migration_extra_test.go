package core

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/adamf123git/git-migrator/internal/progress"
	"github.com/adamf123git/git-migrator/internal/storage"
	"github.com/adamf123git/git-migrator/internal/vcs"
	"github.com/adamf123git/git-migrator/internal/vcs/git"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// lightweight empty iterator for tests
type emptyIter struct{}

func (e *emptyIter) Next() bool          { return false }
func (e *emptyIter) Commit() *vcs.Commit { return nil }
func (e *emptyIter) Err() error          { return nil }

type mockSource struct {
	branches []string
	tags     map[string]string
}

func (m *mockSource) Validate() error                         { return nil }
func (m *mockSource) GetCommits() (vcs.CommitIterator, error) { return &emptyIter{}, nil }
func (m *mockSource) GetBranches() ([]string, error)          { return m.branches, nil }
func (m *mockSource) GetTags() (map[string]string, error)     { return m.tags, nil }
func (m *mockSource) Close() error                            { return nil }

func TestCreateBranches_TargetErrorsAndSuccess(t *testing.T) {
	// Error path: target writer not initialized -> CreateBranch returns error but should not bubble up
	m := &Migrator{
		config:   &MigrationConfig{BranchMap: map[string]string{}},
		source:   &mockSource{branches: []string{"b1"}},
		target:   git.NewWriter(), // uninitialized writer will error on CreateBranch
		reporter: progress.NewReporter(0),
	}

	require.NoError(t, m.createBranches())

	// Success path: initialize repository and ensure branch is created (mapped)
	tmp := t.TempDir()
	w := git.NewWriter()
	require.NoError(t, w.Init(tmp))

	// Create an initial commit so HEAD exists (branches need a commit to point to)
	initialCommit := &vcs.Commit{
		Revision: "initial",
		Author:   "test",
		Email:    "test@example.com",
		Date:     time.Now(),
		Message:  "Initial commit",
		Files: []vcs.FileChange{
			{Path: "README.md", Action: vcs.ActionAdd, Content: []byte("# Test")},
		},
	}
	require.NoError(t, w.ApplyCommit(initialCommit))

	m2 := &Migrator{
		config:   &MigrationConfig{BranchMap: map[string]string{"b1": "mapped"}},
		source:   &mockSource{branches: []string{"b1"}},
		target:   w,
		reporter: progress.NewReporter(0),
	}

	require.NoError(t, m2.createBranches())

	branches, err := w.ListBranches()
	require.NoError(t, err)
	assert.Contains(t, branches, "mapped")
}

func TestCreateTags_TargetErrorsAndSuccess(t *testing.T) {
	// Error path: uninitialized target
	m := &Migrator{
		config:   &MigrationConfig{TagMap: map[string]string{}},
		source:   &mockSource{tags: map[string]string{"v1": "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef"}},
		target:   git.NewWriter(),
		reporter: progress.NewReporter(0),
	}
	require.NoError(t, m.createTags())

	// Success path: initialized repo
	tmp := t.TempDir()
	w := git.NewWriter()
	require.NoError(t, w.Init(tmp))

	m2 := &Migrator{
		config:   &MigrationConfig{TagMap: map[string]string{"v1": "tagged"}},
		source:   &mockSource{tags: map[string]string{"v1": "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef"}},
		target:   w,
		reporter: progress.NewReporter(0),
	}

	require.NoError(t, m2.createTags())
	tags, err := w.ListTags()
	require.NoError(t, err)
	// mapped name should be present
	_, ok := tags["tagged"]
	assert.True(t, ok)
}

func TestMarkCompleteAndSaveState(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "state.db")
	sdb, err := storage.NewStateDB(dbPath)
	require.NoError(t, err)
	defer sdb.Close()

	m := &Migrator{
		config:   &MigrationConfig{SourcePath: "/src", TargetPath: "/t"},
		state:    &MigrationState{migrationID: "mid", lastCommit: "r1", processed: 2, total: 5},
		db:       sdb,
		reporter: progress.NewReporter(0),
	}

	// saveState should persist without error
	require.NoError(t, m.saveState("r1", 2, 5))

	// markComplete should set status to completed
	require.NoError(t, m.markComplete())

	st, err := sdb.Load(m.state.migrationID)
	require.NoError(t, err)
	require.Equal(t, "completed", st.Status)
	require.Equal(t, "r1", st.LastCommit)
	require.Equal(t, 2, st.Processed)
	require.Equal(t, 5, st.Total)
}

func TestSimpleState_SaveLoadClear(t *testing.T) {
	path := filepath.Join(t.TempDir(), "statefile.txt")
	s := NewMigrationState(path)

	require.NoError(t, s.Save("c1", 3, 7))

	c, p, tot, err := s.Load()
	require.NoError(t, err)
	require.Equal(t, "c1", c)
	require.Equal(t, 3, p)
	require.Equal(t, 7, tot)

	// Clear should remove file
	require.NoError(t, s.Clear())
	_, err = os.Stat(path)
	require.True(t, os.IsNotExist(err))
}
