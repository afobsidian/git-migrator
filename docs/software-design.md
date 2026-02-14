# Git-Migrator - Software Design

**Status:** Internal Planning Document  
**Last Updated:** 2025-01-18  
**Version:** 1.0

---

## Implementation Overview

This document details the **implementation design** for Git-Migrator, including package structure, interfaces, data structures, algorithms, and workflows.

---

## Package Structure

```
github.com/adamf123git/git-migrator/
├── cmd/
│   └── git-migrator/
│       ├── main.go
│       └── commands/
│           ├── root.go
│           ├── migrate.go
│           ├── sync.go
│           ├── analyze.go
│           ├── validate.go
│           ├── authors.go
│           ├── web.go
│           └── version.go
│
├── internal/
│   ├── core/
│   │   ├── migration.go
│   │   ├── sync.go
│   │   ├── config.go
│   │   ├── state.go
│   │   └── errors.go
│   │
│   ├── vcs/
│   │   ├── vcs.go              # Interfaces
│   │   ├── cvs/
│   │   │   ├── reader.go
│   │   │   ├── rcs_parser.go
│   │   │   ├── rcs_lexer.go
│   │   │   ├── binary.go
│   │   │   ├── commit.go
│   │   │   └── branch.go
│   │   └── git/
│   │       ├── writer.go
│   │       ├── reader.go
│   │       └── verify.go
│   │
│   ├── mapping/
│   │   ├── authors.go
│   │   ├── branches.go
│   │   ├── tags.go
│   │   └── transformer.go
│   │
│   ├── progress/
│   │   ├── reporter.go
│   │   ├── terminal.go
│   │   ├── websocket.go
│   │   └── composite.go
│   │
│   ├── web/
│   │   ├── server.go
│   │   ├── handlers.go
│   │   ├── websocket.go
│   │   ├── api.go
│   │   └── static/
│   │       ├── index.html
│   │       ├── app.js
│   │       └── style.css
│   │
│   └── storage/
│       ├── sqlite.go
│       └── schema.sql
│
├── pkg/
│   ├── types/
│   │   ├── commit.go
│   │   ├── author.go
│   │   ├── branch.go
│   │   ├── tag.go
│   │   ├── repository.go
│   │   └── config.go
│   │
│   └── utils/
│       ├── file.go
│       ├── encoding.go
│       └── hash.go
│
├── test/
│   ├── fixtures/
│   ├── helpers/
│   ├── requirements/
│   └── regression/
│
└── scripts/
    ├── build.sh
    ├── test.sh
    └── check-requirements.go
```

---

## Core Types & Data Structures

### pkg/types/commit.go

```go
package types

import "time"

type Commit struct {
    // Unique identifier (VCS-specific)
    ID string `json:"id"`
    
    // Author information
    Author     Author `json:"author"`
    Committer  Author `json:"committer"`
    
    // Commit metadata
    Message    string    `json:"message"`
    Date       time.Time `json:"date"`
    
    // File changes
    Files      []FileChange `json:"files"`
    
    // Parent commits (for branches/merges)
    Parents    []string `json:"parents,omitempty"`
    
    // Branch this commit belongs to
    Branch     string `json:"branch"`
    
    // Additional metadata
    Metadata   map[string]string `json:"metadata,omitempty"`
}

type FileChange struct {
    Path     string      `json:"path"`
    Action   FileAction  `json:"action"`
    Content  []byte      `json:"content,omitempty"`
    Mode     int         `json:"mode,omitempty"` // Unix file mode
}

type FileAction int

const (
    FileAdd FileAction = iota
    FileModify
    FileDelete
)

func (a FileAction) String() string {
    return [...]string{"add", "modify", "delete"}[a]
}
```

### pkg/types/author.go

```go
package types

type Author struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

func (a Author) String() string {
    if a.Email != "" {
        return a.Name + " <" + a.Email + ">"
    }
    return a.Name
}

// ParseAuthor parses "Name <email>" format
func ParseAuthor(s string) Author {
    // Implementation
}
```

### pkg/types/repository.go

```go
package types

type RepositoryInfo struct {
    Type         string    `json:"type"` // cvs, git, svn
    Path         string    `json:"path"`
    RootPath     string    `json:"rootPath"`
    Module       string    `json:"module,omitempty"` // CVS module
    
    // Statistics
    CommitCount  int       `json:"commitCount"`
    BranchCount  int       `json:"branchCount"`
    TagCount     int       `json:"tagCount"`
    
    // Metadata
    LastCommit   time.Time `json:"lastCommit"`
    FirstCommit  time.Time `json:"firstCommit"`
}
```

### pkg/types/config.go

