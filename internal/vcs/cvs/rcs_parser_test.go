package cvs

import (
	"strings"
	"testing"
	"time"
)

func TestNewRCSParser(t *testing.T) {
	input := "head 1.0;"
	parser := NewRCSParser(strings.NewReader(input))

	if parser == nil {
		t.Fatal("NewRCSParser returned nil")
	}

	if parser.lexer == nil {
		t.Error("lexer should be initialized")
	}
}

func TestNewRCSParserEmpty(t *testing.T) {
	parser := NewRCSParser(strings.NewReader(""))

	if parser == nil {
		t.Fatal("NewRCSParser returned nil")
	}
}

func TestParserParseEmpty(t *testing.T) {
	parser := NewRCSParser(strings.NewReader(""))

	rcs, err := parser.Parse()
	if err != nil {
		t.Errorf("Parse failed: %v", err)
	}

	if rcs == nil {
		t.Fatal("Parse returned nil")
	}

	if rcs.Deltas == nil {
		t.Error("Deltas map should be initialized")
	}

	if rcs.Symbols == nil {
		t.Error("Symbols map should be initialized")
	}

	if rcs.Locks == nil {
		t.Error("Locks map should be initialized")
	}
}

func TestParserParseHead(t *testing.T) {
	input := "head 1.5;"
	parser := NewRCSParser(strings.NewReader(input))

	rcs, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if rcs.Head != "1.5" {
		t.Errorf("Head = %q, want %q", rcs.Head, "1.5")
	}
}

func TestParserParseBranch(t *testing.T) {
	input := "head 1.5; branch 1.5.2;"
	parser := NewRCSParser(strings.NewReader(input))

	rcs, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if rcs.Branch != "1.5.2" {
		t.Errorf("Branch = %q, want %q", rcs.Branch, "1.5.2")
	}
}

func TestParserParseAccess(t *testing.T) {
	input := "head 1.5; access johndoe janedoe admin;"
	parser := NewRCSParser(strings.NewReader(input))

	rcs, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(rcs.Access) != 3 {
		t.Errorf("Access length = %d, want 3", len(rcs.Access))
	}

	expected := []string{"johndoe", "janedoe", "admin"}
	for i, exp := range expected {
		if i >= len(rcs.Access) {
			continue
		}
		if rcs.Access[i] != exp {
			t.Errorf("Access[%d] = %q, want %q", i, rcs.Access[i], exp)
		}
	}
}

func TestParserParseAccessEmpty(t *testing.T) {
	input := "head 1.5; access;"
	parser := NewRCSParser(strings.NewReader(input))

	rcs, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(rcs.Access) != 0 {
		t.Errorf("Access length = %d, want 0", len(rcs.Access))
	}
}

func TestParserParseSymbols(t *testing.T) {
	input := "head 1.5; symbols REL_1_0:1.4 DEV:1.4.2.1;"
	parser := NewRCSParser(strings.NewReader(input))

	rcs, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(rcs.Symbols) != 2 {
		t.Errorf("Symbols length = %d, want 2", len(rcs.Symbols))
	}

	if rcs.Symbols["REL_1_0"] != "1.4" {
		t.Errorf("Symbols[REL_1_0] = %q, want %q", rcs.Symbols["REL_1_0"], "1.4")
	}

	if rcs.Symbols["DEV"] != "1.4.2.1" {
		t.Errorf("Symbols[DEV] = %q, want %q", rcs.Symbols["DEV"], "1.4.2.1")
	}
}

func TestParserParseSymbolsEmpty(t *testing.T) {
	input := "head 1.5; symbols;"
	parser := NewRCSParser(strings.NewReader(input))

	rcs, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(rcs.Symbols) != 0 {
		t.Errorf("Symbols length = %d, want 0", len(rcs.Symbols))
	}
}

func TestParserParseLocks(t *testing.T) {
	input := "head 1.5; locks johndoe:1.5 janedoe:1.4;"
	parser := NewRCSParser(strings.NewReader(input))

	rcs, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(rcs.Locks) != 2 {
		t.Errorf("Locks length = %d, want 2", len(rcs.Locks))
	}

	if rcs.Locks["johndoe"] != "1.5" {
		t.Errorf("Locks[johndoe] = %q, want %q", rcs.Locks["johndoe"], "1.5")
	}

	if rcs.Locks["janedoe"] != "1.4" {
		t.Errorf("Locks[janedoe] = %q, want %q", rcs.Locks["janedoe"], "1.4")
	}
}

