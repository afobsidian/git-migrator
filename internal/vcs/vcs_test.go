package vcs

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCommitZeroValues(t *testing.T) {
	var c Commit
	require.Equal(t, "", c.Revision)
	require.Equal(t, "", c.Author)
	require.Equal(t, "", c.Email)
	require.True(t, c.Date.IsZero())
	require.Equal(t, "", c.Message)
	require.Nil(t, c.Files)
}

func TestFileChangeAndActionConstants(t *testing.T) {
	fc := FileChange{Path: "a.txt", Action: ActionAdd}
	require.Equal(t, "a.txt", fc.Path)
	require.Equal(t, ActionAdd, fc.Action)

	// Ensure iota ordering
	require.True(t, ActionModify < ActionAdd)
	require.True(t, ActionAdd < ActionDelete)
}

func TestRepositoryInfoZero(t *testing.T) {
	var r RepositoryInfo
	require.Equal(t, "", r.Path)
	require.Equal(t, "", r.VCS)
	require.Equal(t, "", r.Root)
	require.Nil(t, r.Branches)
	require.Nil(t, r.Tags)
}

func TestCommitWithDate(t *testing.T) {
	now := time.Now()
	c := Commit{Date: now}
	require.Equal(t, now, c.Date)
}