```go
package types

type Config struct {
    Source   SourceConfig   `yaml:"source" json:"source"`
    Target   TargetConfig   `yaml:"target" json:"target"`
    Mapping  MappingConfig  `yaml:"mapping" json:"mapping"`
    Options  OptionsConfig  `yaml:"options" json:"options"`
}

type SourceConfig struct {
    Type   string `yaml:"type" json:"type"`     // cvs, svn, etc.
    Path   string `yaml:"path" json:"path"`
    Module string `yaml:"module" json:"module"` // CVS module
    
    // CVS-specific
    CVSMode string `yaml:"cvsMode" json:"cvsMode"` // rcs, binary, auto
}

type TargetConfig struct {
    Type   string `yaml:"type" json:"type"`     // git
    Path   string `yaml:"path" json:"path"`
    Remote string `yaml:"remote" json:"remote"` // Optional: push to remote
}

type MappingConfig struct {
    Authors  map[string]string `yaml:"authors" json:"authors"`
    Branches map[string]string `yaml:"branches" json:"branches"`
    Tags     map[string]string `yaml:"tags" json:"tags"`
}

type OptionsConfig struct {
    DryRun              bool `yaml:"dryRun" json:"dryRun"`
    PreserveEmptyCommits bool `yaml:"preserveEmptyCommits" json:"preserveEmptyCommits"`
    ChunkSize           int  `yaml:"chunkSize" json:"chunkSize"`
    Verbose             bool `yaml:"verbose" json:"verbose"`
    
    // Resume options
    Resume              bool `yaml:"resume" json:"resume"`
    ResumeFromCommit    string `yaml:"resumeFromCommit" json:"resumeFromCommit"`
}
```

---

## VCS Interface Design

### internal/vcs/vcs.go

```go
package vcs

import (
    "context"
    "github.com/adamf123git/git-migrator/pkg/types"
)

// VCSReader reads from source version control system
type VCSReader interface {
    // Repository operations
    Validate(path string) error
    GetInfo() (*types.RepositoryInfo, error)
    
    // History access
    ListBranches() ([]types.Branch, error)
    ListTags() ([]types.Tag, error)
    GetCommitIterator(opts CommitIteratorOptions) (CommitIterator, error)
    
    // Content access
    GetFile(commitID, path string) ([]byte, error)
    ListFiles(commitID string) ([]string, error)
    
    // Cleanup
    Close() error
}

// CommitIterator streams commits for memory efficiency
type CommitIterator interface {
    Next() (*types.Commit, error)
    HasNext() bool
    Close() error
}

type CommitIteratorOptions struct {
    // Filter by branch (optional)
    Branch string
    
    // Time range (optional)
    Since time.Time
    Until time.Time
    
    // Resume from specific commit
    AfterCommit string
    
    // Ordering
    Order OrderType // Chronological, Reverse, Topological
}

type OrderType int

const (
    OrderChronological OrderType = iota
    OrderReverse
    OrderTopological
)

// VCSWriter writes to target version control system
type VCSWriter interface {
    // Repository operations
    Init(path string) error
    Validate(path string) error
    
    // Commit operations
    CreateBranch(name string, commitID string) error
    CreateTag(name string, commitID string, message string) error
    ApplyCommit(commit *types.Commit) (string, error)
    
    // File operations
    SetFile(path string, content []byte, mode int) error
    DeleteFile(path string) error
    
    // Finalization
    Finalize() error
    Close() error
}

// VCSSyncer bidirectional sync (future)
type VCSSyncer interface {
    GetChanges(since time.Time) ([]Change, error)
    ApplyChange(change Change) error
    ResolveConflict(conflict Conflict) (Resolution, error)
}
```

---

## CVS Implementation

### internal/vcs/cvs/reader.go

```go
package cvs

import (
    "context"
    "github.com/adamf123git/git-migrator/internal/vcs"
    "github.com/adamf123git/git-migrator/pkg/types"
)

type CVSReader struct {
    path     string
    module   string
    mode     CVSMode
    parser   *RCSParser
    client   *CVSClient
}

type CVSMode int

const (
    ModeRCS   CVSMode = iota // Parse RCS files directly
    ModeBinary                // Use cvs command
    ModeAuto                  // Try RCS first, fallback to binary
)

func NewCVSReader(config types.SourceConfig) (*CVSReader, error) {
    reader := &CVSReader{
        path:   config.Path,
        module: config.Module,
    }
    
    // Determine mode
    switch config.CVSMode {
    case "rcs":
        reader.mode = ModeRCS
    case "binary":
        reader.mode = ModeBinary
    default:
        reader.mode = ModeAuto
    }
    
    // Initialize parser or client based on mode
    if reader.mode == ModeRCS || reader.mode == ModeAuto {
        reader.parser = NewRCSParser(config.Path)
    }
    if reader.mode == ModeBinary || reader.mode == ModeAuto {
        reader.client = NewCVSClient(config.Path)
    }
    
    return reader, nil
}

func (r *CVSReader) GetCommitIterator(opts vcs.CommitIteratorOptions) (vcs.CommitIterator, error) {
    if r.mode == ModeRCS || r.mode == ModeAuto {
        // Try RCS parsing first
        iter, err := r.parser.GetCommitIterator(opts)
        if err == nil {
            return iter, nil
        }
        
        // Fallback to binary if Auto mode
        if r.mode == ModeAuto {
            return r.client.GetCommitIterator(opts)
        }
        return nil, err
    }
    
    return r.client.GetCommitIterator(opts)
}
```

