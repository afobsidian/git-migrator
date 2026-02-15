// Package core provides migration orchestration for git-migrator.
package core

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/adamf123git/git-migrator/internal/mapping"
	"github.com/adamf123git/git-migrator/internal/progress"
	"github.com/adamf123git/git-migrator/internal/storage"
	"github.com/adamf123git/git-migrator/internal/vcs"
	"github.com/adamf123git/git-migrator/internal/vcs/cvs"
	"github.com/adamf123git/git-migrator/internal/vcs/git"
)

// MigrationConfig holds migration configuration
type MigrationConfig struct {
	SourceType  string            // cvs, svn
	SourcePath  string            // Path to source repo
	TargetPath  string            // Path to target Git repo
	AuthorMap   map[string]string // CVS user -> "Name <email>"
	BranchMap   map[string]string // CVS branch -> Git branch
	TagMap      map[string]string // CVS tag -> Git tag
	DryRun      bool              // Preview without changes
	Resume      bool              // Resume from last checkpoint
	StateFile   string            // Path to state file
	ChunkSize   int               // Save state every N commits
	InterruptAt int               // For testing: interrupt after N commits
}

// Migrator orchestrates the migration process
type Migrator struct {
	config    *MigrationConfig
	source    vcs.VCSReader
	target    *git.Writer
	authorMap *mapping.AuthorMap
	reporter  *progress.Reporter
	state     *MigrationState
	db        *storage.StateDB
}

// NewMigrator creates a new migrator
func NewMigrator(config *MigrationConfig) *Migrator {
	return &Migrator{
		config:    config,
		authorMap: mapping.NewAuthorMap(config.AuthorMap),
		reporter:  progress.NewReporter(0),
	}
}

// Run executes the migration
func (m *Migrator) Run() error {
	// Initialize source reader
	if err := m.initSource(); err != nil {
		return fmt.Errorf("failed to init source: %w", err)
	}

	// Validate source
	if err := m.source.Validate(); err != nil {
		return fmt.Errorf("source validation failed: %w", err)
	}

	// Initialize target
	if !m.config.DryRun {
		if err := m.initTarget(); err != nil {
			return fmt.Errorf("failed to init target: %w", err)
		}
		defer func() {
			if err := m.target.Close(); err != nil {
				// Log error but don't fail - cleanup is best effort
				log.Printf("Warning: failed to close target repository: %v", err)
			}
		}()
	}

	// Initialize state
	if err := m.initState(); err != nil {
		return fmt.Errorf("failed to init state: %w", err)
	}

	// Get commits from source
	iter, err := m.source.GetCommits()
	if err != nil {
		return fmt.Errorf("failed to get commits: %w", err)
	}

	// Collect commits
	var commits []*vcs.Commit
	for iter.Next() {
		commits = append(commits, iter.Commit())
	}
	if err := iter.Err(); err != nil {
		return fmt.Errorf("iterator error: %w", err)
	}

	m.reporter = progress.NewReporter(len(commits))
	m.reporter.Start()
	m.reporter.SetOperation("Starting migration")

	// Determine start position (for resume)
	startIdx := 0
	if m.config.Resume && m.state != nil {
		// Find the commit index to resume from
		for i, c := range commits {
			if c.Revision == m.state.lastCommit {
				startIdx = i + 1 // Resume from next commit
				break
			}
		}
		m.reporter.SetCurrent(m.state.processed)
	}

	// Process commits
	for i := startIdx; i < len(commits); i++ {
		commit := commits[i]

		m.reporter.SetOperation(fmt.Sprintf("Processing commit %s", commit.Revision[:8]))

		// Map author
		name, email := m.authorMap.Get(commit.Author)
		commit.Author = name
		commit.Email = email

		// Apply commit (if not dry run)
		if !m.config.DryRun {
			if err := m.target.ApplyCommit(commit); err != nil {
				return fmt.Errorf("failed to apply commit %s: %w", commit.Revision, err)
			}
		}

		m.reporter.Increment()

		// Save state periodically
		if m.config.ChunkSize > 0 && (i+1)%m.config.ChunkSize == 0 {
			if err := m.saveState(commit.Revision, i+1, len(commits)); err != nil {
				return fmt.Errorf("failed to save state: %w", err)
			}
		}

		// Test interruption
		if m.config.InterruptAt > 0 && i+1 >= m.config.InterruptAt {
			if err := m.saveState(commit.Revision, i+1, len(commits)); err != nil {
				// Log error but continue - this is test interruption
				log.Printf("Warning: failed to save state during test interruption: %v", err)
			}
			return fmt.Errorf("interrupted at commit %d", i+1)
		}
	}

	// Create branches
	if !m.config.DryRun {
		if err := m.createBranches(); err != nil {
			return fmt.Errorf("failed to create branches: %w", err)
		}
	}

	// Create tags
	if !m.config.DryRun {
		if err := m.createTags(); err != nil {
			return fmt.Errorf("failed to create tags: %w", err)
		}
	}

	// Mark complete
	if !m.config.DryRun {
		if err := m.markComplete(); err != nil {
			return fmt.Errorf("failed to mark complete: %w", err)
		}
	}

	m.reporter.SetOperation("Migration complete")

	return nil
}

func (m *Migrator) initSource() error {
	switch m.config.SourceType {
	case "cvs":
		m.source = cvs.NewReader(m.config.SourcePath)
	default:
		return fmt.Errorf("unsupported source type: %s", m.config.SourceType)
	}
	return nil
}

