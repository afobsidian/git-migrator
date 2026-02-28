// Package cvs provides CVS repository reading and writing capabilities for git-migrator.
package cvs

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/adamf123git/git-migrator/internal/vcs"
)

// Writer implements VCSWriter for CVS repositories.
// It applies commits to a CVS repository by operating on a local working
// directory checkout and invoking the system `cvs` binary.
type Writer struct {
	repoPath string // Absolute path to the CVS repository (CVSROOT)
	module   string // CVS module name
	workDir  string // Working directory used for checkouts
}

// NewWriter creates a new CVS repository writer.
// repoPath is the CVSROOT path and module is the CVS module name.
func NewWriter(repoPath, module string) *Writer {
	return &Writer{
		repoPath: repoPath,
		module:   module,
	}
}

// Init checks out the CVS module into path, which becomes the working
// directory for subsequent operations.
func (w *Writer) Init(path string) error {
	if _, err := exec.LookPath("cvs"); err != nil {
		return fmt.Errorf("cvs command not found in PATH: %w", err)
	}

	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("failed to create work directory: %w", err)
	}

	// Check out the module into the work directory
	cmd := exec.Command("cvs", "-d", w.repoPath, "checkout", "-d", ".", w.module) //nolint:gosec
	cmd.Dir = path
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("cvs checkout failed: %w\n%s", err, out)
	}

	w.workDir = path
	return nil
}

// ApplyCommit applies the given commit's file changes to the CVS working
// directory and runs `cvs commit`.
func (w *Writer) ApplyCommit(commit *vcs.Commit) error {
	if w.workDir == "" {
		return fmt.Errorf("CVS working directory not initialised – call Init first")
	}

	var toAdd []string
	var toRemove []string

	for _, fc := range commit.Files {
		fullPath := filepath.Join(w.workDir, fc.Path)

		switch fc.Action {
		case vcs.ActionAdd, vcs.ActionModify:
			if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
				return fmt.Errorf("failed to create directory for %s: %w", fc.Path, err)
			}
			if err := os.WriteFile(fullPath, fc.Content, 0644); err != nil {
				return fmt.Errorf("failed to write file %s: %w", fc.Path, err)
			}
			if fc.Action == vcs.ActionAdd {
				toAdd = append(toAdd, fc.Path)
			}

		case vcs.ActionDelete:
			if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("failed to remove file %s: %w", fc.Path, err)
			}
			toRemove = append(toRemove, fc.Path)
		}
	}

	// Stage additions
	if len(toAdd) > 0 {
		args := append([]string{"-d", w.repoPath, "add"}, toAdd...)
		cmd := exec.Command("cvs", args...) //nolint:gosec
		cmd.Dir = w.workDir
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("cvs add failed: %w\n%s", err, out)
		}
	}

	// Stage removals
	if len(toRemove) > 0 {
		args := append([]string{"-d", w.repoPath, "remove"}, toRemove...)
		cmd := exec.Command("cvs", args...) //nolint:gosec
		cmd.Dir = w.workDir
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("cvs remove failed: %w\n%s", err, out)
		}
	}

	// Commit
	cmd := exec.Command("cvs", "-d", w.repoPath, "commit", "-m", commit.Message) //nolint:gosec
	cmd.Dir = w.workDir
	cmd.Env = append(os.Environ(), fmt.Sprintf("CVS_CLIENT_NAME=%s", commit.Author))
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("cvs commit failed: %w\n%s", err, out)
	}

	return nil
}

// CreateBranch creates a CVS branch tag in the working directory.
func (w *Writer) CreateBranch(name, _ string) error {
	if w.workDir == "" {
		return fmt.Errorf("CVS working directory not initialised – call Init first")
	}

	cmd := exec.Command("cvs", "-d", w.repoPath, "tag", "-b", name) //nolint:gosec
	cmd.Dir = w.workDir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("cvs tag -b %s failed: %w\n%s", name, err, out)
	}
	return nil
}

// CreateTag creates a CVS tag in the working directory.
func (w *Writer) CreateTag(name, _ string) error {
	if w.workDir == "" {
		return fmt.Errorf("CVS working directory not initialised – call Init first")
	}

	cmd := exec.Command("cvs", "-d", w.repoPath, "tag", name) //nolint:gosec
	cmd.Dir = w.workDir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("cvs tag %s failed: %w\n%s", name, err, out)
	}
	return nil
}

// Close releases any resources held by the writer.
func (w *Writer) Close() error {
	return nil
}
