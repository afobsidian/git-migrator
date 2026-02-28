// Package git provides Git repository writing capabilities for git-migrator.
package git

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/adamf123git/git-migrator/internal/vcs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
)

// Writer implements VCSWriter for Git repositories
type Writer struct {
	path       string
	repo       *git.Repository
	worktree   *git.Worktree
	lastCommit plumbing.Hash
}

// NewWriter creates a new Git repository writer
func NewWriter() *Writer {
	return &Writer{}
}

// Init creates a new repository at the given path
func (w *Writer) Init(path string) error {
	// Create directory if needed
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Initialize repository
	repo, err := git.PlainInit(path, false)
	if err != nil {
		return fmt.Errorf("failed to init repository: %w", err)
	}

	w.path = path
	w.repo = repo

	// Ensure .git/objects directory structure exists to prevent race conditions
	// when creating loose objects during concurrent operations
	objectsDir := filepath.Join(path, ".git", "objects")
	subdirs := []string{
		filepath.Join(objectsDir, "info"),
		filepath.Join(objectsDir, "pack"),
	}
	for _, dir := range subdirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create objects directory %s: %w", dir, err)
		}
	}

	// Get worktree
	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}
	w.worktree = worktree

	return nil
}

// InitWithConfig creates a repository with initial configuration
func (w *Writer) InitWithConfig(path string, cfg map[string]string) error {
	if err := w.Init(path); err != nil {
		return err
	}

	// Set configuration
	for key, value := range cfg {
		if err := w.SetConfig(key, value); err != nil {
			return err
		}
	}

	return nil
}

// SetConfig sets a configuration value
func (w *Writer) SetConfig(key, value string) error {
	if w.repo == nil {
		return fmt.Errorf("repository not initialized")
	}

	cfg, err := w.repo.Config()
	if err != nil {
		return err
	}

	// Parse key (e.g., "user.name" -> cfg.User.Name)
	switch key {
	case "user.name":
		cfg.User.Name = value
	case "user.email":
		cfg.User.Email = value
	default:
		// For other options, use sections
		cfg.Raw.AddOption("core", "", key, value)
	}

	return w.repo.SetConfig(cfg)
}

// GetConfig gets a configuration value
func (w *Writer) GetConfig(key string) (string, error) {
	if w.repo == nil {
		return "", fmt.Errorf("repository not initialized")
	}

	cfg, err := w.repo.Config()
	if err != nil {
		return "", err
	}

	switch key {
	case "user.name":
		return cfg.User.Name, nil
	case "user.email":
		return cfg.User.Email, nil
	default:
		// Try to find in raw config
		section := cfg.Raw.Section("core")
		if section != nil {
			for _, o := range section.Options {
				if o.Key == key {
					return o.Value, nil
				}
			}
		}
		return "", fmt.Errorf("config key not found: %s", key)
	}
}

// IsRepo checks if path is a Git repository
func (w *Writer) IsRepo(path string) bool {
	gitDir := filepath.Join(path, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		return true
	}

	// Try to open it
	_, err := git.PlainOpen(path)
	return err == nil
}

