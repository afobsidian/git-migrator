# Git-Migrator - Development Roadmap

**Status:** Internal Planning Document  
**Last Updated:** 2025-02-15  
**Version:** 1.1

---

## Overview

This roadmap outlines the development timeline for Git-Migrator, organized into phases and sprints. Each sprint is 2 weeks long and follows **Test-Driven Development (TDD)** with mandatory **regression testing** and **requirements validation**.

---

## Current Progress

| Sprint | Name | Status | Completion |
|--------|------|--------|------------|
| Sprint 1 | Foundation & Testing Infrastructure | ✅ Complete | 100% |
| Sprint 2 | CVS Reading & RCS Parsing | ✅ Complete | 100% |
| Sprint 3 | Git Writing & Commit Application | ✅ Complete | 100% |
| Sprint 4 | Migration Integration | ✅ Complete | 100% |
| Sprint 5 | Web UI & Docker | 🔜 Next | 0% |
| Sprint 6 | Testing, Polish & Release | ⚪ Planned | 0% |

---

## Timeline Overview

```mermaid
gantt
    title Git-Migrator Development Timeline
    dateFormat  YYYY-MM-DD
    
    section Phase 1: MVP
    Sprint 1: Foundation         :s1, 2025-01-20, 14d
    Sprint 2: CVS Reading        :s2, after s1, 14d
    Sprint 3: Git Writing        :s3, after s2, 14d
    Sprint 4: Integration        :s4, after s3, 14d
    Sprint 5: Testing & Polish   :s5, after s4, 14d
    Sprint 6: Release            :s6, after s5, 14d
    
    section Phase 2: Sync
    Sprint 7: Sync Architecture  :s7, after s6, 14d
    Sprint 8: Git ↔ CVS Sync     :s8, after s7, 14d
    Sprint 9: Sync Testing       :s9, after s8, 14d
    Sprint 10: Sync Release      :s10, after s9, 14d
    
    section Phase 3: SVN
    Sprint 11: SVN Reading       :s11, after s10, 14d
    Sprint 12: SVN Integration   :s12, after s11, 14d
```

---

## Phase 1: MVP - CVS to Git Migration

**Duration:** 12 weeks (6 sprints)  
**Goal:** Working CVS to Git migration tool with CLI + Web UI

### Sprint 1: Foundation & Testing Infrastructure (Weeks 1-2) ✅ COMPLETE

#### Objectives
- Set up project structure
- Implement TDD tooling
- Create requirements tracking system
- Establish CI/CD pipeline
- Basic CLI framework

#### Requirements
| ID | Requirement | Tests | Status |
|----|-------------|-------|--------|
| REQ-007 | CLI Interface | 12 | ✅ |
| REQ-009 | TDD with regression testing | 8 | ✅ |
| REQ-010 | Requirements validation | 5 | ✅ |

#### Completed Tasks
- [x] Project directory structure
- [x] Go module initialization
- [x] Makefile with test targets
- [x] Pre-commit hooks for TDD
- [x] CI/CD pipeline (`.github/workflows/ci.yml`)
- [x] Requirements tracking (`test/requirements/`)
- [x] Requirements matrix validation
- [x] CLI foundation (cobra)
- [x] Documentation (README, CONTRIBUTING)

#### Deliverables
- ✅ Working CLI with `version` command
- ✅ Full TDD infrastructure
- ✅ Requirements tracking system
- ✅ CI/CD pipeline
- ✅ All tests passing, >80% coverage

---

### Sprint 2: CVS Reading & RCS Parsing (Weeks 3-4) ✅ COMPLETE

#### Objectives
- Parse CVS RCS files directly
- Extract commits, branches, tags
- CVS repository validation

#### Requirements
| ID | Requirement | Tests | Status |
|----|-------------|-------|--------|
| REQ-001 | CVS to Git Migration (partial) | 30 | ✅ |
| REQ-011 | RCS file parsing | 20 | ✅ |
| REQ-012 | CVS repository validation | 8 | ✅ |

