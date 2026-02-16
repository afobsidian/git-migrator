package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/adamf123git/git-migrator/internal/core"
	"github.com/stretchr/testify/require"
)

func TestLoadConfigFile_Valid(t *testing.T) {
	// Create a temporary config file
	tmp := t.TempDir()
	cfgPath := filepath.Join(tmp, "cfg.yaml")
	content := `source:
  type: cvs
  path: /tmp/src
target:
  path: /tmp/target
mapping:
  authors: {}
options:
  dryRun: true
  verbose: true
  chunkSize: 10
`
	require.NoError(t, os.WriteFile(cfgPath, []byte(content), 0644))

	cfg, err := loadConfigFile(cfgPath)
	require.NoError(t, err)
	require.Equal(t, "cvs", cfg.Source.Type)
	require.Equal(t, "/tmp/src", cfg.Source.Path)
	require.Equal(t, "/tmp/target", cfg.Target.Path)
	require.Equal(t, 10, cfg.Options.ChunkSize)
}

func TestLoadConfigFile_MissingFields(t *testing.T) {
	tmp := t.TempDir()
	cfgPath := filepath.Join(tmp, "bad.yaml")
	// Missing target.path
	content := `source:
  type: cvs
  path: /tmp/src
`
	require.NoError(t, os.WriteFile(cfgPath, []byte(content), 0644))

	_, err := loadConfigFile(cfgPath)
	require.Error(t, err)
}

func TestPrintMigrationInfo_DoesNotPanic(t *testing.T) {
	buf := &bytes.Buffer{}
	// Temporarily redirect stdout
	orig := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cfg := &ConfigFile{}
	cfg.Source.Type = "cvs"
	cfg.Source.Path = "/src"
	cfg.Target.Path = "/t"
	cfg.Options.DryRun = true
	cfg.Options.Verbose = true
	cfg.Mapping.Authors = map[string]string{"a": "b"}

	mig := &core.MigrationConfig{SourcePath: "/src", TargetPath: "/t", AuthorMap: map[string]string{"a": "b"}}

	// Call the function
	printMigrationInfo(cfg, mig)

	// Restore stdout
	_ = w.Close()
	os.Stdout = orig
	buf.ReadFrom(r)
}

func TestRunAnalyze_InvalidType(t *testing.T) {
	// Backup and restore flags
	oldType := analyzeSourceType
	oldSource := analyzeSource
	analyzeSourceType = "invalid"
	analyzeSource = ""
	defer func() {
		analyzeSourceType = oldType
		analyzeSource = oldSource
	}()

	err := runAnalyze(nil, nil)
	require.Error(t, err)
}

func TestRunAuthorsExtract_InvalidFormat(t *testing.T) {
	oldSource := authorsSource
	oldFormat := authorsFormat
	authorsSource = "/tmp"
	authorsFormat = "xml"
	defer func() {
		authorsSource = oldSource
		authorsFormat = oldFormat
	}()

	err := runAuthorsExtract(nil, nil)
	require.Error(t, err)
}
