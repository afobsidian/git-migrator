// Package cvs provides CVS repository reading and RCS file parsing capabilities.
package cvs

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/adamf123git/git-migrator/internal/vcs"
)

// ValidationMessage represents a validation message
type ValidationMessage struct {
	Field   string
	Message string
}

// ValidationResult represents the result of CVS repository validation
type ValidationResult struct {
	Valid    bool
	Errors   []ValidationMessage
	Warnings []ValidationMessage
	Infos    []ValidationMessage
}

// Reader implements VCSReader for CVS repositories
type Reader struct {
	path     string
	rcsFiles []*RCSFile
	// info caches repository metadata for performance optimization.
	// Reserved for future use to avoid repeated filesystem calls when
	// accessing repository information such as branch counts, file counts,
	// and other metadata that doesn't change during a single migration.
	//
	//nolint:unused // Reserved for future performance optimization
	info *vcs.RepositoryInfo
}

// NewReader creates a new CVS repository reader
func NewReader(path string) *Reader {
	return &Reader{path: path}
}

// Validate checks if the repository is valid and accessible
func (r *Reader) Validate() error {
	result := NewValidator().Validate(r.path)
	if !result.Valid {
		if len(result.Errors) > 0 {
			return fmt.Errorf("validation failed: %s", result.Errors[0].Message)
		}
		return fmt.Errorf("validation failed")
	}
	return nil
}

// GetCommits returns an iterator over all commits
func (r *Reader) GetCommits() (vcs.CommitIterator, error) {
	if err := r.loadRCSFiles(); err != nil {
		return nil, err
	}

	// Collect all commits from all RCS files
	var allCommits []*vcs.Commit
	seen := make(map[string]bool) // Track commits by revision+author+date

	for _, rcs := range r.rcsFiles {
		commits := rcs.GetCommits()
		for _, c := range commits {
			// Create a unique key for deduplication
			key := fmt.Sprintf("%s|%s|%d", c.Revision, c.Author, c.Date.Unix())
			if !seen[key] {
				seen[key] = true
				allCommits = append(allCommits, &vcs.Commit{
					Revision: c.Revision,
					Author:   c.Author,
					Date:     c.Date,
					Message:  c.Message,
					Branch:   c.Branch,
				})
			}
		}
	}

	// Sort commits by date (oldest first for proper application)
	sortCommitsByDate(allCommits)

	return &cvsCommitIterator{commits: allCommits}, nil
}

// GetBranches returns a list of branch names
func (r *Reader) GetBranches() ([]string, error) {
	if err := r.loadRCSFiles(); err != nil {
		return nil, err
	}

	branchSet := make(map[string]bool)
	for _, rcs := range r.rcsFiles {
		for _, branch := range rcs.GetBranches() {
			branchSet[branch] = true
		}
	}

	var branches []string
	for b := range branchSet {
		branches = append(branches, b)
	}
	return branches, nil
}

// GetTags returns a map of tag names to revision identifiers
func (r *Reader) GetTags() (map[string]string, error) {
	if err := r.loadRCSFiles(); err != nil {
		return nil, err
	}

	allTags := make(map[string]string)
	for _, rcs := range r.rcsFiles {
		for name, rev := range rcs.GetTags() {
			allTags[name] = rev
		}
	}
	return allTags, nil
}

// Close releases any resources
func (r *Reader) Close() error {
	return nil
}

// loadRCSFiles loads and parses all RCS files in the repository
func (r *Reader) loadRCSFiles() error {
	if r.rcsFiles != nil {
		return nil // Already loaded
	}

	// Find all ,v files (RCS files)
	err := filepath.Walk(r.path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if info.IsDir() {
			// Skip CVSROOT directory
			if filepath.Base(path) == "CVSROOT" {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if it's an RCS file (ends with ,v)
		if strings.HasSuffix(path, ",v") {
			file, err := os.Open(path)
			if err != nil {
				return nil // Skip files we can't read
			}
			defer func() {
				if err := file.Close(); err != nil {
					log.Printf("Warning: failed to close RCS file %s: %v", path, err)
				}
			}()

			parser := NewRCSParser(file)
			rcs, err := parser.Parse()
			if err != nil {
				return nil // Skip files we can't parse
			}

			r.rcsFiles = append(r.rcsFiles, rcs)
		}

		return nil
	})

	return err
}

// cvsCommitIterator implements CommitIterator for CVS
type cvsCommitIterator struct {
	commits []*vcs.Commit
	index   int
}

func (i *cvsCommitIterator) Next() bool {
	i.index++
	return i.index <= len(i.commits)
}

func (i *cvsCommitIterator) Commit() *vcs.Commit {
	if i.index < 1 || i.index > len(i.commits) {
		return nil
	}
	return i.commits[i.index-1]
}

func (i *cvsCommitIterator) Err() error {
	return nil
}

// sortCommitsByDate sorts commits chronologically (oldest first)
func sortCommitsByDate(commits []*vcs.Commit) {
	for i := 0; i < len(commits)-1; i++ {
		for j := i + 1; j < len(commits); j++ {
			if commits[i].Date.After(commits[j].Date) {
				commits[i], commits[j] = commits[j], commits[i]
			}
		}
	}
}

// Validator validates CVS repositories
type Validator struct{}

// NewValidator creates a new CVS repository validator
func NewValidator() *Validator {
	return &Validator{}
}

// Validate validates a CVS repository at the given path
func (v *Validator) Validate(path string) *ValidationResult {
	result := &ValidationResult{Valid: true}

	// Check path exists
	info, err := os.Stat(path)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationMessage{
			Field:   "path",
			Message: "Path does not exist: " + path,
		})
		return result
	}

	// Check is directory
	if !info.IsDir() {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationMessage{
			Field:   "path",
			Message: "Path is not a directory: " + path,
		})
		return result
	}

	// Check for CVSROOT
	cvsroot := filepath.Join(path, "CVSROOT")
	if _, err := os.Stat(cvsroot); os.IsNotExist(err) {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationMessage{
			Field:   "CVSROOT",
			Message: "CVSROOT directory not found",
		})
		return result
	}

	result.Infos = append(result.Infos, ValidationMessage{
		Field:   "repository",
		Message: "Repository structure is valid",
	})

	// Check for common CVSROOT files (warnings if missing)
	requiredFiles := []string{"history", "val-tags"}
	for _, file := range requiredFiles {
		filePath := filepath.Join(cvsroot, file)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			result.Warnings = append(result.Warnings, ValidationMessage{
				Field:   "CVSROOT/" + file,
				Message: "Optional file not found",
			})
		}
	}

	return result
}
