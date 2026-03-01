package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/adamf123git/git-migrator/internal/core"
	"github.com/stretchr/testify/require"
)

func TestLoadSyncConfigFile_Valid(t *testing.T) {
	tmp := t.TempDir()
	cfgPath := filepath.Join(tmp, "sync.yaml")
	content := `git:
  path: /tmp/git-repo
cvs:
  path: /tmp/cvs-repo
  module: mymod
sync:
  direction: bidirectional
  stateFile: /tmp/sync.json
mapping:
  authors:
    alice: "Alice <alice@example.com>"
options:
  dryRun: true
  verbose: false
`
	require.NoError(t, os.WriteFile(cfgPath, []byte(content), 0644))

	cfg, err := loadSyncConfigFile(cfgPath)
	require.NoError(t, err)
	require.Equal(t, "/tmp/git-repo", cfg.Git.Path)
	require.Equal(t, "/tmp/cvs-repo", cfg.CVS.Path)
	require.Equal(t, "mymod", cfg.CVS.Module)
	require.Equal(t, "bidirectional", cfg.Sync.Direction)
	require.Equal(t, "/tmp/sync.json", cfg.Sync.StateFile)
	require.True(t, cfg.Options.DryRun)
	require.Contains(t, cfg.Mapping.Authors, "alice")
}

func TestLoadSyncConfigFile_MissingGitPath(t *testing.T) {
	tmp := t.TempDir()
	cfgPath := filepath.Join(tmp, "bad.yaml")
	content := `cvs:
  path: /tmp/cvs
  module: mod
`
	require.NoError(t, os.WriteFile(cfgPath, []byte(content), 0644))
	_, err := loadSyncConfigFile(cfgPath)
	require.Error(t, err)
	require.Contains(t, err.Error(), "git.path")
}

func TestLoadSyncConfigFile_MissingCVSPath(t *testing.T) {
	tmp := t.TempDir()
	cfgPath := filepath.Join(tmp, "bad.yaml")
	content := `git:
  path: /tmp/git
cvs:
  module: mod
`
	require.NoError(t, os.WriteFile(cfgPath, []byte(content), 0644))
	_, err := loadSyncConfigFile(cfgPath)
	require.Error(t, err)
	require.Contains(t, err.Error(), "cvs.path")
}

func TestLoadSyncConfigFile_MissingCVSModule(t *testing.T) {
	tmp := t.TempDir()
	cfgPath := filepath.Join(tmp, "bad.yaml")
	content := `git:
  path: /tmp/git
cvs:
  path: /tmp/cvs
`
	require.NoError(t, os.WriteFile(cfgPath, []byte(content), 0644))
	_, err := loadSyncConfigFile(cfgPath)
	require.Error(t, err)
	require.Contains(t, err.Error(), "cvs.module")
}

func TestLoadSyncConfigFile_DefaultDirection(t *testing.T) {
	tmp := t.TempDir()
	cfgPath := filepath.Join(tmp, "sync.yaml")
	content := `git:
  path: /tmp/git
cvs:
  path: /tmp/cvs
  module: mod
`
	require.NoError(t, os.WriteFile(cfgPath, []byte(content), 0644))

	cfg, err := loadSyncConfigFile(cfgPath)
	require.NoError(t, err)
	require.Equal(t, string(core.SyncBidirectional), cfg.Sync.Direction)
}

func TestLoadSyncConfigFile_NonExistent(t *testing.T) {
	_, err := loadSyncConfigFile("/nonexistent/path/sync.yaml")
	require.Error(t, err)
}

func TestPrintSyncInfo_DoesNotPanic(t *testing.T) {
	r, w, _ := os.Pipe()
	orig := os.Stdout
	os.Stdout = w

	cfg := &SyncConfigFile{}
	cfg.Git.Path = "/git"
	cfg.CVS.Path = "/cvs"
	cfg.CVS.Module = "mod"
	cfg.Options.Verbose = true
	cfg.Options.DryRun = true
	cfg.Mapping.Authors = map[string]string{"alice": "Alice <alice@example.com>"}

	syncCfg := &core.SyncConfig{
		GitPath:   "/git",
		CVSPath:   "/cvs",
		Direction: core.SyncBidirectional,
	}

	printSyncInfo(cfg, syncCfg)

	_ = w.Close()
	os.Stdout = orig

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	_ = r.Close()

	require.Contains(t, buf.String(), "/git")
	require.Contains(t, buf.String(), "bidirectional")
}

