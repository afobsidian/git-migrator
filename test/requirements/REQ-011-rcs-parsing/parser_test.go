package requirements

import (
	"strings"
	"testing"
	"time"

	"github.com/adamf123git/git-migrator/internal/vcs/cvs"
)

// TestRCSParserHeader tests parsing RCS header
func TestRCSParserHeader(t *testing.T) {
	input := `head 1.3;
branch 1.2;
access;
symbols RELEASE_1_0:1.3 BETA:1.2;
locks; strict;
comment @# @;`

	parser := cvs.NewRCSParser(strings.NewReader(input))
	rcsFile, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if rcsFile.Head != "1.3" {
		t.Errorf("Expected head 1.3, got %s", rcsFile.Head)
	}

	if rcsFile.Branch != "1.2" {
		t.Errorf("Expected branch 1.2, got %s", rcsFile.Branch)
	}

	if len(rcsFile.Symbols) != 2 {
		t.Errorf("Expected 2 symbols, got %d", len(rcsFile.Symbols))
	}

	if rcsFile.Symbols["RELEASE_1_0"] != "1.3" {
		t.Errorf("Expected RELEASE_1_0=1.3, got %s", rcsFile.Symbols["RELEASE_1_0"])
	}
}

// TestRCSParserDeltas tests parsing delta nodes
func TestRCSParserDeltas(t *testing.T) {
	input := `head 1.2;
1.2
date 2024.01.15.10.30.00; author john; state Exp;
branches;
next 1.1;

1.1
date 2024.01.10.09.00.00; author jane; state Exp;
branches;
next;

desc
@@`

	parser := cvs.NewRCSParser(strings.NewReader(input))
	rcsFile, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(rcsFile.Deltas) != 2 {
		t.Fatalf("Expected 2 deltas, got %d", len(rcsFile.Deltas))
	}

	// Check first delta
	delta := rcsFile.Deltas["1.2"]
	if delta == nil {
		t.Fatal("Delta 1.2 not found")
	}
	if delta.Author != "john" {
		t.Errorf("Expected author john, got %s", delta.Author)
	}
	if delta.State != "Exp" {
		t.Errorf("Expected state Exp, got %s", delta.State)
	}

	// Check date parsing
	expectedDate := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	if !delta.Date.Equal(expectedDate) {
		t.Errorf("Expected date %v, got %v", expectedDate, delta.Date)
	}
}

// TestRCSParserBranches tests parsing branch information
func TestRCSParserBranches(t *testing.T) {
	input := `head 1.3;
1.3
date 2024.01.15.10.30.00; author john; state Exp;
branches 1.2.2.1;
next 1.2;

1.2
date 2024.01.10.09.00.00; author jane; state Exp;
branches;
next 1.1;

1.2.2.1
date 2024.01.12.14.00.00; author bob; state Exp;
branches;
next;

desc
@@`

	parser := cvs.NewRCSParser(strings.NewReader(input))
	rcsFile, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	delta := rcsFile.Deltas["1.3"]
	if delta == nil {
		t.Fatal("Delta 1.3 not found")
	}

	if len(delta.Branches) != 1 {
		t.Errorf("Expected 1 branch, got %d", len(delta.Branches))
	}

	if delta.Branches[0] != "1.2.2.1" {
		t.Errorf("Expected branch 1.2.2.1, got %s", delta.Branches[0])
	}
}

// TestRCSParserDescription tests parsing description
func TestRCSParserDescription(t *testing.T) {
	input := `head 1.1;
desc
@This is the file description.
It can span multiple lines.@`

	parser := cvs.NewRCSParser(strings.NewReader(input))
	rcsFile, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	expected := "This is the file description.\nIt can span multiple lines."
	if rcsFile.Description != expected {
		t.Errorf("Expected description %q, got %q", expected, rcsFile.Description)
	}
}