### internal/vcs/cvs/rcs_parser.go

```go
package cvs

import (
    "bufio"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "regexp"
    "strings"
    "time"
    
    "github.com/adamf123git/git-migrator/internal/vcs"
    "github.com/adamf123git/git-migrator/pkg/types"
)

// RCSParser parses CVS RCS files directly (,v files)
type RCSParser struct {
    rootPath string
    files    []string // All ,v files
}

// RCS file format constants
const (
    RCSHeader      = "head"
    RCSBranch      = "branch"
    RCSAccess      = "access"
    RCSymbols      = "symbols"
    RCSLocks       = "locks"
    RCSStrict      = "strict"
    RCSExpand      = "expand"
    RCSDesc        = "desc"
    RCSRevision    = "revision"
    RCSDate        = "date"
    RCSAuthor      = "author"
    RCSState       = "state"
    RCSBranches    = "branches"
    RCSNext        = "next"
    RCSLog         = "log"
    RCSVText       = "text"
)

type RCSFile struct {
    Head      string
    Branch    string
    Symbols   map[string]string // tag -> revision
    Locks     map[string]string
    Revisions []RCSRevision
    Desc      string
}

type RCSRevision struct {
    Number   string
    Date     time.Time
    Author   string
    State    string
    Branches []string
    Next     string
    Log      string
    Text     string // Delta text
}

func NewRCSParser(rootPath string) *RCSParser {
    return &RCSParser{
        rootPath: rootPath,
    }
}

func (p *RCSParser) ParseFile(path string) (*RCSFile, error) {
    file, err := os.Open(path)
    if err != nil {
        return nil, fmt.Errorf("opening RCS file: %w", err)
    }
    defer file.Close()
    
    rcs := &RCSFile{
        Symbols: make(map[string]string),
        Locks:   make(map[string]string),
    }
    
    scanner := bufio.NewScanner(file)
    lexer := NewRCSLexer(scanner)
    
    if err := p.parseHeader(lexer, rcs); err != nil {
        return nil, err
    }
    
    if err := p.parseRevisions(lexer, rcs); err != nil {
        return nil, err
    }
    
    return rcs, nil
}

func (p *RCSParser) parseHeader(lexer *RCSLexer, rcs *RCSFile) error {
    for {
        token := lexer.NextToken()
        if token == nil {
            return io.EOF
        }
        
        switch token.Value {
        case RCSHeader:
            rcs.Head = lexer.NextToken().Value
        case RCSBranch:
            rcs.Branch = lexer.NextToken().Value
        case RCSSymbols:
            // Parse symbols: tag1:1.1; tag2:1.2;
            for {
                next := lexer.PeekToken()
                if next == nil || isKeyword(next.Value) {
                    break
                }
                tag := lexer.NextToken().Value
                lexer.NextToken() // Skip ":"
                rev := lexer.NextToken().Value
                rcs.Symbols[tag] = rev
                
                // Skip semicolon
                if lexer.PeekToken().Value == ";" {
                    lexer.NextToken()
                }
            }
        case RCSDesc:
            rcs.Desc = p.parseString(lexer)
            return nil
        }
    }
}

func (p *RCSParser) parseRevisions(lexer *RCSLexer, rcs *RCSFile) error {
    for {
        token := lexer.NextToken()
        if token == nil {
            return nil
        }
        
        if token.Value == RCSRevision {
            rev := RCSRevision{
                Number: lexer.NextToken().Value,
            }
            
            // Parse revision metadata
            for {
                next := lexer.NextToken()
                if next == nil || next.Value == RCSRevision {
                    lexer.PushToken(next)
                    break
                }
                
                switch next.Value {
                case RCSDate:
                    dateStr := lexer.NextToken().Value
                    rev.Date = parseRCSDate(dateStr)
                case RCSAuthor:
                    rev.Author = lexer.NextToken().Value
                case RCSState:
                    rev.State = lexer.NextToken().Value
                case RCSBranches:
                    // Parse branch numbers
                    for {
                        b := lexer.NextToken()
                        if b.Value == ";" {
                            break
                        }
                        if b.Value != "" && b.Value != RCSNext {
                            rev.Branches = append(rev.Branches, b.Value)
                        }
                    }
                case RCSNext:
                    rev.Next = lexer.NextToken().Value
                case RCSLog:
                    rev.Log = p.parseString(lexer)
                case RCSVText:
                    rev.Text = p.parseString(lexer)
                    rcs.Revisions = append(rcs.Revisions, rev)
                }
            }
        }
    }
}

func (p *RCSParser) parseString(lexer *RCSLexer) string {
    // Strings in RCS are enclosed in @
    token := lexer.NextToken()
    if token.Value != "@" {
        return token.Value
    }
    
    var sb strings.Builder
    for {
        token = lexer.NextToken()
        if token.Value == "@" {
            // Check for escaped @@ 
            next := lexer.PeekToken()
            if next != nil && next.Value == "@" {
                sb.WriteByte('@')
                lexer.NextToken()
            } else {
                break
            }
        } else {
            sb.WriteString(token.Value)
        }
    }
    
    return sb.String()
}

func parseRCSDate(s string) time.Time {
    // RCS date format: YYYY.MM.DD.HH.MM.SS
    parts := strings.Split(s, ".")
    if len(parts) != 6 {
        return time.Time{}
    }
    
    year, _ := strconv.Atoi(parts[0])
    month, _ := strconv.Atoi(parts[1])
    day, _ := strconv.Atoi(parts[2])
    hour, _ := strconv.Atoi(parts[3])
    min, _ := strconv.Atoi(parts[4])
    sec, _ := strconv.Atoi(parts[5])
    
    return time.Date(year, time.Month(month), day, hour, min, sec, 0, time.UTC)
}

func (p *RCSParser) GetCommitIterator(opts vcs.CommitIteratorOptions) (vcs.CommitIterator, error) {
    // Discover all ,v files
    if p.files == nil {
        err := filepath.Walk(p.rootPath, func(path string, info os.FileInfo, err error) error {
            if err != nil {
                return err
            }
            if strings.HasSuffix(path, ",v") {
                p.files = append(p.files, path)
            }
            return nil
        })
        if err != nil {
            return nil, fmt.Errorf("walking CVS root: %w", err)
        }
    }
    
    return &RCSCommitIterator{
        parser:  p,
        files:   p.files,
        options: opts,
    }, nil
}
```