func TestParserParseStrictLocks(t *testing.T) {
	input := "head 1.5; strict;"
	parser := NewRCSParser(strings.NewReader(input))

	rcs, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if !rcs.StrictLocks {
		t.Error("StrictLocks should be true")
	}
}

func TestParserParseComment(t *testing.T) {
	input := "head 1.5; comment @# @;"
	parser := NewRCSParser(strings.NewReader(input))

	rcs, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if rcs.Comment != "# " {
		t.Errorf("Comment = %q, want %q", rcs.Comment, "# ")
	}
}

func TestParserParseDescription(t *testing.T) {
	input := "head 1.5; desc @This is a test file@;"
	parser := NewRCSParser(strings.NewReader(input))

	rcs, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if rcs.Description != "This is a test file" {
		t.Errorf("Description = %q, want %q", rcs.Description, "This is a test file")
	}
}

func TestParserParseDelta(t *testing.T) {
	input := `head 1.5;
1.5
date 2024.1.15.12.30.45; author johndoe; state Exp;
branches 1.5.2.1;
next 1.4;
desc @@;`

	parser := NewRCSParser(strings.NewReader(input))

	rcs, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	delta, ok := rcs.Deltas["1.5"]
	if !ok {
		t.Fatal("Delta 1.5 not found")
	}

	if delta.Revision != "1.5" {
		t.Errorf("Revision = %q, want %q", delta.Revision, "1.5")
	}

	if delta.Author != "johndoe" {
		t.Errorf("Author = %q, want %q", delta.Author, "johndoe")
	}

	if delta.State != "Exp" {
		t.Errorf("State = %q, want %q", delta.State, "Exp")
	}

	if delta.Next != "1.4" {
		t.Errorf("Next = %q, want %q", delta.Next, "1.4")
	}

	if len(delta.Branches) != 1 || delta.Branches[0] != "1.5.2.1" {
		t.Errorf("Branches = %v, want [1.5.2.1]", delta.Branches)
	}

	// Check date
	expectedDate := time.Date(2024, 1, 15, 12, 30, 45, 0, time.UTC)
	if !delta.Date.Equal(expectedDate) {
		t.Errorf("Date = %v, want %v", delta.Date, expectedDate)
	}
}

func TestParserParseMultipleDeltas(t *testing.T) {
	input := `head 1.5;
1.5
date 2024.1.15.12.30.45; author user3; state Exp;
next 1.4;

1.4
date 2024.1.10.10.20.30; author user2; state Exp;
next 1.3;

1.3
date 2024.1.5.8.15.0; author user1; state Exp;
desc @@;`

	parser := NewRCSParser(strings.NewReader(input))

	rcs, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(rcs.Deltas) != 3 {
		t.Errorf("Deltas length = %d, want 3", len(rcs.Deltas))
	}

	// Check DeltaOrder
	if len(rcs.DeltaOrder) != 3 {
		t.Errorf("DeltaOrder length = %d, want 3", len(rcs.DeltaOrder))
	}

	expectedOrder := []string{"1.5", "1.4", "1.3"}
	for i, exp := range expectedOrder {
		if i >= len(rcs.DeltaOrder) {
			continue
		}
		if rcs.DeltaOrder[i] != exp {
			t.Errorf("DeltaOrder[%d] = %q, want %q", i, rcs.DeltaOrder[i], exp)
		}
	}
}

func TestParserParseDeltaTexts(t *testing.T) {
	// Delta texts are parsed after desc section
	input := `head 1.2;
1.2
date 2024.1.15.12.30.0; author test; state Exp;
1.1
date 2024.1.10.10.0.0; author test; state Exp;
desc @@;
1.2
log @Added feature X@
text @new content@;
1.1
log @Initial revision@
text @initial content@;`

	parser := NewRCSParser(strings.NewReader(input))

	rcs, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	delta2 := rcs.Deltas["1.2"]
	if delta2 == nil {
		t.Fatal("Delta 1.2 not found")
	}
	if delta2.Log != "Added feature X" {
		t.Errorf("Log = %q, want %q", delta2.Log, "Added feature X")
	}
	if delta2.Text != "new content" {
		t.Errorf("Text = %q, want %q", delta2.Text, "new content")
	}

	delta1 := rcs.Deltas["1.1"]
	if delta1 == nil {
		t.Fatal("Delta 1.1 not found")
	}
	if delta1.Log != "Initial revision" {
		t.Errorf("Log = %q, want %q", delta1.Log, "Initial revision")
	}
	if delta1.Text != "initial content" {
		t.Errorf("Text = %q, want %q", delta1.Text, "initial content")
	}
}