// ApplyCommit applies a commit to the repository
func (w *Writer) ApplyCommit(commit *vcs.Commit) error {
	if w.repo == nil || w.worktree == nil {
		return fmt.Errorf("repository not initialized")
	}

	// Process file changes
	for _, fc := range commit.Files {
		fullPath := filepath.Join(w.path, fc.Path)

		switch fc.Action {
		case vcs.ActionAdd, vcs.ActionModify:
			// Create directory if needed
			dir := filepath.Dir(fullPath)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}

			// Write file
			if err := os.WriteFile(fullPath, fc.Content, 0644); err != nil {
				return fmt.Errorf("failed to write file: %w", err)
			}

			// Add to staging
			_, err := w.worktree.Add(fc.Path)
			if err != nil {
				return fmt.Errorf("failed to add file: %w", err)
			}

		case vcs.ActionDelete:
			// Remove file
			if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("failed to remove file: %w", err)
			}

			// Remove from staging
			_, err := w.worktree.Remove(fc.Path)
			if err != nil {
				// Log if file wasn't tracked - this is expected for some deletions
				log.Printf("Debug: file %s not tracked in git, skipping removal: %v", fc.Path, err)
			}
		}
	}

	// Create commit
	hash, err := w.worktree.Commit(commit.Message, &git.CommitOptions{
		AllowEmptyCommits: true,
		Author: &object.Signature{
			Name:  commit.Author,
			Email: commit.Email,
			When:  commit.Date,
		},
		Committer: &object.Signature{
			Name:  commit.Author,
			Email: commit.Email,
			When:  commit.Date,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create commit: %w", err)
	}

	w.lastCommit = hash
	return nil
}

// CreateBranch creates a new branch
func (w *Writer) CreateBranch(name, revision string) error {
	if w.repo == nil {
		return fmt.Errorf("repository not initialized")
	}

	// Resolve revision to hash
	var hash plumbing.Hash
	if revision == "HEAD" {
		if w.lastCommit.IsZero() {
			// Get HEAD reference
			head, err := w.repo.Head()
			if err != nil {
				return fmt.Errorf("failed to get HEAD: %w", err)
			}
			hash = head.Hash()
		} else {
			hash = w.lastCommit
		}
	} else {
		// Parse revision
		h, err := w.repo.ResolveRevision(plumbing.Revision(revision))
		if err != nil {
			// Try as raw hash
			hash = plumbing.NewHash(revision)
			if hash.IsZero() {
				return fmt.Errorf("failed to resolve revision: %w", err)
			}
		} else {
			hash = *h
		}
	}

	// Create branch reference
	ref := plumbing.NewHashReference(plumbing.ReferenceName("refs/heads/"+name), hash)
	return w.repo.Storer.SetReference(ref)
}

// CreateTag creates a new tag
func (w *Writer) CreateTag(name, revision, message string) error {
	if w.repo == nil {
		return fmt.Errorf("repository not initialized")
	}

	// Resolve revision to hash
	var hash plumbing.Hash
	if revision == "HEAD" {
		if w.lastCommit.IsZero() {
			head, err := w.repo.Head()
			if err != nil {
				return fmt.Errorf("failed to get HEAD: %w", err)
			}
			hash = head.Hash()
		} else {
			hash = w.lastCommit
		}
	} else {
		h, err := w.repo.ResolveRevision(plumbing.Revision(revision))
		if err != nil {
			hash = plumbing.NewHash(revision)
			if hash.IsZero() {
				return fmt.Errorf("failed to resolve revision: %w", err)
			}
		} else {
			hash = *h
		}
	}

	if message == "" {
		// Lightweight tag
		ref := plumbing.NewHashReference(plumbing.ReferenceName("refs/tags/"+name), hash)
		return w.repo.Storer.SetReference(ref)
	}

	// Annotated tag - get commit for tagger info
	commit, err := w.repo.CommitObject(hash)
	if err != nil {
		return fmt.Errorf("failed to get commit: %w", err)
	}

	// Create tag object using object storage
	tag := &object.Tag{
		Name:       name,
		Tagger:     commit.Author,
		Message:    message,
		TargetType: plumbing.CommitObject,
		Target:     hash,
	}

	// Get object writer from storer
	objStorer, ok := w.repo.Storer.(storer.EncodedObjectStorer)
	if !ok {
		// Fallback to lightweight tag if we can't create annotated
		ref := plumbing.NewHashReference(plumbing.ReferenceName("refs/tags/"+name), hash)
		return w.repo.Storer.SetReference(ref)
	}

	// Create new encoded object using plumbing
	obj := new(plumbing.MemoryObject)
	if err := tag.Encode(obj); err != nil {
		return fmt.Errorf("failed to encode tag: %w", err)
	}

	tagHash, err := objStorer.SetEncodedObject(obj)
	if err != nil {
		return fmt.Errorf("failed to store tag object: %w", err)
	}

	// Create tag reference pointing to tag object
	ref := plumbing.NewHashReference(plumbing.ReferenceName("refs/tags/"+name), tagHash)
	return w.repo.Storer.SetReference(ref)
}

// ListBranches returns a list of branch names
func (w *Writer) ListBranches() ([]string, error) {
	if w.repo == nil {
		return nil, fmt.Errorf("repository not initialized")
	}

	refs, err := w.repo.References()
	if err != nil {
		return nil, err
	}

	var branches []string
	err = refs.ForEach(func(ref *plumbing.Reference) error {
		if ref.Name().IsBranch() {
			branches = append(branches, ref.Name().Short())
		}
		return nil
	})

	return branches, err
}

// ListTags returns a map of tag names to commit hashes
func (w *Writer) ListTags() (map[string]string, error) {
	if w.repo == nil {
		return nil, fmt.Errorf("repository not initialized")
	}

	refs, err := w.repo.References()
	if err != nil {
		return nil, err
	}

	tags := make(map[string]string)
	err = refs.ForEach(func(ref *plumbing.Reference) error {
		if ref.Name().IsTag() {
			tags[ref.Name().Short()] = ref.Hash().String()
		}
		return nil
	})

	return tags, err
}

// GetLastCommit returns the last commit info
func (w *Writer) GetLastCommit() (*vcs.Commit, error) {
	if w.repo == nil {
		return nil, fmt.Errorf("repository not initialized")
	}

	var hash plumbing.Hash
	if w.lastCommit.IsZero() {
		head, err := w.repo.Head()
		if err != nil {
			return nil, fmt.Errorf("failed to get HEAD: %w", err)
		}
		hash = head.Hash()
	} else {
		hash = w.lastCommit
	}

	commit, err := w.repo.CommitObject(hash)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit: %w", err)
	}

	return &vcs.Commit{
		Revision: commit.Hash.String(),
		Author:   commit.Author.Name,
		Email:    commit.Author.Email,
		Date:     commit.Author.When,
		Message:  commit.Message,
	}, nil
}

// GetCommitHashes returns all commit hashes in chronological order (oldest first)
func (w *Writer) GetCommitHashes() ([]string, error) {
	if w.repo == nil {
		return nil, fmt.Errorf("repository not initialized")
	}

	// Get all commits via log
	head, err := w.repo.Head()
	if err != nil {
		return nil, err
	}

	commitIter, err := w.repo.Log(&git.LogOptions{From: head.Hash()})
	if err != nil {
		return nil, err
	}
	defer commitIter.Close()

	var hashes []string
	err = commitIter.ForEach(func(c *object.Commit) error {
		hashes = append([]string{c.Hash.String()}, hashes...) // Prepend for oldest first
		return nil
	})

	return hashes, err
}

// Close releases any resources
func (w *Writer) Close() error {
	return nil
}

// Open opens an existing repository
func (w *Writer) Open(path string) error {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	w.path = path
	w.repo = repo

	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}
	w.worktree = worktree

	return nil
}

// ResolveRevision resolves a revision string to a hash
func (w *Writer) ResolveRevision(rev string) (string, error) {
	if w.repo == nil {
		return "", fmt.Errorf("repository not initialized")
	}

	// Simple hash check
	if len(rev) == 40 {
		return rev, nil
	}

	// Try as reference
	h, err := w.repo.ResolveRevision(plumbing.Revision(rev))
	if err != nil {
		return "", err
	}

	return h.String(), nil
}