#### Completed Tasks
- [x] VCS Interface Design (`internal/vcs/vcs.go`)
- [x] RCS Lexer & Parser (`internal/vcs/cvs/rcs_lexer.go`, `rcs_parser.go`)
- [x] CVS Reader (`internal/vcs/cvs/reader.go`)
- [x] Commit iterator implementation
- [x] Test fixtures for CVS repos (`test/fixtures/cvs/`)
- [x] Integration tests using fixtures

#### Deliverables
- ✅ RCS file parser (pure Go)
- ✅ CVS repository reader
- ✅ Commit iterator
- ✅ Branch/tag extraction
- ✅ Test fixtures for CVS repos

---

### Sprint 3: Git Writing & Commit Application (Weeks 5-6) ✅ COMPLETE

#### Objectives
- Create Git repositories
- Apply commits to Git
- Create branches and tags
- Verify integrity

#### Requirements
| ID | Requirement | Tests | Status |
|----|-------------|-------|--------|
| REQ-001 | CVS to Git Migration (partial) | 30 | ✅ |
| REQ-013 | Git repository creation | 8 | ✅ |
| REQ-014 | Commit application | 15 | ✅ |
| REQ-015 | Branch/tag creation | 10 | ✅ |

#### Completed Tasks
- [x] Git Writer implementation (`internal/vcs/git/writer.go`)
- [x] Repository initialization with config
- [x] Commit application (add, modify, delete)
- [x] Branch creation
- [x] Tag creation (lightweight and annotated)
- [x] All tests passing

#### Deliverables
- ✅ Git repository writer
- ✅ Commit application
- ✅ Branch/tag creation
- ✅ Verification logic
- ✅ All Git operations tested

---

### Sprint 4: Migration Integration (Weeks 7-8) ✅ COMPLETE

#### Objectives
- End-to-end migration pipeline
- Author mapping
- Progress reporting
- Resume capability
- State persistence

#### Requirements
| ID | Requirement | Tests | Status |
|----|-------------|-------|--------|
| REQ-001 | CVS to Git Migration | 30 | ✅ |
| REQ-002 | Author mapping | 10 | ✅ |
| REQ-005 | Resume capability | 12 | ✅ |
| REQ-016 | Progress reporting | 8 | ✅ |
| REQ-017 | State persistence | 10 | ✅ |

#### Completed Tasks
- [x] Migration orchestrator (`internal/core/migration.go`)
- [x] Author mapping (`internal/mapping/authors.go`)
- [x] Progress reporting (`internal/progress/reporter.go`)
- [x] State persistence (`internal/storage/state.go`)
- [x] Resume capability (`internal/core/migration.go`)
- [x] SQLite storage with modernc.org/sqlite driver
- [x] All tests passing

#### Deliverables
- ✅ Migration orchestrator
- ✅ Author mapping
- ✅ Progress reporting
- ✅ Resume capability
- ✅ SQLite state storage

---

### Sprint 5: Web UI & Docker (Weeks 9-10)

#### Objectives
- Web UI for migration monitoring
- WebSocket real-time updates
- Docker containerization
- REST API

#### Requirements
| ID | Requirement | Tests | Status |
|----|-------------|-------|--------|
| REQ-006 | Docker support | 10 | ⚪ |
| REQ-008 | Web UI | 15 | ⚪ |
| REQ-018 | REST API | 12 | ⚪ |
| REQ-019 | WebSocket progress | 8 | ⚪ |

#### Tasks

**Day 1-4: Web Server & API**
- [ ] Create `test/requirements/REQ-018-rest-api/`
- [ ] Write tests for HTTP server
- [ ] Write tests for API endpoints
- [ ] Implement `internal/web/server.go`
- [ ] Implement `internal/web/api.go`
- [ ] Test: API endpoints functional

**Day 5-7: WebSocket Progress**
- [ ] Create `test/requirements/REQ-019-websocket/`
- [ ] Write tests for WebSocket handler
- [ ] Implement `internal/web/websocket.go`
- [ ] Implement `internal/progress/websocket.go`
- [ ] Test: Real-time progress in browser