func TestParseRCSDate(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Time
	}{
		{
			"2024.1.15.12.30.45",
			time.Date(2024, 1, 15, 12, 30, 45, 0, time.UTC),
		},
		{
			"2023.12.31.23.59.59",
			time.Date(2023, 12, 31, 23, 59, 59, 0, time.UTC),
		},
		{
			"2020.6.1.0.0.0",
			time.Date(2020, 6, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseRCSDate(tt.input)
			if !result.Equal(tt.expected) {
				t.Errorf("parseRCSDate(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseRCSDateInvalid(t *testing.T) {
	tests := []string{
		"invalid",
		"2024.1.15",             // too few parts
		"2024.1.15.12.30",       // too few parts
		"2024.1.15.12.30.45.60", // too many parts
		"",                      // empty
	}

	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			result := parseRCSDate(tt)
			if !result.IsZero() {
				t.Errorf("parseRCSDate(%q) = %v, want zero time", tt, result)
			}
		})
	}
}

func TestParserFullRCSFile(t *testing.T) {
	input := `head 1.5;
access;
symbols
	REL_1_0:1.4
	DEV:1.4.2.1;
locks; strict;
comment @# @;

1.5
date 2024.1.15.12.30.0; author johndoe; state Exp;
branches
	1.5.2.1;
next 1.4;

1.4
date 2024.1.10.10.0.0; author janedoe; state Exp;
branches
	1.4.2.1
	1.4.4.1;
next 1.3;

1.4.2.1
date 2024.1.12.14.0.0; author bob; state Exp;
next ;

1.3
date 2024.1.5.8.0.0; author alice; state Exp;
next ;

desc
@Sample RCS file for testing@

1.5
log @Added new feature@
text @new content@;

1.4
log @Fixed bug@
text @fixed content@;`

	parser := NewRCSParser(strings.NewReader(input))

	rcs, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Verify header
	if rcs.Head != "1.5" {
		t.Errorf("Head = %q, want %q", rcs.Head, "1.5")
	}

	if !rcs.StrictLocks {
		t.Error("StrictLocks should be true")
	}

	if rcs.Comment != "# " {
		t.Errorf("Comment = %q, want %q", rcs.Comment, "# ")
	}

	// Verify symbols
	if len(rcs.Symbols) != 2 {
		t.Errorf("Symbols length = %d, want 2", len(rcs.Symbols))
	}

	// Verify deltas
	if len(rcs.Deltas) != 4 {
		t.Errorf("Deltas length = %d, want 4", len(rcs.Deltas))
	}

	// Verify description
	if rcs.Description != "Sample RCS file for testing" {
		t.Errorf("Description = %q, want %q", rcs.Description, "Sample RCS file for testing")
	}
}

func TestParserSkipSemicolon(t *testing.T) {
	input := "head 1.5;;;" // Multiple semicolons
	parser := NewRCSParser(strings.NewReader(input))

	_, _ = parser.Parse()

	// Should not panic - skipSemicolon handles this gracefully
}

func TestParserUnknownHeaderField(t *testing.T) {
	input := "head 1.5; unknown_field somevalue; branch 1.5.1;"
	parser := NewRCSParser(strings.NewReader(input))

	rcs, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Should still parse known fields
	if rcs.Head != "1.5" {
		t.Errorf("Head = %q, want %q", rcs.Head, "1.5")
	}
}

func TestParserUnknownDeltaField(t *testing.T) {
	input := `head 1.5;
1.5
date 2024.1.15.12.30.0; author test; unknown_field value; state Exp;
desc @@;`

	parser := NewRCSParser(strings.NewReader(input))

	rcs, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	delta := rcs.Deltas["1.5"]
	if delta == nil {
		t.Fatal("Delta not found")
	}

	// Should still parse known fields
	if delta.Author != "test" {
		t.Errorf("Author = %q, want %q", delta.Author, "test")
	}
	if delta.State != "Exp" {
		t.Errorf("State = %q, want %q", delta.State, "Exp")
	}
}

func TestParserDeltaWithoutDate(t *testing.T) {
	input := `head 1.5;
1.5
author test; state Exp;
desc @@;`

	parser := NewRCSParser(strings.NewReader(input))

	rcs, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	delta := rcs.Deltas["1.5"]
	if delta == nil {
		t.Fatal("Delta not found")
	}

	if !delta.Date.IsZero() {
		t.Errorf("Date should be zero, got %v", delta.Date)
	}
}

func TestParserDeltaWithoutAuthor(t *testing.T) {
	input := `head 1.5;
1.5
date 2024.1.15.12.30.0; state Exp;
desc @@;`

	parser := NewRCSParser(strings.NewReader(input))

	rcs, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	delta := rcs.Deltas["1.5"]
	if delta == nil {
		t.Fatal("Delta not found")
	}

	if delta.Author != "" {
		t.Errorf("Author should be empty, got %q", delta.Author)
	}
}

func TestParserDeltaWithEmptyBranches(t *testing.T) {
	input := `head 1.5;
1.5
date 2024.1.15.12.30.0; author test; state Exp;
branches
;
next ;
desc @@;`

	parser := NewRCSParser(strings.NewReader(input))

	rcs, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	delta := rcs.Deltas["1.5"]
	if delta == nil {
		t.Fatal("Delta not found")
	}

	if len(delta.Branches) != 0 {
		t.Errorf("Branches should be empty, got %v", delta.Branches)
	}
}

func TestParserDeltaWithMultipleBranches(t *testing.T) {
	input := `head 1.5;
1.5
date 2024.1.15.12.30.0; author test; state Exp;
branches
	1.5.2.1
	1.5.4.1
	1.5.6.1;
next ;
desc @@;`

	parser := NewRCSParser(strings.NewReader(input))

	rcs, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	delta := rcs.Deltas["1.5"]
	if delta == nil {
		t.Fatal("Delta not found")
	}

	if len(delta.Branches) != 3 {
		t.Errorf("Branches length = %d, want 3", len(delta.Branches))
	}

	expected := []string{"1.5.2.1", "1.5.4.1", "1.5.6.1"}
	for i, exp := range expected {
		if i >= len(delta.Branches) {
			continue
		}
		if delta.Branches[i] != exp {
			t.Errorf("Branches[%d] = %q, want %q", i, delta.Branches[i], exp)
		}
	}
}

func TestParserDeltaTextWithoutLog(t *testing.T) {
	input := `head 1.5;
1.5
date 2024.1.15.12.30.0; author test; state Exp;
desc @@;
1.5
text @content@;`

	parser := NewRCSParser(strings.NewReader(input))

	rcs, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	delta := rcs.Deltas["1.5"]
	if delta == nil {
		t.Fatal("Delta not found")
	}

	if delta.Text != "content" {
		t.Errorf("Text = %q, want %q", delta.Text, "content")
	}
}

func TestParserDeltaTextWithUnknownField(t *testing.T) {
	input := `head 1.5;
1.5
date 2024.1.15.12.30.0; author test; state Exp;
desc @@;
1.5
log @message@;
unknown @value@;
text @content@;`

	parser := NewRCSParser(strings.NewReader(input))

	rcs, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	delta := rcs.Deltas["1.5"]
	if delta == nil {
		t.Fatal("Delta not found")
	}

	if delta.Log != "message" {
		t.Errorf("Log = %q, want %q", delta.Log, "message")
	}
	if delta.Text != "content" {
		t.Errorf("Text = %q, want %q", delta.Text, "content")
	}
}

func TestParserDeltaTextCreatesNewDelta(t *testing.T) {
	// Delta text section without corresponding delta metadata
	input := `head 1.5;
1.5
log @message@;
text @content@;
desc @@;`

	parser := NewRCSParser(strings.NewReader(input))

	rcs, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Delta should be created even without metadata
	delta := rcs.Deltas["1.5"]
	if delta == nil {
		t.Fatal("Delta should be created from delta text section")
	}
}

func TestParserAdvance(t *testing.T) {
	input := "head 1.5;"
	parser := NewRCSParser(strings.NewReader(input))

	// Initial token should be "head"
	if parser.token.Type != TokenIdent || parser.token.Value != "head" {
		t.Errorf("Initial token = %v %q, want Ident 'head'", parser.token.Type, parser.token.Value)
	}

	// Advance should move to next token
	parser.advance()
	if parser.token.Type != TokenNumber || parser.token.Value != "1.5" {
		t.Errorf("After advance, token = %v %q, want Number '1.5'", parser.token.Type, parser.token.Value)
	}
}