### internal/vcs/cvs/rcs_lexer.go

```go
package cvs

import (
    "bufio"
    "strings"
)

type TokenType int

const (
    TokenKeyword TokenType = iota
    TokenString
    TokenNumber
    TokenSymbol
    TokenSemicolon
    TokenColon
    TokenAt
    TokenEOF
)

type Token struct {
    Type  TokenType
    Value string
    Line  int
}

type RCSLexer struct {
    scanner   *bufio.Scanner
    tokens    []Token
    pos       int
    pushback  *Token
}

func NewRCSLexer(scanner *bufio.Scanner) *RCSLexer {
    return &RCSLexer{
        scanner: scanner,
    }
}

func (l *RCSLexer) NextToken() *Token {
    if l.pushback != nil {
        token := l.pushback
        l.pushback = nil
        return token
    }
    
    // Read next token
    for l.scanner.Scan() {
        text := strings.TrimSpace(l.scanner.Text())
        if text == "" {
            continue
        }
        
        return l.tokenize(text)
    }
    
    return &Token{Type: TokenEOF}
}

func (l *RCSLexer) PeekToken() *Token {
    token := l.NextToken()
    l.pushback = token
    return token
}

func (l *RCSLexer) PushToken(token *Token) {
    l.pushback = token
}

func (l *RCSLexer) tokenize(text string) *Token {
    if text == "@" {
        return &Token{Type: TokenAt, Value: text}
    }
    if text == ";" {
        return &Token{Type: TokenSemicolon, Value: text}
    }
    if text == ":" {
        return &Token{Type: TokenColon, Value: text}
    }
    if isKeyword(text) {
        return &Token{Type: TokenKeyword, Value: text}
    }
    
    // Default to symbol/string
    return &Token{Type: TokenSymbol, Value: text}
}

func isKeyword(s string) bool {
    keywords := []string{
        "head", "branch", "access", "symbols", "locks",
        "strict", "expand", "desc", "revision", "date",
        "author", "state", "branches", "next", "log", "text",
    }
    for _, k := range keywords {
        if s == k {
            return true
        }
    }
    return false
}
```

---

## Git Implementation

### internal/vcs/git/writer.go