// TestRCSParserLogMessages tests parsing log messages
func TestRCSParserLogMessages(t *testing.T) {
	input := `head 1.2;
1.2
date 2024.01.15.10.30.00; author john; state Exp;
branches;
next 1.1;

1.1
date 2024.01.10.09.00.00; author jane; state Exp;
branches;
next;

desc
@@

1.2
log
@Added new feature X
@
text
@@

1.1
log
@Initial commit
@
text
@@`

	parser := cvs.NewRCSParser(strings.NewReader(input))
	rcsFile, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	delta2 := rcsFile.Deltas["1.2"]
	if delta2 == nil {
		t.Fatal("Delta 1.2 not found")
	}
	if delta2.Log != "Added new feature X" {
		t.Errorf("Expected log 'Added new feature X', got %q", delta2.Log)
	}

	delta1 := rcsFile.Deltas["1.1"]
	if delta1 == nil {
		t.Fatal("Delta 1.1 not found")
	}
	if delta1.Log != "Initial commit" {
		t.Errorf("Expected log 'Initial commit', got %q", delta1.Log)
	}
}

// TestRCSParserSymbols tests parsing symbolic names
func TestRCSParserSymbols(t *testing.T) {
	input := `head 1.5;
symbols
	RELEASE_1_0:1.3
	RELEASE_2_0:1.5
	BETA_1:1.2.2.1
	dev_branch:1.2.0.2;`

	parser := cvs.NewRCSParser(strings.NewReader(input))
	rcsFile, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	expected := map[string]string{
		"RELEASE_1_0": "1.3",
		"RELEASE_2_0": "1.5",
		"BETA_1":      "1.2.2.1",
		"dev_branch":  "1.2.0.2",
	}

	for sym, rev := range expected {
		if rcsFile.Symbols[sym] != rev {
			t.Errorf("Expected symbol %s=%s, got %s", sym, rev, rcsFile.Symbols[sym])
		}
	}
}

// TestRCSParserGetCommits tests extracting commits in order
func TestRCSParserGetCommits(t *testing.T) {
	input := `head 1.3;
1.3
date 2024.01.20.10.00.00; author alice; state Exp;
branches;
next 1.2;

1.2
date 2024.01.15.10.00.00; author bob; state Exp;
branches;
next 1.1;

1.1
date 2024.01.10.10.00.00; author carol; state Exp;
branches;
next;

desc
@@`

	parser := cvs.NewRCSParser(strings.NewReader(input))
	rcsFile, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	commits := rcsFile.GetCommits()
	if len(commits) != 3 {
		t.Fatalf("Expected 3 commits, got %d", len(commits))
	}

	// Commits should be in reverse chronological order (newest first)
	if commits[0].Author != "alice" {
		t.Errorf("Expected first commit by alice, got %s", commits[0].Author)
	}
	if commits[2].Author != "carol" {
		t.Errorf("Expected last commit by carol, got %s", commits[2].Author)
	}
}

// TestRCSParserGetBranches tests extracting branches
func TestRCSParserGetBranches(t *testing.T) {
	input := `head 1.3;
symbols FEATURE_X:1.2.0.2 BUGFIX:1.2.2.1;
1.3
date 2024.01.20.10.00.00; author alice; state Exp;
branches;
next 1.2;

1.2
date 2024.01.15.10.00.00; author bob; state Exp;
branches 1.2.2.1;
next 1.1;

1.2.2.1
date 2024.01.17.10.00.00; author charlie; state Exp;
branches;
next;

1.1
date 2024.01.10.10.00.00; author dave; state Exp;
branches;
next;

desc
@@`

	parser := cvs.NewRCSParser(strings.NewReader(input))
	rcsFile, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	branches := rcsFile.GetBranches()
	// Should have at least the branches from symbols
	if len(branches) < 2 {
		t.Errorf("Expected at least 2 branches, got %d", len(branches))
	}
}

// TestRCSParserGetTags tests extracting tags
func TestRCSParserGetTags(t *testing.T) {
	input := `head 1.3;
symbols RELEASE_1_0:1.3 FEATURE_BRANCH:1.2.0.2;
1.3
date 2024.01.20.10.00.00; author alice; state Exp;
branches;
next 1.2;

1.2
date 2024.01.15.10.00.00; author bob; state Exp;
branches;
next 1.1;

1.1
date 2024.01.10.10.00.00; author carol; state Exp;
branches;
next;

desc
@@`

	parser := cvs.NewRCSParser(strings.NewReader(input))
	rcsFile, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	tags := rcsFile.GetTags()
	// RELEASE_1_0 points to 1.3 which is on trunk, so it's a tag
	if _, ok := tags["RELEASE_1_0"]; !ok {
		t.Error("Expected RELEASE_1_0 to be a tag")
	}
}
