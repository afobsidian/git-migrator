package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

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
