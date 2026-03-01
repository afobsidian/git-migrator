// Package git provides Git repository reading and writing capabilities for git-migrator.
package git

import (
	"fmt"

	"github.com/adamf123git/git-migrator/internal/vcs"
	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// Reader implements VCSReader for Git repositories
type Reader struct {
	path string
	repo *gogit.Repository
}

// NewReader creates a new Git repository reader
func NewReader(path string) *Reader {
	return &Reader{path: path}
}

// Validate checks if the Git repository is valid and accessible
func (r *Reader) Validate() error {
	repo, err := gogit.PlainOpen(r.path)
	if err != nil {
		return fmt.Errorf("failed to open git repository at %s: %w", r.path, err)
	}
	r.repo = repo
	return nil
}

// GetCommits returns an iterator over all commits (oldest first)
func (r *Reader) GetCommits() (vcs.CommitIterator, error) {
	if r.repo == nil {
		if err := r.Validate(); err != nil {
			return nil, err
		}
	}

	head, err := r.repo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD: %w", err)
	}

	commitIter, err := r.repo.Log(&gogit.LogOptions{From: head.Hash()})
	if err != nil {
		return nil, fmt.Errorf("failed to get commit log: %w", err)
	}
	defer commitIter.Close()

	// Collect commits (Log returns newest first; reverse for oldest first)
	var commits []*vcs.Commit
	err = commitIter.ForEach(func(c *object.Commit) error {
		commits = append(commits, &vcs.Commit{
			Revision: c.Hash.String(),
			Author:   c.Author.Name,
			Email:    c.Author.Email,
			Date:     c.Author.When,
			Message:  c.Message,
		})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to iterate commits: %w", err)
	}

	// Reverse to oldest-first order
	for i, j := 0, len(commits)-1; i < j; i, j = i+1, j-1 {
		commits[i], commits[j] = commits[j], commits[i]
	}

	return &gitCommitIterator{commits: commits}, nil
}

// GetCommitsSince returns an iterator over commits that come after the given
// revision hash (exclusive). If revision is empty, all commits are returned.
func (r *Reader) GetCommitsSince(revision string) (vcs.CommitIterator, error) {
	allIter, err := r.GetCommits()
	if err != nil {
		return nil, err
	}

	var all []*vcs.Commit
	for allIter.Next() {
		all = append(all, allIter.Commit())
	}
	if err := allIter.Err(); err != nil {
		return nil, err
	}

	if revision == "" {
		return &gitCommitIterator{commits: all}, nil
	}

	// Find the index of the given revision and return everything after it
	for i, c := range all {
		if c.Revision == revision {
			return &gitCommitIterator{commits: all[i+1:]}, nil
		}
	}

	// Revision not found – return all commits
	return &gitCommitIterator{commits: all}, nil
}

// GetBranches returns a list of branch names.
// If the repository has not been opened yet, Validate is called automatically.
func (r *Reader) GetBranches() ([]string, error) {
	if r.repo == nil {
		if err := r.Validate(); err != nil {
			return nil, err
		}
	}

	refs, err := r.repo.References()
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

// GetTags returns a map of tag names to commit hashes.
// If the repository has not been opened yet, Validate is called automatically.
func (r *Reader) GetTags() (map[string]string, error) {
	if r.repo == nil {
		if err := r.Validate(); err != nil {
			return nil, err
		}
	}

	refs, err := r.repo.References()
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

// GetHeadRevision returns the SHA of the current HEAD commit, or an empty
// string if the repository has no commits.
func (r *Reader) GetHeadRevision() (string, error) {
	if r.repo == nil {
		if err := r.Validate(); err != nil {
			return "", err
		}
	}

	head, err := r.repo.Head()
	if err != nil {
		return "", err
	}

	return head.Hash().String(), nil
}

// Close releases any resources held by the reader
func (r *Reader) Close() error {
	return nil
}

// gitCommitIterator iterates over a slice of vcs.Commit
type gitCommitIterator struct {
	commits []*vcs.Commit
	index   int
}

func (i *gitCommitIterator) Next() bool {
	i.index++
	return i.index <= len(i.commits)
}

func (i *gitCommitIterator) Commit() *vcs.Commit {
	if i.index < 1 || i.index > len(i.commits) {
		return nil
	}
	return i.commits[i.index-1]
}

func (i *gitCommitIterator) Err() error {
	return nil
}