```go
package git

import (
    "fmt"
    "os"
    "path/filepath"
    
    "github.com/go-git/go-git/v5"
    "github.com/go-git/go-git/v5/plumbing"
    "github.com/go-git/go-git/v5/plumbing/object"
    "github.com/go-git/go-git/v5/plumbing/filemode"
    
    "github.com/adamf123git/git-migrator/internal/vcs"
    "github.com/adamf123git/git-migrator/pkg/types"
)

type GitWriter struct {
    repo   *git.Repository
    path   string
    commits map[string]plumbing.Hash // source ID -> git hash
}

func NewGitWriter() *GitWriter {
    return &GitWriter{
        commits: make(map[string]plumbing.Hash),
    }
}

func (w *GitWriter) Init(path string) error {
    w.path = path
    
    // Create directory if it doesn't exist
    if err := os.MkdirAll(path, 0755); err != nil {
        return fmt.Errorf("creating directory: %w", err)
    }
    
    // Initialize git repository
    repo, err := git.PlainInit(path, false)
    if err != nil {
        return fmt.Errorf("initializing git repo: %w", err)
    }
    
    w.repo = repo
    return nil
}

func (w *GitWriter) ApplyCommit(commit *types.Commit) (string, error) {
    // Create tree with file changes
    tree := &object.Tree{}
    
    for _, fileChange := range commit.Files {
        switch fileChange.Action {
        case types.FileAdd, types.FileModify:
            entry := object.TreeEntry{
                Name: filepath.Base(fileChange.Path),
                Mode: filemode.FileMode(fileChange.Mode),
                Hash: plumbing.NewHash(""), // Will be set when we hash the content
            }
            tree.Entries = append(tree.Entries, entry)
            
        case types.FileDelete:
            // Mark for deletion
        }
    }
    
    // Create commit
    gitCommit := &object.Commit{
        Author: object.Signature{
            Name:  commit.Author.Name,
            Email: commit.Author.Email,
            When:  commit.Date,
        },
        Committer: object.Signature{
            Name:  commit.Committer.Name,
            Email: commit.Committer.Email,
            When:  commit.Date,
        },
        Message:  commit.Message,
    }
    
    // Set parent commits
    if len(commit.Parents) > 0 {
        for _, parentID := range commit.Parents {
            if parentHash, ok := w.commits[parentID]; ok {
                gitCommit.ParentHashes = append(gitCommit.ParentHashes, parentHash)
            }
        }
    }
    
    // Write commit to repository
    obj := w.repo.Storer.NewEncodedObject()
    if err := gitCommit.Encode(obj); err != nil {
        return "", fmt.Errorf("encoding commit: %w", err)
    }
    
    hash, err := w.repo.Storer.SetEncodedObject(obj)
    if err != nil {
        return "", fmt.Errorf("storing commit: %w", err)
    }
    
    // Map source commit ID to Git hash
    w.commits[commit.ID] = hash
    
    return hash.String(), nil
}

func (w *GitWriter) CreateBranch(name string, commitID string) error {
    hash, ok := w.commits[commitID]
    if !ok {
        return fmt.Errorf("commit %s not found", commitID)
    }
    
    ref := plumbing.NewBranchReferenceName(name)
    return w.repo.Storer.SetReference(plumbing.NewHashReference(ref, hash))
}

func (w *GitWriter) CreateTag(name string, commitID string, message string) error {
    hash, ok := w.commits[commitID]
    if !ok {
        return fmt.Errorf("commit %s not found", commitID)
    }
    
    ref := plumbing.NewTagReferenceName(name)
    return w.repo.Storer.SetReference(plumbing.NewHashReference(ref, hash))
}

func (w *GitWriter) Finalize() error {
    // Ensure HEAD points to main branch
    ref := plumbing.NewBranchReferenceName("main")
    return w.repo.Storer.SetReference(plumbing.NewSymbolicReference(plumbing.HEAD, ref))
}

func (w *GitWriter) Close() error {
    // Cleanup resources
    return nil
}
```

---

## Migration Orchestrator

### internal/core/migration.go

