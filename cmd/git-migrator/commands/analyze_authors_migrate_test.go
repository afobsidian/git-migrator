package commands

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func makeEmptyCVSRepo(t *testing.T) string {
	dir := t.TempDir()
	cvsroot := filepath.Join(dir, "CVSROOT")
	require.NoError(t, os.MkdirAll(cvsroot, 0755))
	// create optional files
	_ = os.WriteFile(filepath.Join(cvsroot, "history"), []byte(""), 0644)
	return dir
}

func TestRunAnalyze_SuccessEmptyRepo(t *testing.T) {
	dir := makeEmptyCVSRepo(t)

	oldType := analyzeSourceType
	oldSource := analyzeSource
	analyzeSourceType = "cvs"
	analyzeSource = dir
	defer func() { analyzeSourceType = oldType; analyzeSource = oldSource }()

	err := runAnalyze(nil, nil)
	require.NoError(t, err)
}

func TestRunAuthorsExtract_SuccessEmptyRepo(t *testing.T) {
	dir := makeEmptyCVSRepo(t)

	oldSource := authorsSource
	oldFormat := authorsFormat
	authorsSource = dir
	authorsFormat = "text"
	defer func() { authorsSource = oldSource; authorsFormat = oldFormat }()

	err := runAuthorsExtract(nil, nil)
	require.NoError(t, err)
}

func TestRunMigrate_DryRunEmptySource(t *testing.T) {
	// Create temp cfg that points to empty CVS repo
	src := makeEmptyCVSRepo(t)
	tgt := t.TempDir()

	cfgPath := filepath.Join(t.TempDir(), "cfg.yaml")
	cfgContent := map[string]interface{}{
		"source": map[string]interface{}{
			"type": "cvs",
			"path": src,
		},
		"target": map[string]interface{}{
			"path": tgt,
		},
		"mapping": map[string]interface{}{},
		"options": map[string]interface{}{
			"dryRun": true,
		},
	}
	b, err := json.Marshal(cfgContent)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(cfgPath, b, 0644))

	oldCfg := migrateConfigFile
	oldDry := migrateDryRun
	migrateConfigFile = cfgPath
	migrateDryRun = true
	defer func() { migrateConfigFile = oldCfg; migrateDryRun = oldDry }()

	err = runMigrate(nil, nil)
	require.NoError(t, err)
}

func TestRunAnalyze_InvalidSourceType(t *testing.T) {
	oldType := analyzeSourceType
	oldSource := analyzeSource
	analyzeSourceType = "invalid"
	analyzeSource = "/tmp"
	defer func() { analyzeSourceType = oldType; analyzeSource = oldSource }()

	err := runAnalyze(nil, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported source type")
}

func TestRunAnalyze_SVNNotImplemented(t *testing.T) {
	oldType := analyzeSourceType
	oldSource := analyzeSource
	analyzeSourceType = "svn"
	analyzeSource = "/tmp"
	defer func() { analyzeSourceType = oldType; analyzeSource = oldSource }()

	err := runAnalyze(nil, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "SVN support is not yet implemented")
}

func TestRunAnalyze_ValidationFailure(t *testing.T) {
	oldType := analyzeSourceType
	oldSource := analyzeSource
	analyzeSourceType = "cvs"
	analyzeSource = "/nonexistent/path"
	defer func() { analyzeSourceType = oldType; analyzeSource = oldSource }()

	err := runAnalyze(nil, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "validation failed")
}

func TestRunAnalyze_WithRCSFiles(t *testing.T) {
	dir := makeEmptyCVSRepo(t)

	// Create an RCS file with commits
	rcsContent := `head	1.2;
access;
symbols
	RELEASE_1_0:1.2
	BETA_1:1.1;
locks; strict;
1.2
date	2023.12.01.00.00.00;	author user1;	state Exp;
branches;
next	1.1;
1.1
date	2023.01.01.00.00.00;	author user2;	state Exp;
branches;
next	;
desc
@@
1.2
log
@Second revision@
text
@updated content@
1.1
log
@Initial revision@
text
@initial content@
`
	rcsFile := filepath.Join(dir, "file.txt,v")
	require.NoError(t, os.WriteFile(rcsFile, []byte(rcsContent), 0644))

	oldType := analyzeSourceType
	oldSource := analyzeSource
	analyzeSourceType = "cvs"
	analyzeSource = dir
	defer func() { analyzeSourceType = oldType; analyzeSource = oldSource }()

	err := runAnalyze(nil, nil)
	require.NoError(t, err)
}

func TestRunAuthorsExtract_ValidationFailure(t *testing.T) {
	oldSource := authorsSource
	oldFormat := authorsFormat
	authorsSource = "/nonexistent/path"
	authorsFormat = "text"
	defer func() { authorsSource = oldSource; authorsFormat = oldFormat }()

	err := runAuthorsExtract(nil, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "validation failed")
}

func TestRunAuthorsExtract_YAMLFormat(t *testing.T) {
	dir := makeEmptyCVSRepo(t)

	oldSource := authorsSource
	oldFormat := authorsFormat
	authorsSource = dir
	authorsFormat = "yaml"
	defer func() { authorsSource = oldSource; authorsFormat = oldFormat }()

	err := runAuthorsExtract(nil, nil)
	require.NoError(t, err)
}

func TestRunAuthorsExtract_WithRCSFiles(t *testing.T) {
	dir := makeEmptyCVSRepo(t)

	// Create an RCS file with commits
	rcsContent := `head	1.2;
access;
symbols;
locks; strict;
1.2
date	2023.12.01.00.00.00;	author alice;	state Exp;
branches;
next	1.1;
1.1
date	2023.01.01.00.00.00;	author bob;	state Exp;
branches;
next	;
desc
@@
1.2
log
@Second revision@
text
@updated@
1.1
log
@Initial revision@
text
@initial@
`
	rcsFile := filepath.Join(dir, "file.txt,v")
	require.NoError(t, os.WriteFile(rcsFile, []byte(rcsContent), 0644))

	oldSource := authorsSource
	oldFormat := authorsFormat
	authorsSource = dir
	authorsFormat = "text"
	defer func() { authorsSource = oldSource; authorsFormat = oldFormat }()

	err := runAuthorsExtract(nil, nil)
	require.NoError(t, err)
}