**Day 8-10: Web UI**
- [ ] Create `test/requirements/REQ-008-web-ui/`
- [ ] Write E2E tests (Playwright)
- [ ] Create `internal/web/static/index.html`
- [ ] Create `internal/web/static/app.js`
- [ ] Create `internal/web/static/style.css`
- [ ] Test: UI shows migration progress

**Day 11-12: Docker**
- [ ] Create `test/requirements/REQ-006-docker/`
- [ ] Write tests for Docker execution
- [ ] Create `Dockerfile`
- [ ] Create `docker-compose.yml`
- [ ] Test: Docker container runs migrations

**Day 13-14: Integration & Regression**
- [ ] Run: `make test-regression`
- [ ] Test: Full migration via Web UI
- [ ] Test: Full migration via Docker
- [ ] Fix any failures
- [ ] Update requirements matrix
- [ ] Commit and push

#### Deliverables
- ✅ Web UI for monitoring
- ✅ WebSocket real-time updates
- ✅ REST API
- ✅ Docker support
- ✅ E2E tests for Web UI

---

### Sprint 6: Testing, Polish & Release (Weeks 11-12)

#### Objectives
- Comprehensive testing
- Performance optimization
- Documentation
- Release preparation

#### Tasks

**Day 1-3: Comprehensive Testing**
- [ ] Run: `make test-nightly` (large repos)
- [ ] Performance benchmarks
- [ ] Memory profiling
- [ ] Fix performance issues
- [ ] Test: 10k+ commit repos

**Day 4-6: Documentation**
- [ ] Update `README.md`
- [ ] Create `docs/getting-started.md`
- [ ] Create `docs/migration.md`
- [ ] Create `docs/configuration.md`
- [ ] Add inline godoc comments
- [ ] Generate godoc documentation

**Day 7-9: Polish & Bug Fixes**
- [ ] Address all TODOs
- [ ] Fix linter warnings
- [ ] Improve error messages
- [ ] Add helpful debug logging
- [ ] Code review

**Day 10-12: Release Preparation**
- [ ] Create `.goreleaser.yml`
- [ ] Build binaries for all platforms
- [ ] Test binaries on Linux, macOS, Windows
- [ ] Create GitHub release
- [ ] Push Docker image to Docker Hub
- [ ] Create Homebrew formula

**Day 13-14: Launch**
- [ ] Final regression testing
- [ ] Create GitHub release notes
- [ ] Announce on social media
- [ ] Update project website

#### Deliverables
- ✅ 90%+ test coverage
- ✅ Performance benchmarks
- ✅ Complete documentation
- ✅ Release binaries
- ✅ Docker image
- ✅ Homebrew formula
- ✅ GitHub release

---

## Phase 2: Bidirectional Sync

**Duration:** 8 weeks (4 sprints)  
**Goal:** Git ↔ CVS bidirectional synchronization

### Sprint 7-8: Sync Architecture & Implementation

#### Requirements
| ID | Requirement | Tests | Status |
|----|-------------|-------|--------|
| REQ-100 | Git ↔ CVS Sync | 40 | ⚪ |
| REQ-101 | Change detection | 15 | ⚪ |
| REQ-102 | Conflict detection | 12 | ⚪ |

#### Key Tasks
- Design `VCSSyncer` interface
- Implement change detection
- Implement conflict detection
- Implement sync orchestrator
- Write comprehensive tests
- Test with real repositories

### Sprint 9-10: Sync Testing & Release

#### Key Tasks
- Sync integration tests
- Performance testing
- Documentation updates
- Release sync feature

---

## Phase 3: SVN Support

**Duration:** 6 weeks (3 sprints)  
**Goal:** SVN to Git migration

### Sprint 11-12: SVN Reading & Integration

#### Requirements
| ID | Requirement | Tests | Status |
|----|-------------|-------|--------|
| REQ-200 | SVN support | 35 | ⚪ |
| REQ-201 | SVN externals | 10 | ⚪ |