```go
package core

import (
    "context"
    "fmt"
    "sync"
    
    "github.com/adamf123git/git-migrator/internal/vcs"
    "github.com/adamf123git/git-migrator/pkg/types"
)

type MigrationOrchestrator struct {
    source     vcs.VCSReader
    target     vcs.VCSWriter
    mapper     *Mapper
    progress   ProgressReporter
    stateStore StateStore
    config     *types.Config
    
    mu         sync.Mutex
    cancelled  bool
}

func NewMigrationOrchestrator(
    source vcs.VCSReader,
    target vcs.VCSWriter,
    config *types.Config,
) *MigrationOrchestrator {
    return &MigrationOrchestrator{
        source:     source,
        target:     target,
        mapper:     NewMapper(config.Mapping),
        progress:   NewCompositeProgress(),
        stateStore: NewSQLiteStateStore(),
        config:     config,
    }
}

func (o *MigrationOrchestrator) Migrate(ctx context.Context) error {
    // 1. Validate source repository
    o.progress.Info("Validating source repository")
    if err := o.source.Validate(o.config.Source.Path); err != nil {
        return fmt.Errorf("validating source: %w", err)
    }
    
    // 2. Initialize target repository
    o.progress.Info("Initializing target repository")
    if err := o.target.Init(o.config.Target.Path); err != nil {
        return fmt.Errorf("initializing target: %w", err)
    }
    
    // 3. Load or create migration state
    state, err := o.stateStore.Load()
    if err != nil {
        state = NewMigrationState(o.config)
        if err := o.stateStore.Save(state); err != nil {
            return fmt.Errorf("saving initial state: %w", err)
        }
    }
    
    // 4. Get commit iterator
    opts := vcs.CommitIteratorOptions{
        AfterCommit: state.LastCommitID,
        Order:       vcs.OrderChronological,
    }
    
    iterator, err := o.source.GetCommitIterator(opts)
    if err != nil {
        return fmt.Errorf("getting commits: %w", err)
    }
    defer iterator.Close()
    
    // 5. Process commits
    o.progress.Info("Starting migration")
    commitCount := 0
    
    for iterator.HasNext() {
        // Check for cancellation
        select {
        case <-ctx.Done():
            o.progress.Warn("Migration cancelled")
            return ctx.Err()
        default:
        }
        
        // Get next commit
        commit, err := iterator.Next()
        if err != nil {
            return fmt.Errorf("reading commit: %w", err)
        }
        
        // Apply mappings
        o.applyMappings(commit)
        
        // Skip empty commits if configured
        if len(commit.Files) == 0 && !o.config.Options.PreserveEmptyCommits {
            continue
        }
        
        // Dry run mode
        if o.config.Options.DryRun {
            o.progress.Info(fmt.Sprintf("[DRY RUN] Would apply commit: %s", commit.ID))
            continue
        }
        
        // Apply commit to target
        gitHash, err := o.target.ApplyCommit(commit)
        if err != nil {
            o.stateStore.MarkFailed(commit.ID, err)
            return fmt.Errorf("applying commit %s: %w", commit.ID, err)
        }
        
        // Update state
        state.LastCommitID = commit.ID
        state.ProcessedCommits++
        commitCount++
        
        if commitCount%o.config.Options.ChunkSize == 0 {
            if err := o.stateStore.Save(state); err != nil {
                return fmt.Errorf("saving state: %w", err)
            }
        }
        
        // Report progress
        o.progress.Update(commitCount, fmt.Sprintf("Applied commit %s -> %s", commit.ID, gitHash))
    }
    
    // 6. Create branches
    o.progress.Info("Creating branches")
    branches, err := o.source.ListBranches()
    if err != nil {
        return fmt.Errorf("listing branches: %w", err)
    }
    
    for _, branch := range branches {
        gitBranchName := o.mapper.MapBranch(branch.Name)
        if err := o.target.CreateBranch(gitBranchName, branch.CommitID); err != nil {
            return fmt.Errorf("creating branch %s: %w", gitBranchName, err)
        }
    }
    
    // 7. Create tags
    o.progress.Info("Creating tags")
    tags, err := o.source.ListTags()
    if err != nil {
        return fmt.Errorf("listing tags: %w", err)
    }
    
    for _, tag := range tags {
        gitTagName := o.mapper.MapTag(tag.Name)
        if err := o.target.CreateTag(gitTagName, tag.CommitID, tag.Message); err != nil {
            return fmt.Errorf("creating tag %s: %w", gitTagName, err)
        }
    }
    
    // 8. Finalize
    o.progress.Info("Finalizing migration")
    if err := o.target.Finalize(); err != nil {
        return fmt.Errorf("finalizing: %w", err)
    }
    
    // 9. Mark complete
    state.Status = StatusComplete
    if err := o.stateStore.Save(state); err != nil {
        return fmt.Errorf("saving final state: %w", err)
    }
    
    o.progress.Complete()
    return nil
}

func (o *MigrationOrchestrator) applyMappings(commit *types.Commit) {
    // Map author
    commit.Author = o.mapper.MapAuthor(commit.Author.Name)
    commit.Committer = o.mapper.MapAuthor(commit.Committer.Name)
    
    // Branch mapping is done during branch creation
    // Tag mapping is done during tag creation
}
```

---

## Progress Reporting

### internal/progress/terminal.go

```go
package progress

import (
    "fmt"
    "github.com/schollz/progressbar/v3"
)

type TerminalProgress struct {
    bar     *progressbar.ProgressBar
    total   int
    current int
}

func NewTerminalProgress(total int) *TerminalProgress {
    return &TerminalProgress{
        total: total,
        bar: progressbar.NewOptions(total,
            progressbar.OptionSetDescription("Migrating commits"),
            progressbar.OptionSetWriter(os.Stderr),
            progressbar.OptionShowCount(),
            progressbar.OptionShowIts(),
            progressbar.OptionSetItsString("commits"),
        ),
    }
}

func (p *TerminalProgress) Update(current int, message string) {
    p.current = current
    p.bar.Add(1)
    if message != "" {
        fmt.Fprintf(os.Stderr, "\r%s", message)
    }
}

func (p *TerminalProgress) Complete() {
    p.bar.Finish()
    fmt.Fprintln(os.Stderr)
}

func (p *TerminalProgress) Error(err error) {
    fmt.Fprintf(os.Stderr, "\nError: %v\n", err)
}
```

### internal/progress/websocket.go

