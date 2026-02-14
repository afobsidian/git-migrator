# Contributing to Git-Migrator

Thank you for your interest in contributing to Git-Migrator! This document provides guidelines and instructions for contributing.

## 🌟 Ways to Contribute

- **Report bugs** - Use the [issue tracker](https://github.com/adamf123git/git-migrator/issues)
- **Suggest features** - Start a [discussion](https://github.com/adamf123git/git-migrator/discussions)
- **Submit pull requests** - Fix bugs or add features
- **Improve documentation** - Make our docs better
- **Share feedback** - Tell us how you use Git-Migrator

## 🧪 Development Philosophy

### Test-Driven Development (TDD)

**Git-Migrator follows strict TDD practices. No code without tests.**

#### TDD Workflow

```
1. Write requirement document (test/requirements/REQ-XXX/)
2. Write failing test
3. Run test (confirm failure)
4. Write minimal code to pass
5. Run test (confirm pass)
6. Refactor if needed
7. Run regression tests
8. Commit
```

#### Example

```bash
# 1. Write test
cat > internal/vcs/cvs/rcs_test.go << 'EOF'
func TestParseRCSHeader(t *testing.T) {
    input := "head    1.12;"
    want := "1.12"
    got, err := ParseRCSHeader(input)
    assert.NoError(t, err)
    assert.Equal(t, want, got)
}
EOF

# 2. Run test (should fail)
make test-unit

# 3. Implement
cat > internal/vcs/cvs/rcs.go << 'EOF'
func ParseRCSHeader(input string) (string, error) {
    // Implementation
}
EOF

# 4. Run test (should pass)
make test-unit

# 5. Run regression
make test-regression

# 6. Commit
git commit -m "Add RCS header parsing"
```

## 🛠️ Development Setup

### Prerequisites

- **Go 1.21+** - [Install Go](https://golang.org/doc/install)
- **Make** - Build automation
- **Git** - Version control
- **Docker** (optional) - For containerized testing

### Getting Started

```bash
# 1. Fork the repository on GitHub

# 2. Clone your fork
git clone https://github.com/YOUR_USERNAME/git-migrator.git
cd git-migrator

# 3. Add upstream remote
git remote add upstream https://github.com/adamf123git/git-migrator.git

# 4. Install dependencies
go mod download

# 5. Run tests
make test

# 6. Run linter
make lint
```

### Development Tools

```bash
# Install development tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Install pre-commit hooks (optional but recommended)
make install-hooks
```

## 📋 Development Workflow

### Creating a Feature

1. **Create a branch**
   ```bash
   git checkout -b feature/my-feature
   ```

2. **Create requirement document**
   ```bash
   mkdir -p test/requirements/REQ-XXX-feature-name
   cat > test/requirements/REQ-XXX-feature-name/requirement.md << 'EOF'
   # REQ-XXX: Feature Name
   
   ## Requirement
   [Description]
   
   ## Acceptance Criteria
   - [ ] Criterion 1
   - [ ] Criterion 2
   EOF
   ```

3. **Write tests first** (TDD)
   ```bash
   # Write tests in test/requirements/REQ-XXX-feature-name/test.go
   ```

4. **Implement feature**
   ```bash
   # Write code to make tests pass
   ```

5. **Update requirements matrix**
   ```bash
   # Edit test/requirements/matrix_test.go
   ```

6. **Run tests**
   ```bash
   make test              # All tests
   make test-regression   # Regression suite
   make test-coverage     # Coverage check (≥ 80%)
   make test-requirements # Requirements validation
   ```

7. **Commit changes**
   ```bash
   git add .
   git commit -m "Add REQ-XXX: Feature description"
   ```

8. **Push and create PR**
   ```bash
   git push origin feature/my-feature
   # Create PR on GitHub
   ```

### Before Submitting PR

- [ ] All tests passing (`make test`)
- [ ] Coverage ≥ 80% (`make test-coverage`)
- [ ] No linter errors (`make lint`)
- [ ] Requirements documented (`test/requirements/REQ-XXX/`)
- [ ] Requirements matrix updated
- [ ] Documentation updated (if needed)
- [ ] CHANGELOG.md updated (if applicable)

## 🎨 Code Style

### Go Code

Follow standard Go conventions:

```bash
# Format code
go fmt ./...

# Run linter
golangci-lint run
```

### Key Guidelines

- Use `gofmt` for formatting
- Follow [Effective Go](https://golang.org/doc/effective_go)
- Write godoc comments for exported functions
- Keep functions small and focused
- Use table-driven tests

### Example Code Style

```go
// ParseRCSHeader extracts the head revision from an RCS file header.
// It expects input in the format "head    1.12;" and returns "1.12".
func ParseRCSHeader(input string) (string, error) {
    parts := strings.Fields(input)
    if len(parts) < 2 {
        return "", fmt.Errorf("invalid RCS header format: %s", input)
    }
    return strings.TrimSuffix(parts[1], ";"), nil
}
```

## 🧪 Testing

### Test Categories

```bash
# Unit tests (fast, isolated)
make test-unit

# Integration tests (slower, real components)
make test-integration

# Regression tests (full suite)
make test-regression

# Requirements validation
make test-requirements

# Coverage report
make test-coverage

# All tests
make test
```

### Writing Tests

Use table-driven tests:

```go
func TestParseRCSHeader(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {
            name:  "valid header",
            input: "head    1.12;",
            want:  "1.12",
        },
        {
            name:    "invalid format",
            input:   "invalid",
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ParseRCSHeader(tt.input)
            if tt.wantErr {
                assert.Error(t, err)
                return
            }
            assert.NoError(t, err)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

### Test Coverage

- **Minimum**: 80% overall
- **Core packages** (`internal/vcs`, `internal/core`): 90%
- **Per requirement**: 100% of acceptance criteria

Check coverage:

```bash
make test-coverage
# Open coverage.html in browser
```

## 📝 Documentation

### When to Update

- **New features**: Update user docs in `docs/`
- **API changes**: Update API reference
- **Bug fixes**: Update troubleshooting if applicable
- **Configuration**: Update configuration guide

### Documentation Style

- Use clear, simple language
- Include code examples
- Add diagrams where helpful (Mermaid)
- Keep lines under 80 characters

## 🐛 Reporting Issues

### Bug Reports

Include:

1. **Description** - What happened?
2. **Steps to reproduce** - How can we reproduce it?
3. **Expected behavior** - What should happen?
4. **Actual behavior** - What actually happened?
5. **Environment** - OS, Go version, Git-Migrator version
6. **Logs** - Error messages, verbose output

Use the bug report template when creating issues.

### Feature Requests

Include:

1. **Use case** - Why do you need this?
2. **Proposed solution** - How should it work?
3. **Alternatives** - What else have you considered?

## 🔀 Pull Request Process

1. **Create feature branch** from `main`
2. **Follow TDD workflow** (tests first!)
3. **Ensure all CI checks pass**
4. **Update documentation**
5. **Request review**
6. **Address review feedback**
7. **Squash commits** (if requested)

### PR Checklist

- [ ] Tests written first (TDD)
- [ ] All tests passing
- [ ] Coverage ≥ 80%
- [ ] No linter errors
- [ ] Requirements documented
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
- [ ] PR description clear

## 📜 Code of Conduct

### Our Standards

- Be respectful and inclusive
- Welcome newcomers
- Accept constructive criticism
- Focus on what's best for the community
- Show empathy towards others

### Unacceptable Behavior

- Harassment or discrimination
- Trolling or insulting comments
- Public or private harassment
- Publishing others' private information
- Other unprofessional conduct

## 📞 Getting Help

- 💬 [GitHub Discussions](https://github.com/adamf123git/git-migrator/discussions)
- 🐛 [Issue Tracker](https://github.com/adamf123git/git-migrator/issues)
- 📧 Email: adamf123git@example.com (replace with actual)

## 🙏 Recognition

Contributors are recognized in:

- [CONTRIBUTORS.md](./CONTRIBUTORS.md) file
- Release notes
- Project README

---

**Thank you for contributing to Git-Migrator! 🎉**

By participating in this project, you agree to abide by our Code of Conduct.
