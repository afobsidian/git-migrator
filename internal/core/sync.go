// Package core provides migration and sync orchestration for git-migrator.
package core

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/adamf123git/git-migrator/internal/mapping"
	"github.com/adamf123git/git-migrator/internal/progress"
	"github.com/adamf123git/git-migrator/internal/vcs"
	cvspkg "github.com/adamf123git/git-migrator/internal/vcs/cvs"
	gitpkg "github.com/adamf123git/git-migrator/internal/vcs/git"
)

// SyncDirection specifies which direction(s) to synchronise.
type SyncDirection string

const (
	// SyncGitToCVS propagates new Git commits into the CVS repository.
	SyncGitToCVS SyncDirection = "git-to-cvs"
	// SyncCVSToGit propagates new CVS commits into the Git repository.
	SyncCVSToGit SyncDirection = "cvs-to-git"
	// SyncBidirectional syncs in both directions (CVS→Git first, then Git→CVS).
	SyncBidirectional SyncDirection = "bidirectional"
)

// SyncConfig holds the configuration for a sync operation.
type SyncConfig struct {
	GitPath    string            // Absolute path to the Git repository
	CVSPath    string            // Absolute path to the CVS repository (CVSROOT)
	CVSModule  string            // CVS module name
	CVSWorkDir string            // Optional: working directory for CVS checkouts
	Direction  SyncDirection     // One of SyncGitToCVS, SyncCVSToGit, SyncBidirectional
	AuthorMap  map[string]string // CVS user → "Name <email>" (or Git name → CVS user)
	StateFile  string            // Path to the JSON state file (empty = no persistence)
	DryRun     bool              // When true, log planned changes without applying them
}

// SyncState records the most recent sync position for each direction.
type SyncState struct {
	LastGitCommit string    `json:"last_git_commit"` // Hash of the last Git commit synced to CVS
	LastCVSSync   time.Time `json:"last_cvs_sync"`   // Timestamp of the last CVS commit synced to Git
	SyncedAt      time.Time `json:"synced_at"`       // Wall-clock time of the last sync
}

// Syncer orchestrates bidirectional synchronisation between a Git repository
// and a CVS repository.
type Syncer struct {
	config    *SyncConfig
	authorMap *mapping.AuthorMap
	reporter  *progress.Reporter
	state     *SyncState
}

// NewSyncer creates a new Syncer from the supplied configuration.
func NewSyncer(config *SyncConfig) *Syncer {
	return &Syncer{
		config:    config,
		authorMap: mapping.NewAuthorMap(config.AuthorMap),
		reporter:  progress.NewReporter(0),
	}
}

// Run executes the configured sync operation.
func (s *Syncer) Run() error {
	if err := s.loadState(); err != nil {
		return fmt.Errorf("failed to load sync state: %w", err)
	}

	switch s.config.Direction {
	case SyncGitToCVS:
		return s.syncGitToCVS()
	case SyncCVSToGit:
		return s.syncCVSToGit()
	case SyncBidirectional:
		// Apply CVS changes first so that a Git→CVS pass won't re-sync them
		if err := s.syncCVSToGit(); err != nil {
			return fmt.Errorf("cvs-to-git sync failed: %w", err)
		}
		// Advance LastGitCommit to the current Git HEAD so that the
		// subsequent Git→CVS pass does not re-apply the commits that were
		// just imported from CVS (which would create an infinite sync loop).
		gitReader := gitpkg.NewReader(s.config.GitPath)
		if validateErr := gitReader.Validate(); validateErr == nil {
			if headCommit, headErr := gitReader.GetHeadRevision(); headErr == nil && headCommit != "" {
				s.state.LastGitCommit = headCommit
			} else if headErr != nil {
				log.Printf("Warning: could not read Git HEAD after cvs-to-git sync; bidirectional cycle prevention may not work: %v", headErr)
			}
			_ = gitReader.Close()
		} else {
			log.Printf("Warning: could not open Git repo after cvs-to-git sync; bidirectional cycle prevention may not work: %v", validateErr)
		}
		return s.syncGitToCVS()
	default:
		return fmt.Errorf("unknown sync direction: %q", s.config.Direction)
	}
}

// syncGitToCVS fetches commits from Git that are newer than the last sync
// and applies them to the CVS repository.
func (s *Syncer) syncGitToCVS() error {
	s.reporter.SetOperation("Syncing Git → CVS")

	gitReader := gitpkg.NewReader(s.config.GitPath)
	if err := gitReader.Validate(); err != nil {
		return fmt.Errorf("failed to open git repository: %w", err)
	}
	defer func() {
		if err := gitReader.Close(); err != nil {
			log.Printf("Warning: failed to close git reader: %v", err)
		}
	}()

	iter, err := gitReader.GetCommitsSince(s.state.LastGitCommit)
	if err != nil {
		return fmt.Errorf("failed to get git commits: %w", err)
	}

	var newCommits []*vcs.Commit
	for iter.Next() {
		newCommits = append(newCommits, iter.Commit())
	}
	if err := iter.Err(); err != nil {
		return fmt.Errorf("error iterating git commits: %w", err)
	}

	if len(newCommits) == 0 {
		s.reporter.SetOperation("Git → CVS: up to date")
		return nil
	}

	s.reporter.SetOperation(fmt.Sprintf("Git → CVS: %d new commit(s)", len(newCommits)))

	if s.config.DryRun {
		for _, c := range newCommits {
			rev := c.Revision
			if len(rev) > 8 {
				rev = rev[:8]
			}
			log.Printf("DRY RUN: would sync git commit %s (%s) to CVS", rev, c.Message)
		}
		return nil
	}

	// Prepare CVS working directory
	workDir, cleanup, err := s.prepareCVSWorkDir()
	if err != nil {
		return err
	}
	if cleanup != nil {
		defer cleanup()
	}

	cvsWriter := cvspkg.NewWriter(s.config.CVSPath, s.config.CVSModule)
	if err := cvsWriter.Init(workDir); err != nil {
		return fmt.Errorf("failed to initialise CVS writer: %w", err)
	}
	defer func() {
		if err := cvsWriter.Close(); err != nil {
			log.Printf("Warning: failed to close CVS writer: %v", err)
		}
	}()

	for _, commit := range newCommits {
		rev := commit.Revision
		if len(rev) > 8 {
			rev = rev[:8]
		}
		s.reporter.SetOperation(fmt.Sprintf("Applying git commit %s to CVS", rev))

		if err := cvsWriter.ApplyCommit(commit); err != nil {
			return fmt.Errorf("failed to apply git commit %s to CVS: %w", commit.Revision, err)
		}

		s.state.LastGitCommit = commit.Revision
		s.state.SyncedAt = time.Now()
		if err := s.saveState(); err != nil {
			log.Printf("Warning: failed to save sync state: %v", err)
		}
	}

	s.reporter.SetOperation(fmt.Sprintf("Git → CVS: synced %d commit(s)", len(newCommits)))
	return nil
}