```go
package progress

import (
    "encoding/json"
    "sync"
    
    "github.com/gorilla/websocket"
)

type WebSocketProgress struct {
    clients map[*websocket.Conn]bool
    mu      sync.RWMutex
}

func NewWebSocketProgress() *WebSocketProgress {
    return &WebSocketProgress{
        clients: make(map[*websocket.Conn]bool),
    }
}

func (p *WebSocketProgress) AddClient(conn *websocket.Conn) {
    p.mu.Lock()
    defer p.mu.Unlock()
    p.clients[conn] = true
}

func (p *WebSocketProgress) RemoveClient(conn *websocket.Conn) {
    p.mu.Lock()
    defer p.mu.Unlock()
    delete(p.clients, conn)
}

func (p *WebSocketProgress) broadcast(message ProgressMessage) {
    p.mu.RLock()
    defer p.mu.RUnlock()
    
    data, err := json.Marshal(message)
    if err != nil {
        return
    }
    
    for client := range p.clients {
        client.WriteMessage(websocket.TextMessage, data)
    }
}

type ProgressMessage struct {
    Type    string `json:"type"` // update, complete, error
    Current int    `json:"current"`
    Total   int    `json:"total"`
    Message string `json:"message"`
}

func (p *WebSocketProgress) Update(current int, message string) {
    p.broadcast(ProgressMessage{
        Type:    "update",
        Current: current,
        Message: message,
    })
}

func (p *WebSocketProgress) Complete() {
    p.broadcast(ProgressMessage{
        Type: "complete",
    })
}
```

---

## Web Server

### internal/web/server.go

```go
package web

import (
    "embed"
    "net/http"
    
    "github.com/gorilla/websocket"
)

//go:embed static/*
var staticFiles embed.FS

type Server struct {
    port       int
    wsUpgrader websocket.Upgrader
    wsClients  map[*websocket.Conn]bool
}

func NewServer(port int) *Server {
    return &Server{
        port: port,
        wsUpgrader: websocket.Upgrader{
            CheckOrigin: func(r *http.Request) bool {
                return true // Allow all origins for MVP
            },
        },
        wsClients: make(map[*websocket.Conn]bool),
    }
}

func (s *Server) Start() error {
    // API routes
    http.HandleFunc("/api/migrations", s.handleMigrations)
    http.HandleFunc("/api/migrations/", s.handleMigration)
    http.HandleFunc("/ws/progress/", s.handleWebSocket)
    
    // Static files
    http.Handle("/", http.FileServer(http.FS(staticFiles)))
    
    return http.ListenAndServe(fmt.Sprintf(":%d", s.port), nil)
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
    conn, err := s.wsUpgrader.Upgrade(w, r, nil)
    if err != nil {
        return
    }
    defer conn.Close()
    
    s.wsClients[conn] = true
    
    // Keep connection alive
    for {
        _, _, err := conn.ReadMessage()
        if err != nil {
            delete(s.wsClients, conn)
            break
        }
    }
}
```

---

## Configuration Management

### internal/core/config.go

```go
package core

import (
    "fmt"
    "os"
    
    "github.com/spf13/viper"
    "github.com/adamf123git/git-migrator/pkg/types"
)

func LoadConfig(configPath string) (*types.Config, error) {
    v := viper.New()
    
    // Set defaults
    v.SetDefault("source.cvsMode", "auto")
    v.SetDefault("options.dryRun", false)
    v.SetDefault("options.preserveEmptyCommits", false)
    v.SetDefault("options.chunkSize", 100)
    v.SetDefault("options.verbose", false)
    v.SetDefault("options.resume", false)
    
    // Read config file
    if configPath != "" {
        v.SetConfigFile(configPath)
        if err := v.ReadInConfig(); err != nil {
            return nil, fmt.Errorf("reading config: %w", err)
        }
    }
    
    // Override with environment variables
    v.SetEnvPrefix("GIT_MIGRATOR")
    v.AutomaticEnv()
    
    // Unmarshal
    var config types.Config
    if err := v.Unmarshal(&config); err != nil {
        return nil, fmt.Errorf("unmarshaling config: %w", err)
    }
    
    // Validate
    if err := validateConfig(&config); err != nil {
        return nil, err
    }
    
    return &config, nil
}

func validateConfig(config *types.Config) error {
    if config.Source.Path == "" {
        return fmt.Errorf("source path is required")
    }
    if config.Target.Path == "" {
        return fmt.Errorf("target path is required")
    }
    if config.Source.Type == "" {
        return fmt.Errorf("source type is required")
    }
    if config.Target.Type == "" {
        config.Target.Type = "git" // Default
    }
    
    return nil
}
```

---

## CLI Commands

### cmd/git-migrator/commands/migrate.go