func (m *Migrator) initTarget() error {
	m.target = git.NewWriter()

	// Check if target exists
	if _, err := os.Stat(m.config.TargetPath); os.IsNotExist(err) {
		// Create new repo
		if err := m.target.Init(m.config.TargetPath); err != nil {
			return err
		}
	} else {
		// Open existing repo
		if err := m.target.Open(m.config.TargetPath); err != nil {
			return err
		}
	}

	return nil
}

func (m *Migrator) initState() error {
	if m.config.StateFile == "" {
		m.config.StateFile = filepath.Join(m.config.TargetPath, ".migration-state.db")
	}

	db, err := storage.NewStateDB(m.config.StateFile)
	if err != nil {
		return err
	}
	m.db = db

	migrationID := m.generateMigrationID()

	// Try to load existing state
	state, err := db.Load(migrationID)
	if err == nil && m.config.Resume {
		m.state = &MigrationState{
			migrationID: migrationID,
			lastCommit:  state.LastCommit,
			processed:   state.Processed,
			total:       state.Total,
		}
	} else {
		m.state = &MigrationState{
			migrationID: migrationID,
		}
	}

	return nil
}

func (m *Migrator) generateMigrationID() string {
	// Generate a unique ID based on source and target paths
	data := m.config.SourcePath + ":" + m.config.TargetPath
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:8])
}

func (m *Migrator) saveState(lastCommit string, processed, total int) error {
	m.state.lastCommit = lastCommit
	m.state.processed = processed
	m.state.total = total

	state := &storage.MigrationState{
		MigrationID: m.state.migrationID,
		LastCommit:  lastCommit,
		Processed:   processed,
		Total:       total,
		SourcePath:  m.config.SourcePath,
		TargetPath:  m.config.TargetPath,
		Status:      "in_progress",
	}

	return m.db.Save(state)
}

func (m *Migrator) createBranches() error {
	branches, err := m.source.GetBranches()
	if err != nil {
		return err
	}

	for _, branch := range branches {
		gitBranch := branch
		if mapped, ok := m.config.BranchMap[branch]; ok {
			gitBranch = mapped
		}

		m.reporter.SetOperation(fmt.Sprintf("Creating branch %s", gitBranch))
		if err := m.target.CreateBranch(gitBranch, "HEAD"); err != nil {
			// Log error but don't fail - branch creation is best effort
			log.Printf("Warning: failed to create branch %s: %v", gitBranch, err)
		}
	}

	return nil
}

func (m *Migrator) createTags() error {
	tags, err := m.source.GetTags()
	if err != nil {
		return err
	}

	for tagName, commitHash := range tags {
		gitTag := tagName
		if mapped, ok := m.config.TagMap[tagName]; ok {
			gitTag = mapped
		}

		m.reporter.SetOperation(fmt.Sprintf("Creating tag %s", gitTag))
		if err := m.target.CreateTag(gitTag, commitHash, ""); err != nil {
			// Log error but don't fail - tag creation is best effort
			log.Printf("Warning: failed to create tag %s: %v", gitTag, err)
		}
	}

	return nil
}

func (m *Migrator) markComplete() error {
	m.reporter.SetOperation("Finalizing migration")

	state := &storage.MigrationState{
		MigrationID: m.state.migrationID,
		LastCommit:  m.state.lastCommit,
		Processed:   m.state.processed,
		Total:       m.state.total,
		SourcePath:  m.config.SourcePath,
		TargetPath:  m.config.TargetPath,
		Status:      "completed",
	}

	if err := m.db.Save(state); err != nil {
		return err
	}

	return m.db.Complete(m.state.migrationID)
}

// ProgressReporter returns the progress reporter for subscribing to updates
func (m *Migrator) ProgressReporter() *progress.Reporter {
	return m.reporter
}

// MigrationState tracks the current migration state
type MigrationState struct {
	migrationID string
	lastCommit  string
	processed   int
	total       int
}

// NewMigrationState creates a simple migration state (for testing)
func NewMigrationState(path string) *SimpleState {
	return &SimpleState{path: path}
}

// SimpleState is a simple file-based state for testing
type SimpleState struct {
	path       string
	lastCommit string
	processed  int
	total      int
}

// Save saves the state to file
func (s *SimpleState) Save(commit string, processed, total int) error {
	s.lastCommit = commit
	s.processed = processed
	s.total = total

	// Write to file
	data := fmt.Sprintf("%s\n%d\n%d", commit, processed, total)
	return os.WriteFile(s.path, []byte(data), 0644)
}

// Load loads the state from file
func (s *SimpleState) Load() (string, int, int, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", 0, 0, nil
		}
		return "", 0, 0, err
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) < 3 {
		return "", 0, 0, fmt.Errorf("invalid state file")
	}

	var processed, total int
	if _, err := fmt.Sscanf(lines[1], "%d", &processed); err != nil {
		return "", 0, 0, fmt.Errorf("failed to parse processed count: %w", err)
	}
	if _, err := fmt.Sscanf(lines[2], "%d", &total); err != nil {
		return "", 0, 0, fmt.Errorf("failed to parse total count: %w", err)
	}

	s.lastCommit = lines[0]
	s.processed = processed
	s.total = total

	return lines[0], processed, total, nil
}

// Clear clears the state
func (s *SimpleState) Clear() error {
	s.lastCommit = ""
	s.processed = 0
	s.total = 0

	// Remove state file
	if err := os.Remove(s.path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// Close closes the state
func (s *SimpleState) Close() error {
	return nil
}