// syncCVSToGit fetches CVS commits newer than the last sync timestamp and
// applies them to the Git repository.
func (s *Syncer) syncCVSToGit() error {
	s.reporter.SetOperation("Syncing CVS → Git")

	cvsReader := cvspkg.NewReader(s.config.CVSPath)
	if err := cvsReader.Validate(); err != nil {
		return fmt.Errorf("failed to open CVS repository: %w", err)
	}
	defer func() {
		if err := cvsReader.Close(); err != nil {
			log.Printf("Warning: failed to close CVS reader: %v", err)
		}
	}()

	iter, err := cvsReader.GetCommits()
	if err != nil {
		return fmt.Errorf("failed to get CVS commits: %w", err)
	}

	var newCommits []*vcs.Commit
	for iter.Next() {
		c := iter.Commit()
		if !s.state.LastCVSSync.IsZero() && !c.Date.After(s.state.LastCVSSync) {
			continue
		}
		newCommits = append(newCommits, c)
	}
	if err := iter.Err(); err != nil {
		return fmt.Errorf("error iterating CVS commits: %w", err)
	}

	if len(newCommits) == 0 {
		s.reporter.SetOperation("CVS → Git: up to date")
		return nil
	}

	s.reporter.SetOperation(fmt.Sprintf("CVS → Git: %d new commit(s)", len(newCommits)))

	if s.config.DryRun {
		for _, c := range newCommits {
			log.Printf("DRY RUN: would sync CVS commit %s (%s) to Git", c.Revision, c.Message)
		}
		return nil
	}

	gitWriter := gitpkg.NewWriter()
	if err := gitWriter.Open(s.config.GitPath); err != nil {
		return fmt.Errorf("failed to open git repository: %w", err)
	}
	defer func() {
		if err := gitWriter.Close(); err != nil {
			log.Printf("Warning: failed to close git writer: %v", err)
		}
	}()

	for _, commit := range newCommits {
		name, email := s.authorMap.Get(commit.Author)
		commit.Author = name
		commit.Email = email

		s.reporter.SetOperation(fmt.Sprintf("Applying CVS commit %s to Git", commit.Revision))

		if err := gitWriter.ApplyCommit(commit); err != nil {
			return fmt.Errorf("failed to apply CVS commit %s to Git: %w", commit.Revision, err)
		}

		s.state.LastCVSSync = commit.Date
		s.state.SyncedAt = time.Now()
		if err := s.saveState(); err != nil {
			log.Printf("Warning: failed to save sync state: %v", err)
		}
	}

	s.reporter.SetOperation(fmt.Sprintf("CVS → Git: synced %d commit(s)", len(newCommits)))
	return nil
}

// prepareCVSWorkDir returns the CVS working directory path and an optional
// cleanup function.  When CVSWorkDir is configured it is used directly;
// otherwise a temporary directory is created.
func (s *Syncer) prepareCVSWorkDir() (dir string, cleanup func(), err error) {
	if s.config.CVSWorkDir != "" {
		return s.config.CVSWorkDir, nil, nil
	}

	tmp, err := os.MkdirTemp("", "git-migrator-cvs-")
	if err != nil {
		return "", nil, fmt.Errorf("failed to create CVS work directory: %w", err)
	}
	return tmp, func() {
		if err := os.RemoveAll(tmp); err != nil {
			log.Printf("Warning: failed to clean up CVS work directory: %v", err)
		}
	}, nil
}

// loadState reads the sync state from disk.  Missing state file is not an
// error; it simply means the sync starts from scratch.
func (s *Syncer) loadState() error {
	s.state = &SyncState{}

	if s.config.StateFile == "" {
		return nil
	}

	data, err := os.ReadFile(s.config.StateFile)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to read state file: %w", err)
	}

	if err := json.Unmarshal(data, s.state); err != nil {
		return fmt.Errorf("failed to parse state file: %w", err)
	}
	return nil
}

// saveState writes the current sync state to disk.
func (s *Syncer) saveState() error {
	if s.config.StateFile == "" || s.config.DryRun {
		return nil
	}

	data, err := json.Marshal(s.state)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	if err := os.WriteFile(s.config.StateFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}
	return nil
}

// ProgressReporter returns the reporter for subscribing to sync progress.
func (s *Syncer) ProgressReporter() *progress.Reporter {
	return s.reporter
}
