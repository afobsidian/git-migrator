package cvs

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/adamf123git/git-migrator/internal/vcs"
	"github.com/stretchr/testify/require"
)

func TestValidator_NonExistentPath(t *testing.T) {
	v := NewValidator()
	res := v.Validate("/this-path-should-not-exist-12345")
	require.False(t, res.Valid)
	require.Greater(t, len(res.Errors), 0)
}

func TestValidate_WithCVSROOT(t *testing.T) {
	dir := t.TempDir()
	cvsroot := filepath.Join(dir, "CVSROOT")
	require.NoError(t, os.MkdirAll(cvsroot, 0755))

	// missing optional files should produce warnings but overall invalid? per implementation, missing optional files are warnings only
	v := NewValidator()
	res := v.Validate(dir)
	require.True(t, res.Valid)
}

func TestReader_NoRCSFiles_ReturnsEmptyCollections(t *testing.T) {
	dir := t.TempDir()
	cvsroot := filepath.Join(dir, "CVSROOT")
	require.NoError(t, os.MkdirAll(cvsroot, 0755))

	r := NewReader(dir)
	// Validate should succeed
	require.NoError(t, r.Validate())

	branches, err := r.GetBranches()
	require.NoError(t, err)
	require.Empty(t, branches)

	tags, err := r.GetTags()
	require.NoError(t, err)
	require.Empty(t, tags)

	it, err := r.GetCommits()
	require.NoError(t, err)
	// iterator should have no commits
	require.False(t, it.Next())
}

func TestReader_Close(t *testing.T) {
	dir := t.TempDir()
	cvsroot := filepath.Join(dir, "CVSROOT")
	require.NoError(t, os.MkdirAll(cvsroot, 0755))

	r := NewReader(dir)
	// Close should always return nil
	require.NoError(t, r.Close())
	// Close can be called multiple times
	require.NoError(t, r.Close())
}

func TestCommitIterator_Commit_And_Err(t *testing.T) {
	dir := t.TempDir()
	cvsroot := filepath.Join(dir, "CVSROOT")
	require.NoError(t, os.MkdirAll(cvsroot, 0755))

	r := NewReader(dir)
	it, err := r.GetCommits()
	require.NoError(t, err)

	// Test Commit() before calling Next() - should return nil
	require.Nil(t, it.Commit())

	// Test Err() - should always return nil for CVS iterator
	require.NoError(t, it.Err())

	// Test with no commits
	require.False(t, it.Next())
	require.Nil(t, it.Commit())
	require.NoError(t, it.Err())

	// Test calling Next() multiple times after iterator is exhausted
	require.False(t, it.Next())
	require.False(t, it.Next())
}

func TestCommitIterator_Bounds(t *testing.T) {
	dir := t.TempDir()
	cvsroot := filepath.Join(dir, "CVSROOT")
	require.NoError(t, os.MkdirAll(cvsroot, 0755))

	r := NewReader(dir)
	it, err := r.GetCommits()
	require.NoError(t, err)

	// Exhaust the iterator
	for it.Next() {
	}

	// Commit should return nil after exhaustion
	require.Nil(t, it.Commit())

	// Calling Next again should return false
	require.False(t, it.Next())
}

func TestSortCommitsByDate(t *testing.T) {
	// Test sorting commits by date
	now := time.Now()
	commits := []*vcs.Commit{
		{Revision: "3", Date: now.Add(2 * time.Hour)},
		{Revision: "1", Date: now},
		{Revision: "2", Date: now.Add(1 * time.Hour)},
	}

	sortCommitsByDate(commits)

	// Should be sorted oldest first
	require.Equal(t, "1", commits[0].Revision)
	require.Equal(t, "2", commits[1].Revision)
	require.Equal(t, "3", commits[2].Revision)
}

func TestSortCommitsByDate_Empty(t *testing.T) {
	// Test sorting empty slice
	commits := []*vcs.Commit{}
	sortCommitsByDate(commits)
	require.Empty(t, commits)
}

func TestSortCommitsByDate_Single(t *testing.T) {
	// Test sorting single commit
	commits := []*vcs.Commit{
		{Revision: "1", Date: time.Now()},
	}
	sortCommitsByDate(commits)
	require.Len(t, commits, 1)
}

func TestSortCommitsByDate_AlreadySorted(t *testing.T) {
	// Test already sorted commits
	now := time.Now()
	commits := []*vcs.Commit{
		{Revision: "1", Date: now},
		{Revision: "2", Date: now.Add(1 * time.Hour)},
		{Revision: "3", Date: now.Add(2 * time.Hour)},
	}

	sortCommitsByDate(commits)

	require.Equal(t, "1", commits[0].Revision)
	require.Equal(t, "2", commits[1].Revision)
	require.Equal(t, "3", commits[2].Revision)
}

func TestGetTags_WithRCSFiles(t *testing.T) {
	dir := t.TempDir()
	cvsroot := filepath.Join(dir, "CVSROOT")
	require.NoError(t, os.MkdirAll(cvsroot, 0755))

	// Create a simple RCS file with tags
	rcsContent := `head	1.2;
access;
symbols
	RELEASE_1_0:1.1
	RELEASE_1_1:1.2;
locks; strict;
1.2
date	2023.12.01.00.00.00;	author user;	state Exp;
branches;
next	1.1;
1.1
date	2023.01.01.00.00.00;	author user;	state Exp;
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
@content@
`
	rcsFile := filepath.Join(dir, "file.txt,v")
	require.NoError(t, os.WriteFile(rcsFile, []byte(rcsContent), 0644))

	r := NewReader(dir)
	tags, err := r.GetTags()
	require.NoError(t, err)
	require.NotEmpty(t, tags, "Should find tags in RCS file")
	require.Contains(t, tags, "RELEASE_1_0")
	require.Contains(t, tags, "RELEASE_1_1")
}

func TestValidate_NotDirectory(t *testing.T) {
	dir := t.TempDir()
	// Create a file instead of directory
	file := filepath.Join(dir, "notadir")
	require.NoError(t, os.WriteFile(file, []byte("test"), 0644))

	v := NewValidator()
	res := v.Validate(file)
	require.False(t, res.Valid)
	require.Greater(t, len(res.Errors), 0)
}
