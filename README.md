# Git-Migrator

[![Go Report Card](https://goreportcard.com/badge/github.com/adamf123git/git-migrator)](https://goreportcard.com/report/github.com/adamf123git/git-migrator)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21%2B-00ADD8?logo=go)](https://golang.org)

**Migrate version control repositories with full history preservation**

Git-Migrator is an open-source tool for migrating repositories from legacy version control systems (CVS, SVN) to Git while preserving complete history, including commits, branches, tags, and author information. It can also keep repositories synchronized bidirectionally after migration.

## ✨ Features

- 🔄 **Full History Migration** - Preserve all commits, branches, tags, and metadata
- 🔌 **Plugin Architecture** - Support for multiple VCS systems (CVS, SVN, and more)
- 🐳 **Dual Runtime** - Run locally or in Docker with equal functionality
- 🖥️ **Dual Interface** - Command-line tool and web UI for monitoring
- ⏸️ **Resume Capability** - Resume interrupted migrations from where they left off
- 📊 **Progress Reporting** - Real-time progress via terminal or web dashboard
- 🎯 **Author Mapping** - Map CVS/SVN usernames to Git authors
- ✅ **Dry Run Mode** - Preview migrations without making changes
- 🔒 **Verification** - Validate migrated repositories match source

## 🚀 Quick Start

### Using Docker (Recommended)

```bash
# Pull the image
docker pull adamf123git/git-migrator:latest

# Run migration
docker run --rm \
  -v /path/to/cvs/repo:/source \
  -v /path/to/git/repo:/target \
  -v $(pwd)/config.yaml:/config.yaml \
  adamf123git/git-migrator migrate --config /config.yaml

# Or start web UI
docker run -d \
  -p 8080:8080 \
  -v /path/to/repos:/repos \
  adamf123git/git-migrator web
```

### Using Binary

```bash
# Download latest release
curl -sL https://github.com/adamf123git/git-migrator/releases/latest/download/git-migrator-linux-amd64 -o git-migrator
chmod +x git-migrator

# Run migration
./git-migrator migrate --config config.yaml
```

## 📦 Installation

### Pre-built Binaries

Download from [Releases](https://github.com/adamf123git/git-migrator/releases) page:

| Platform | Architecture | Download |
|----------|--------------|----------|
| Linux | amd64 | `git-migrator-linux-amd64` |
| Linux | arm64 | `git-migrator-linux-arm64` |
| macOS | amd64 | `git-migrator-darwin-amd64` |
| macOS | arm64 | `git-migrator-darwin-arm64` |
| Windows | amd64 | `git-migrator-windows-amd64.exe` |

### Homebrew (macOS/Linux)

```bash
brew tap adamf123git/git-migrator
brew install git-migrator
```

### From Source

```bash
git clone https://github.com/adamf123git/git-migrator.git
cd git-migrator
go build -o git-migrator ./cmd/git-migrator
```

### Docker

```bash
docker pull adamf123git/git-migrator:latest
```

## 📖 Usage

### CLI Commands

```bash
# Migrate CVS repository to Git
git-migrator migrate --config config.yaml

# Analyze source repository
git-migrator analyze --source-type cvs --source /path/to/cvs/repo

# Validate configuration
git-migrator validate --config config.yaml

# Extract author list from source repository
git-migrator authors extract --source-type cvs --source /path/to/cvs/repo > authors.txt

# Start web UI
git-migrator web --port 8080
```

### Configuration

Create a `config.yaml` file:

```yaml
source:
  type: cvs                    # cvs, svn (future)
  path: /path/to/cvs/repository
  module: mymodule             # CVS module name
  cvsMode: auto                # rcs, binary, auto

target:
  type: git
  path: /path/to/output/git/repository
  remote: git@github.com:user/repo.git  # Optional: push after migration

mapping:
  authors:
    cvsuser1: "John Doe <john@example.com>"
    cvsuser2: "Jane Smith <jane@example.com>"
  
  branches:
    "MAIN": "main"
    "RELEASE_1_0": "release/1.0"
  
  tags:
    "V1_0": "v1.0"
    "V2_0": "v2.0"

options:
  dryRun: false                # Preview without making changes
  preserveEmptyCommits: false  # Keep commits with no file changes
  chunkSize: 100               # State save interval
  verbose: false               # Detailed output
  resume: false                # Resume interrupted migration
```

### CLI Flags

Override configuration with flags:

```bash
git-migrator migrate \
  --source /path/to/cvs \
  --target /path/to/git \
  --source-type cvs \
  --dry-run \
  --resume \
  --verbose
```

### Web UI

Start the web interface:

```bash
# Local
git-migrator web --port 8080

# Docker
docker run -d -p 8080:8080 -v /path/to/repos:/repos adamf123git/git-migrator web
```

Then open http://localhost:8080 in your browser.

**Features:**
- Migration wizard
- Real-time progress dashboard
- Configuration editor
- Log viewer

## 🔧 Advanced Usage

### Resume Interrupted Migration

If a migration is interrupted (network issue, Ctrl+C, crash):

```bash
# Resume from last checkpoint
git-migrator migrate --config config.yaml --resume
```

State is saved every N commits (configurable via `chunkSize`).

### Dry Run

Preview migration without making changes:

```bash
git-migrator migrate --config config.yaml --dry-run --verbose
```

### Author Mapping

Extract authors from CVS repository and generate mapping template:

```bash
# Extract unique authors
git-migrator authors extract --source /path/to/cvs > authors.txt

# Edit authors.txt, then use in config:
# mapping:
#   authors:
#     cvsuser1: "Full Name <email@example.com>"
```

### Branch and Tag Mapping

Map CVS branch/tag names to Git equivalents:

```yaml
mapping:
  branches:
    "MAIN": "main"
    "DEV": "develop"
  
  tags:
    "release-1-0": "v1.0.0"
```

### Verification

After migration, verify repository integrity:

```bash
# Check commit count
git rev-list --all --count

# Verify all branches exist
git branch -a

# Verify all tags exist
git tag -l
```

## 🏗️ Architecture

Git-Migrator uses a **plugin-based architecture** for maximum extensibility:

```
┌─────────────────────────────────────────┐
│      Git-Migrator Core                  │
│                                         │
│  ┌──────────┐      ┌──────────┐        │
│  │ CVS      │      │ SVN      │        │
│  │ Plugin   │      │ Plugin   │ (Soon) │
│  └──────────┘      └──────────┘        │
│                                         │
│  ┌──────────────────────────────────┐  │
│  │      VCS Plugin Interface        │  │
│  └──────────────────────────────────┘  │
└─────────────────────────────────────────┘
```

**Key Components:**
- **VCS Interface** - Standardized API for version control operations
- **Migration Orchestrator** - Coordinates the migration workflow
- **State Manager** - Persists progress for resume capability
- **Progress Reporter** - Real-time updates (CLI + Web UI)

## 🤝 Contributing

We welcome contributions! Git-Migrator follows **Test-Driven Development (TDD)**.

### Development Setup

```bash
# Clone repository
git clone https://github.com/adamf123git/git-migrator.git
cd git-migrator

# Install dependencies
go mod download

# Run tests
make test

# Run linter
make lint

# Build
make build
```

### Development Workflow

1. **Write test first** (TDD)
2. **Implement feature**
3. **Ensure tests pass**: `make test`
4. **Check coverage**: `make test-coverage` (must be ≥ 80%)
5. **Run linter**: `make lint`
6. **Submit PR**

See [CONTRIBUTING.md](./CONTRIBUTING.md) for detailed guidelines.

### Running Tests

```bash
# Unit tests
make test-unit

# Integration tests
make test-integration

# Full regression suite
make test-regression

# Test coverage
make test-coverage

# All tests
make test
```

## 📚 Documentation

- [Getting Started](./docs/getting-started.md) - Detailed tutorial
- [Configuration Guide](./docs/configuration.md) - All configuration options
- [Migration Guide](./docs/migration.md) - Step-by-step migration
- [Architecture](./docs/software-architecture.md) - System design
- [API Reference](./docs/api.md) - REST API documentation

## 🐛 Troubleshooting

### Common Issues

**Issue:** CVS binary not found
```
Solution: Install CVS or use RCS mode (cvsMode: rcs in config)
```

**Issue:** Permission denied
```
Solution: Ensure read access to source, write access to target
```

**Issue:** Migration interrupted
```
Solution: Resume with --resume flag
```

**Issue:** Empty commits not preserved
```
Solution: Set preserveEmptyCommits: true in config
```

### Debug Mode

Enable verbose logging:

```bash
git-migrator migrate --config config.yaml --verbose
```

### Getting Help

- 📖 [Documentation](./docs/)
- 🐛 [Issue Tracker](https://github.com/adamf123git/git-migrator/issues)
- 💬 [Discussions](https://github.com/adamf123git/git-migrator/discussions)

## 🗺️ Roadmap

### Current Version (v1.0)
- ✅ CVS to Git migration
- ✅ Full history preservation
- ✅ Branch and tag migration
- ✅ Author mapping
- ✅ Resume capability
- ✅ Web UI for monitoring
- ✅ Docker support

### Coming Soon (v2.0)
- 🔜 Git ↔ CVS bidirectional sync
- 🔜 SVN to Git migration
- 🔜 Mercurial support

### Future
- 📋 Git LFS support
- 📋 Monorepo splitting
- 📋 Multi-repository batch migration
- 📋 Advanced conflict resolution

See [Roadmap](./docs/roadmap.md) for detailed timeline.

## 📄 License

Git-Migrator is open-source software licensed under the [MIT License](./LICENSE).

## 🙏 Acknowledgments

Inspired by and built upon the shoulders of:
- [cvs2git](https://www.march-hare.com/cvsnt/) - CVS to Git migration
- [git-svn](https://git-scm.com/docs/git-svn) - Git SVN bridge
- [go-git](https://github.com/go-git/go-git) - Pure Go Git implementation

## 📊 Project Status

- **Development Status:** Active development
- **Current Version:** v0.1.0 (pre-release)
- **Target Release:** v1.0.0 - Q1 2025

---

**Made with ❤️ by [Adam Farrell](https://github.com/adamf123git)**

**Star ⭐ this repository if you find it useful!**