```go
package commands

import (
    "context"
    "fmt"
    "os"
    "os/signal"
    "syscall"
    
    "github.com/spf13/cobra"
    
    "github.com/adamf123git/git-migrator/internal/core"
    "github.com/adamf123git/git-migrator/internal/vcs/cvs"
    "github.com/adamf123git/git-migrator/internal/vcs/git"
)

var migrateCmd = &cobra.Command{
    Use:   "migrate",
    Short: "Migrate repository from source to target",
    Long:  `Migrate a version control repository from source (e.g., CVS) to target (e.g., Git).`,
    Run:   runMigrate,
}

func init() {
    rootCmd.AddCommand(migrateCmd)
    
    migrateCmd.Flags().StringP("config", "c", "", "Configuration file path")
    migrateCmd.Flags().String("source", "", "Source repository path")
    migrateCmd.Flags().String("target", "", "Target repository path")
    migrateCmd.Flags().Bool("dry-run", false, "Dry run mode")
    migrateCmd.Flags().Bool("resume", false, "Resume interrupted migration")
}

func runMigrate(cmd *cobra.Command, args []string) {
    // Load configuration
    configPath, _ := cmd.Flags().GetString("config")
    config, err := core.LoadConfig(configPath)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
        os.Exit(1)
    }
    
    // Override config with CLI flags
    applyCliFlags(cmd, config)
    
    // Create source reader
    var sourceReader vcs.VCSReader
    switch config.Source.Type {
    case "cvs":
        sourceReader, err = cvs.NewCVSReader(config.Source)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Error creating CVS reader: %v\n", err)
            os.Exit(1)
        }
    default:
        fmt.Fprintf(os.Stderr, "Unsupported source type: %s\n", config.Source.Type)
        os.Exit(1)
    }
    defer sourceReader.Close()
    
    // Create target writer
    var targetWriter vcs.VCSWriter
    switch config.Target.Type {
    case "git":
        targetWriter = git.NewGitWriter()
    default:
        fmt.Fprintf(os.Stderr, "Unsupported target type: %s\n", config.Target.Type)
        os.Exit(1)
    }
    defer targetWriter.Close()
    
    // Create orchestrator
    orchestrator := core.NewMigrationOrchestrator(sourceReader, targetWriter, config)
    
    // Setup signal handling for graceful shutdown
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
    
    go func() {
        <-sigChan
        fmt.Println("\nReceived interrupt, gracefully shutting down...")
        cancel()
    }()
    
    // Run migration
    if err := orchestrator.Migrate(ctx); err != nil {
        if err == context.Canceled {
            fmt.Println("Migration cancelled")
            os.Exit(130)
        }
        fmt.Fprintf(os.Stderr, "Migration failed: %v\n", err)
        os.Exit(1)
    }
    
    fmt.Println("Migration completed successfully!")
}
```

---

## State Management

### internal/core/state.go

```go
package core

import (
    "time"
)

type MigrationStatus string

const (
    StatusPending  MigrationStatus = "pending"
    StatusInProgress MigrationStatus = "in_progress"
    StatusComplete MigrationStatus = "complete"
    StatusFailed   MigrationStatus = "failed"
    StatusPaused   MigrationStatus = "paused"
)

type MigrationState struct {
    ID              string          `json:"id"`
    Status          MigrationStatus `json:"status"`
    SourcePath      string          `json:"sourcePath"`
    TargetPath      string          `json:"targetPath"`
    StartedAt       time.Time       `json:"startedAt"`
    CompletedAt     time.Time       `json:"completedAt,omitempty"`
    LastCommitID    string          `json:"lastCommitId"`
    ProcessedCommits int            `json:"processedCommits"`
    TotalCommits    int             `json:"totalCommits"`
    Errors          []MigrationError `json:"errors,omitempty"`
}

type MigrationError struct {
    CommitID  string    `json:"commitId"`
    Error     string    `json:"error"`
    Timestamp time.Time `json:"timestamp"`
}

func NewMigrationState(config *types.Config) *MigrationState {
    return &MigrationState{
        ID:         generateID(),
        Status:     StatusPending,
        SourcePath: config.Source.Path,
        TargetPath: config.Target.Path,
        StartedAt:  time.Now(),
    }
}

func (s *MigrationState) MarkFailed(commitID string, err error) {
    s.Errors = append(s.Errors, MigrationError{
        CommitID:  commitID,
        Error:     err.Error(),
        Timestamp: time.Now(),
    })
    s.Status = StatusFailed
}
```

---

## Error Handling

### internal/core/errors.go

```go
package core

import "fmt"

type MigrationError struct {
    Type    ErrorType
    Message string
    Cause   error
}

type ErrorType int

const (
    ErrorTypeValidation ErrorType = iota
    ErrorTypeSource
    ErrorTypeTarget
    ErrorTypeConfig
    ErrorTypeState
)

func (e *MigrationError) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("%s: %v", e.Message, e.Cause)
    }
    return e.Message
}

func (e *MigrationError) Unwrap() error {
    return e.Cause
}

func NewValidationError(message string, cause error) *MigrationError {
    return &MigrationError{
        Type:    ErrorTypeValidation,
        Message: message,
        Cause:   cause,
    }
}

func NewSourceError(message string, cause error) *MigrationError {
    return &MigrationError{
        Type:    ErrorTypeSource,
        Message: message,
        Cause:   cause,
    }
}
```

---

## Related Documents

- [Project Plan](./project-plan.md) - Overall project goals
- [Software Architecture](./software-architecture.md) - System architecture
- [Roadmap](./roadmap.md) - Development timeline

---

## Change Log

| Date | Version | Changes |
|------|---------|---------|
| 2025-01-18 | 1.0 | Initial design document |