// createSyncTestGitRepo creates a minimal Git repo at a temp path and returns it.
func createSyncTestGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	repo, err := gogit.PlainInit(dir, false)
	require.NoError(t, err)
	w, err := repo.Worktree()
	require.NoError(t, err)
	f := filepath.Join(dir, "README.md")
	require.NoError(t, os.WriteFile(f, []byte("hello"), 0644))
	_, err = w.Add("README.md")
	require.NoError(t, err)
	_, err = w.Commit("initial commit", &gogit.CommitOptions{
		Author: &object.Signature{
			Name:  "Test",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	require.NoError(t, err)
	return dir
}

// createSyncTestCVSRepo creates a minimal CVS CVSROOT structure.
func createSyncTestCVSRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "CVSROOT"), 0755))
	return dir
}

// TestRunSync_DryRun_GitToCVS calls runSync end-to-end with a valid config
// pointing at real repos in dry-run mode.  No CVS binary or CVS commits are
// required because DryRun=true.
func TestRunSync_DryRun_GitToCVS(t *testing.T) {
	gitDir := createSyncTestGitRepo(t)
	cvsDir := createSyncTestCVSRepo(t)

	tmp := t.TempDir()
	cfgPath := filepath.Join(tmp, "sync.yaml")
	content := "git:\n  path: " + gitDir + "\ncvs:\n  path: " + cvsDir + "\n  module: mod\nsync:\n  direction: git-to-cvs\noptions:\n  dryRun: true\n  verbose: true\n"
	require.NoError(t, os.WriteFile(cfgPath, []byte(content), 0644))

	// Back up and restore flag values
	origCfg := syncConfigFile
	origDry := syncDryRun
	origVerbose := syncVerbose
	origDir := syncDirection
	defer func() {
		syncConfigFile = origCfg
		syncDryRun = origDry
		syncVerbose = origVerbose
		syncDirection = origDir
	}()

	syncConfigFile = cfgPath
	syncDryRun = false   // comes from config file
	syncVerbose = false
	syncDirection = ""

	err := runSync(nil, nil)
	require.NoError(t, err)
}

// TestRunSync_DryRun_FlagOverrides verifies that CLI flags override config values.
func TestRunSync_DryRun_FlagOverrides(t *testing.T) {
	gitDir := createSyncTestGitRepo(t)
	cvsDir := createSyncTestCVSRepo(t)

	tmp := t.TempDir()
	cfgPath := filepath.Join(tmp, "sync.yaml")
	content := "git:\n  path: " + gitDir + "\ncvs:\n  path: " + cvsDir + "\n  module: mod\nsync:\n  direction: cvs-to-git\n"
	require.NoError(t, os.WriteFile(cfgPath, []byte(content), 0644))

	origCfg := syncConfigFile
	origDry := syncDryRun
	origVerbose := syncVerbose
	origDir := syncDirection
	defer func() {
		syncConfigFile = origCfg
		syncDryRun = origDry
		syncVerbose = origVerbose
		syncDirection = origDir
	}()

	syncConfigFile = cfgPath
	syncDryRun = true    // override: dry-run via flag
	syncVerbose = true   // override: verbose via flag
	syncDirection = "git-to-cvs" // override direction

	err := runSync(nil, nil)
	require.NoError(t, err)
}

// TestRunSync_InvalidConfig ensures runSync returns an error for a bad config.
func TestRunSync_InvalidConfig(t *testing.T) {
	origCfg := syncConfigFile
	defer func() { syncConfigFile = origCfg }()

	syncConfigFile = "/nonexistent/sync.yaml"
	err := runSync(nil, nil)
	require.Error(t, err)
}

// TestRunSync_SyncerRunFails covers the path where syncer.Run() returns an
// error (valid YAML config but unreachable repositories).
func TestRunSync_SyncerRunFails(t *testing.T) {
	tmp := t.TempDir()
	cfgPath := filepath.Join(tmp, "sync.yaml")
	content := "git:\n  path: /nonexistent/git\ncvs:\n  path: /nonexistent/cvs\n  module: mod\nsync:\n  direction: cvs-to-git\n"
	require.NoError(t, os.WriteFile(cfgPath, []byte(content), 0644))

	origCfg := syncConfigFile
	origDry := syncDryRun
	origVerbose := syncVerbose
	origDir := syncDirection
	defer func() {
		syncConfigFile = origCfg
		syncDryRun = origDry
		syncVerbose = origVerbose
		syncDirection = origDir
	}()

	syncConfigFile = cfgPath
	syncDryRun = false
	syncVerbose = false
	syncDirection = ""

	err := runSync(nil, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "sync failed")
}

// TestLoadSyncConfigFile_InvalidYAML tests that malformed YAML returns an error.
func TestLoadSyncConfigFile_InvalidYAML(t *testing.T) {
	tmp := t.TempDir()
	cfgPath := filepath.Join(tmp, "invalid.yaml")
	require.NoError(t, os.WriteFile(cfgPath, []byte(":\ninvalid::\n  bad"), 0644))
	_, err := loadSyncConfigFile(cfgPath)
	require.Error(t, err)
}
