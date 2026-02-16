package helpers

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateTempDirAndLoadFixture(t *testing.T) {
	dir := CreateTempDir(t)
	require.DirExists(t, dir)

	// Create a testdata dir and a file
	td := filepath.Join(dir, "testdata")
	require.NoError(t, os.MkdirAll(td, 0755))
	f := filepath.Join(td, "a.txt")
	require.NoError(t, os.WriteFile(f, []byte("x"), 0644))

	// Temporarily change working dir
	oldwd, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	defer os.Chdir(oldwd)

	data := LoadFixture(t, "a.txt")
	require.Equal(t, []byte("x"), data)
}
