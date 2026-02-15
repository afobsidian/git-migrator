# Git-Migrator Documentation

This directory contains comprehensive internal planning documentation for the Git-Migrator project.

## Documents

### 📋 [Project Plan](./project-plan.md)
**High-level overview and goals**

- Executive summary
- Core requirements
- MVP scope
- Technology stack
- Quality assurance strategy
- Success metrics
- Risk management

### 🏗️ [Software Architecture](./software-architecture.md)
**System design and component interactions**

- Layered architecture
- Plugin-based VCS system
- Component diagrams
- Data flow
- Deployment architecture
- Security architecture
- Extensibility points

### 💻 [Software Design](./software-design.md)
**Implementation details**

- Package structure
- Core types and data structures
- Interface definitions
- Algorithm implementations
- Code examples
- Detailed workflows

### 📅 [Roadmap](./roadmap.md)
**Development timeline and sprints**

- Phase breakdown (MVP, Sync, SVN)
- Sprint-by-sprint tasks
- Release schedule
- Quality gates
- Risk mitigation
- Success metrics

## Quick Navigation

| Question | Document |
|----------|----------|
| What are we building? | [Project Plan](./project-plan.md) |
| What are the requirements? | [Project Plan - Requirements Table](./project-plan.md#core-requirements) |
| How does the system work? | [Software Architecture](./software-architecture.md) |
| How is the code organized? | [Software Design - Package Structure](./software-design.md#package-structure) |
| What are the key interfaces? | [Software Design - VCS Interfaces](./software-design.md#vcs-interface-design) |
| What's the timeline? | [Roadmap](./roadmap.md) |
| What are we doing this sprint? | [Roadmap - Sprint Details](./roadmap.md#phase-1-mvp---cvs-to-git-migration) |
| How do we track requirements? | [Project Plan - Quality Assurance](./project-plan.md#quality-assurance-strategy) |

## Document Status

| Document | Status | Last Updated | Version |
|----------|--------|--------------|---------|
| Project Plan | ✅ Complete | 2025-01-18 | 1.0 |
| Software Architecture | ✅ Complete | 2025-01-18 | 1.0 |
| Software Design | ✅ Complete | 2025-01-18 | 1.0 |
| Roadmap | ✅ Complete | 2025-02-15 | 1.1 |

## Related Resources

### External Documentation
- [../README.md](../README.md) - User-facing documentation
- [../CONTRIBUTING.md](../CONTRIBUTING.md) - Contribution guidelines

### Planning Artifacts
- [../test/requirements/](../test/requirements/) - Requirements tracking
- [../test/requirements/STATUS.md](../test/requirements/STATUS.md) - Requirements status

### Diagrams

All diagrams use **Mermaid** syntax and can be rendered in:
- GitHub/GitLab markdown preview
- VS Code with Mermaid extension
- Online: [mermaid.live](https://mermaid.live/)

## Updating These Documents

### When to Update

- **Project Plan:** When requirements change, milestones reached, risks identified
- **Software Architecture:** When adding new components, changing design
- **Software Design:** When implementing new features, changing APIs
- **Roadmap:** At sprint boundaries, when timeline changes

### Update Process

1. Update the relevant document
2. Update the "Last Updated" date and version
3. Update the document status table in this README
4. Commit with message: "docs: update [document-name]"

## Diagram Index

### Architecture Diagrams
- [High-Level Architecture](./software-architecture.md#architecture-overview) - System layers
- [Component Interaction](./software-architecture.md#component-interaction) - Migration workflow
- [Deployment Architecture](./software-architecture.md#deployment-architecture) - Local vs Docker
- [Data Flow](./software-architecture.md#data-flow) - Commit processing
- [Test Pyramid](./software-architecture.md#testing-architecture) - Test categories

### Workflow Diagrams
- [Migration Workflow](./software-architecture.md#migration-workflow) - Sequence diagram
- [Resume Capability](./software-architecture.md#resume-capability) - State diagram

### Timeline Diagrams
- [Timeline Overview](./roadmap.md#timeline-overview) - Gantt chart

---

**Note:** These are **internal planning documents** not intended for public audiences. For user-facing documentation, see the main [README.md](../README.md).
