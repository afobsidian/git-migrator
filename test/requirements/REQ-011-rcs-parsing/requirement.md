# REQ-011: RCS File Parsing

## Requirement

The system must parse CVS RCS files (`,v` files) directly to extract version history, commit metadata, branches, and tags.

## Acceptance Criteria

- [ ] Parse RCS file header (head, branch, access, symbols, locks)
- [ ] Parse delta nodes (revision numbers, dates, authors, states)
- [ ] Parse delta text (commit messages and file contents)
- [ ] Handle binary files in RCS format
- [ ] Support keyword expansion detection
- [ ] Handle malformed RCS files gracefully
- [ ] Parse branch numbers and symbolic names
- [ ] Extract commit log messages
- [ ] All parsing is pure Go (no external dependencies)

## Test Coverage

- `rcs_test.go`: Validates all acceptance criteria
- Coverage: Must be 100% for this requirement

## Implementation

**Package:** `internal/vcs/cvs`
**Files:**
- `rcs_lexer.go` - Tokenizer for RCS format
- `rcs_parser.go` - Parser for RCS structures
- `rcs_types.go` - Data types for RCS data

## RCS File Format Reference

RCS files have the following structure:
```
head    1.3;
access;
symbols
    RELEASE_1_0:1.3
    BETA:1.2.0.2;
locks; strict;
comment @# @;

1.3
date    2024.01.15.10.30.00;  author john;  state Exp;
branches;
next    1.2;

1.2
date    2024.01.10.09.00.00;  author jane;  state Exp;
branches 1.2.2.1;
next    1.1;

desc
@@

1.3
log
@Added feature X
@
text
@... file content ...
@

1.2
log
@Fixed bug Y
@
text
@d3 1
a3 1
...
@
```

## Status

- [ ] Requirement defined
- [ ] Tests written
- [ ] Implementation complete
- [ ] All tests passing

## Related Requirements

- REQ-001: CVS to Git Migration
- REQ-012: CVS Repository Validation
