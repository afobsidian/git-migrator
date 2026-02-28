package cvs

import (
	"os/exec"
	"testing"
	"time"

	"github.com/adamf123git/git-migrator/internal/vcs"
)

func TestNewCVSWriter(t *testing.T) {
	w := NewWriter("/path/to/cvsroot", "mymodule")
	if w == nil {
		t.Fatal("NewWriter returned nil")
	}
	if w.repoPath != "/path/to/cvsroot" {
		t.Errorf("repoPath = %q, want %q", w.repoPath, "/path/to/cvsroot")
	}
	if w.module != "mymodule" {
		t.Errorf("module = %q, want %q", w.module, "mymodule")
	}
}

func TestCVSWriterClose(t *testing.T) {
	w := NewWriter("/tmp/cvsroot", "mod")
	if err := w.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

func TestCVSWriterInit_NoCVS(t *testing.T) {
	// Skip the test if cvs happens to be installed (we don't want a real checkout)
	if _, err := exec.LookPath("cvs"); err == nil {
		t.Skip("cvs binary present – skipping no-cvs test")
	}

	w := NewWriter("/tmp/cvsroot", "mod")
	err := w.Init(t.TempDir())
	if err == nil {
		t.Error("Init should fail when cvs is not available")
	}
}

func TestCVSWriterApplyCommit_NoWorkDir(t *testing.T) {
	w := NewWriter("/tmp/cvsroot", "mod")
	err := w.ApplyCommit(&vcs.Commit{
		Revision: "1.1",
		Author:   "test",
		Date:     time.Now(),
		Message:  "test commit",
	})
	if err == nil {
		t.Error("ApplyCommit should fail when working directory is not initialised")
	}
}

func TestCVSWriterCreateBranch_NoWorkDir(t *testing.T) {
	w := NewWriter("/tmp/cvsroot", "mod")
	if err := w.CreateBranch("mybranch", ""); err == nil {
		t.Error("CreateBranch should fail when working directory is not initialised")
	}
}

func TestCVSWriterCreateTag_NoWorkDir(t *testing.T) {
	w := NewWriter("/tmp/cvsroot", "mod")
	if err := w.CreateTag("v1.0", ""); err == nil {
		t.Error("CreateTag should fail when working directory is not initialised")
	}
}

func TestCVSWriterImplementsVCSWriter(t *testing.T) {
	var _ interface {
		Init(path string) error
		ApplyCommit(commit *vcs.Commit) error
		CreateBranch(name, revision string) error
		CreateTag(name, revision string) error
		Close() error
	} = (*Writer)(nil)
}