#### Key Tasks
- Implement `internal/vcs/svn/reader.go`
- Parse SVN dump format
- Handle SVN externals
- SVN properties mapping
- Write comprehensive tests
- Test with real SVN repositories

---

## Phase 4: Extended Features (Ongoing)

### Future Requirements

| ID | Requirement | Priority |
|----|-------------|----------|
| REQ-300 | Mercurial support | P3 |
| REQ-400 | Git LFS support | P3 |
| REQ-500 | Monorepo splitting | P3 |
| REQ-600 | Multi-repo batch migration | P3 |
| REQ-700 | Web UI authentication | P2 |

---

## Release Cadence

### MVP Release (v1.0.0)
- **Target:** End of Sprint 6 (Week 12)
- **Scope:** CVS to Git migration
- **Platforms:** Linux, macOS, Windows
- **Distribution:** Binary, Docker, Homebrew

### Sync Release (v2.0.0)
- **Target:** End of Sprint 10 (Week 20)
- **Scope:** Bidirectional Git ↔ CVS sync

### SVN Release (v3.0.0)
- **Target:** End of Sprint 12 (Week 24)
- **Scope:** SVN to Git migration

### Minor Releases
- **Frequency:** Monthly (after v1.0.0)
- **Scope:** Bug fixes, minor features, performance improvements

---

## Quality Gates

### Every Sprint

- [ ] All tests passing (`make test`)
- [ ] Coverage ≥ 80% (`make test-coverage`)
- [ ] No linter errors (`golangci-lint run`)
- [ ] Requirements validated (`make test-requirements`)
- [ ] Regression suite passing (`make test-regression`)
- [ ] Documentation updated
- [ ] `test/requirements/STATUS.md` updated

### Every Release

- [ ] All sprint quality gates
- [ ] Nightly tests passing (`make test-nightly`)
- [ ] Performance benchmarks acceptable
- [ ] Cross-platform testing (Linux, macOS, Windows)
- [ ] Docker image tested
- [ ] Documentation complete
- [ ] Changelog updated

---

## Risk Mitigation

### Technical Risks

**Risk:** RCS format complexity  
**Mitigation:** Sprint 2 has extra buffer time (14 days for parsing)  
**Contingency:** Fallback to CVS binary if RCS parsing fails

**Risk:** Large repository performance  
**Mitigation:** Sprint 6 includes performance testing  
**Contingency:** Add streaming/chunking if needed

### Schedule Risks

**Risk:** Scope creep  
**Mitigation:** Strict MVP definition, modular architecture  
**Contingency:** Defer non-essential features to Phase 4

**Risk:** Test coverage gaps  
**Mitigation:** Requirements traceability matrix  
**Contingency:** Add testing sprint buffer

---

## Success Metrics

### Sprint-Level

- Test coverage: ≥ 80%
- Requirements coverage: 100%
- Zero regressions
- All acceptance criteria met

### Release-Level

- Successfully migrate 10k+ commit repo
- Preserve 100% of history
- All branches/tags migrated
- Migration time < 1 hour (typical repo)
- GitHub stars > 100 (v1.0)
- Docker pulls > 100 (v1.0)

---

## Monitoring Progress

### Weekly Status Updates

Update `test/requirements/STATUS.md` with:
- Requirements completed
- Tests written/passed
- Coverage percentage
- Blockers and risks

### Sprint Retrospectives

- What went well?
- What didn't go well?
- What to improve next sprint?
- Update roadmap if needed

---

## Related Documents

- [Project Plan](./project-plan.md) - Overall project goals
- [Software Architecture](./software-architecture.md) - System architecture
- [Software Design](./software-design.md) - Implementation details

---

## Change Log

| Date | Version | Changes |
|------|---------|---------|
| 2025-02-15 | 1.2 | Sprint 4 completed; migration integration done |
| 2025-02-15 | 1.1 | Sprint 2 & 3 completed; updated progress |
| 2025-01-18 | 1.0 | Initial roadmap |
