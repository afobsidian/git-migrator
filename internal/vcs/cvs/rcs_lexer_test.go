package cvs

import (
	"strings"
	"testing"
)

func TestNewRCSLexer(t *testing.T) {
	input := "head 1.0;"
	lexer := NewRCSLexer(strings.NewReader(input))

	if lexer == nil {
		t.Fatal("NewRCSLexer returned nil")
	}

	if lexer.reader == nil {
		t.Error("reader should be initialized")
	}

	if lexer.line != 1 {
		t.Errorf("line = %d, want 1", lexer.line)
	}
}

func TestNewRCSLexerEmpty(t *testing.T) {
	lexer := NewRCSLexer(strings.NewReader(""))

	if lexer == nil {
		t.Fatal("NewRCSLexer returned nil")
	}
}

func TestLexerTokenEOF(t *testing.T) {
	lexer := NewRCSLexer(strings.NewReader(""))

	token := lexer.NextToken()
	if token.Type != TokenEOF {
		t.Errorf("token type = %v, want TokenEOF", token.Type)
	}
}

func TestLexerTokenSemicolon(t *testing.T) {
	lexer := NewRCSLexer(strings.NewReader(";"))

	token := lexer.NextToken()
	if token.Type != TokenSemicolon {
		t.Errorf("token type = %v, want TokenSemicolon", token.Type)
	}
	if token.Value != ";" {
		t.Errorf("token value = %q, want %q", token.Value, ";")
	}
}

func TestLexerTokenColon(t *testing.T) {
	lexer := NewRCSLexer(strings.NewReader(":"))

	token := lexer.NextToken()
	if token.Type != TokenColon {
		t.Errorf("token type = %v, want TokenColon", token.Type)
	}
	if token.Value != ":" {
		t.Errorf("token value = %q, want %q", token.Value, ":")
	}
}

func TestLexerSimpleString(t *testing.T) {
	lexer := NewRCSLexer(strings.NewReader("@hello world@"))

	token := lexer.NextToken()
	if token.Type != TokenString {
		t.Errorf("token type = %v, want TokenString", token.Type)
	}
	if token.Value != "hello world" {
		t.Errorf("token value = %q, want %q", token.Value, "hello world")
	}
}

func TestLexerEmptyString(t *testing.T) {
	lexer := NewRCSLexer(strings.NewReader("@@"))

	token := lexer.NextToken()
	if token.Type != TokenString {
		t.Errorf("token type = %v, want TokenString", token.Type)
	}
	if token.Value != "" {
		t.Errorf("token value = %q, want empty", token.Value)
	}
}

func TestLexerStringWithNewline(t *testing.T) {
	lexer := NewRCSLexer(strings.NewReader("@line1\nline2@"))

	token := lexer.NextToken()
	if token.Type != TokenString {
		t.Errorf("token type = %v, want TokenString", token.Type)
	}
	if token.Value != "line1\nline2" {
		t.Errorf("token value = %q, want %q", token.Value, "line1\nline2")
	}
}

func TestLexerStringWithEscapedAt(t *testing.T) {
	lexer := NewRCSLexer(strings.NewReader("@hello@@world@"))

	token := lexer.NextToken()
	if token.Type != TokenString {
		t.Errorf("token type = %v, want TokenString", token.Type)
	}
	if token.Value != "hello@world" {
		t.Errorf("token value = %q, want %q", token.Value, "hello@world")
	}
}

func TestLexerStringWithMultipleEscapedAt(t *testing.T) {
	lexer := NewRCSLexer(strings.NewReader("@a@@b@@c@"))

	token := lexer.NextToken()
	if token.Type != TokenString {
		t.Errorf("token type = %v, want TokenString", token.Type)
	}
	if token.Value != "a@b@c" {
		t.Errorf("token value = %q, want %q", token.Value, "a@b@c")
	}
}

func TestLexerNumberSimple(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1", "1"},
		{"1.0", "1.0"},
		{"1.2.3.4", "1.2.3.4"},
		{"1.2.0.2", "1.2.0.2"},
		{"2024.1.15.12.30.45", "2024.1.15.12.30.45"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			lexer := NewRCSLexer(strings.NewReader(tt.input))
			token := lexer.NextToken()
			if token.Type != TokenNumber {
				t.Errorf("token type = %v, want TokenNumber", token.Type)
			}
			if token.Value != tt.expected {
				t.Errorf("token value = %q, want %q", token.Value, tt.expected)
			}
		})
	}
}

func TestLexerNumberTrailingDot(t *testing.T) {
	// Number followed by something else
	lexer := NewRCSLexer(strings.NewReader("1.2;"))

	token := lexer.NextToken()
	if token.Type != TokenNumber {
		t.Errorf("token type = %v, want TokenNumber", token.Type)
	}
	if token.Value != "1.2" {
		t.Errorf("token value = %q, want %q", token.Value, "1.2")
	}
}

func TestLexerIdentSimple(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"head", "head"},
		{"branch", "branch"},
		{"author", "author"},
		{"date", "date"},
		{"HEAD", "HEAD"},
		{"MyTag", "MyTag"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			lexer := NewRCSLexer(strings.NewReader(tt.input))
			token := lexer.NextToken()
			if token.Type != TokenIdent {
				t.Errorf("token type = %v, want TokenIdent", token.Type)
			}
			if token.Value != tt.expected {
				t.Errorf("token value = %q, want %q", token.Value, tt.expected)
			}
		})
	}
}

func TestLexerIdentWithUnderscore(t *testing.T) {
	lexer := NewRCSLexer(strings.NewReader("my_tag_name"))

	token := lexer.NextToken()
	if token.Type != TokenIdent {
		t.Errorf("token type = %v, want TokenIdent", token.Type)
	}
	if token.Value != "my_tag_name" {
		t.Errorf("token value = %q, want %q", token.Value, "my_tag_name")
	}
}

func TestLexerIdentWithHyphen(t *testing.T) {
	lexer := NewRCSLexer(strings.NewReader("tag-name"))

	token := lexer.NextToken()
	if token.Type != TokenIdent {
		t.Errorf("token type = %v, want TokenIdent", token.Type)
	}
	if token.Value != "tag-name" {
		t.Errorf("token value = %q, want %q", token.Value, "tag-name")
	}
}

func TestLexerIdentWithNumbers(t *testing.T) {
	lexer := NewRCSLexer(strings.NewReader("tag123"))

	token := lexer.NextToken()
	if token.Type != TokenIdent {
		t.Errorf("token type = %v, want TokenIdent", token.Type)
	}
	if token.Value != "tag123" {
		t.Errorf("token value = %q, want %q", token.Value, "tag123")
	}
}

func TestLexerSkipWhitespace(t *testing.T) {
	lexer := NewRCSLexer(strings.NewReader("   head   "))

	token := lexer.NextToken()
	if token.Type != TokenIdent {
		t.Errorf("token type = %v, want TokenIdent", token.Type)
	}
	if token.Value != "head" {
		t.Errorf("token value = %q, want %q", token.Value, "head")
	}
}

func TestLexerSkipTabs(t *testing.T) {
	lexer := NewRCSLexer(strings.NewReader("\t\thead\t\t"))

	token := lexer.NextToken()
	if token.Type != TokenIdent {
		t.Errorf("token type = %v, want TokenIdent", token.Type)
	}
	if token.Value != "head" {
		t.Errorf("token value = %q, want %q", token.Value, "head")
	}
}

func TestLexerSkipCarriageReturn(t *testing.T) {
	lexer := NewRCSLexer(strings.NewReader("\r\rhead\r\r"))

	token := lexer.NextToken()
	if token.Type != TokenIdent {
		t.Errorf("token type = %v, want TokenIdent", token.Type)
	}
	if token.Value != "head" {
		t.Errorf("token value = %q, want %q", token.Value, "head")
	}
}

func TestLexerLineCounting(t *testing.T) {
	lexer := NewRCSLexer(strings.NewReader("head\nbranch\n1.0"))

	token := lexer.NextToken()
	if token.Line != 1 {
		t.Errorf("first token line = %d, want 1", token.Line)
	}

	token = lexer.NextToken()
	if token.Line != 2 {
		t.Errorf("second token line = %d, want 2", token.Line)
	}

	token = lexer.NextToken()
	if token.Line != 3 {
		t.Errorf("third token line = %d, want 3", token.Line)
	}
}

func TestLexerLineCountingInString(t *testing.T) {
	lexer := NewRCSLexer(strings.NewReader("@line1\nline2\nline3@"))

	token := lexer.NextToken()
	if token.Type != TokenString {
		t.Errorf("token type = %v, want TokenString", token.Type)
	}

	// After the string, line should be 3
	token = lexer.NextToken()
	if token.Line != 3 {
		t.Errorf("after string, line = %d, want 3", token.Line)
	}
}

func TestLexerMultipleTokens(t *testing.T) {
	input := "head 1.0;"
	lexer := NewRCSLexer(strings.NewReader(input))

	// head
	token := lexer.NextToken()
	if token.Type != TokenIdent {
		t.Errorf("token 1 type = %v, want TokenIdent", token.Type)
	}
	if token.Value != "head" {
		t.Errorf("token 1 value = %q, want %q", token.Value, "head")
	}

	// 1.0
	token = lexer.NextToken()
	if token.Type != TokenNumber {
		t.Errorf("token 2 type = %v, want TokenNumber", token.Type)
	}
	if token.Value != "1.0" {
		t.Errorf("token 2 value = %q, want %q", token.Value, "1.0")
	}

	// ;
	token = lexer.NextToken()
	if token.Type != TokenSemicolon {
		t.Errorf("token 3 type = %v, want TokenSemicolon", token.Type)
	}

	// EOF
	token = lexer.NextToken()
	if token.Type != TokenEOF {
		t.Errorf("token 4 type = %v, want TokenEOF", token.Type)
	}
}

func TestLexerSymbols(t *testing.T) {
	input := "symbols REL_1_0:1.0 DEV:1.1.0.2;"
	lexer := NewRCSLexer(strings.NewReader(input))

	expected := []struct {
		tokenType TokenType
		value     string
	}{
		{TokenIdent, "symbols"},
		{TokenIdent, "REL_1_0"},
		{TokenColon, ":"},
		{TokenNumber, "1.0"},
		{TokenIdent, "DEV"},
		{TokenColon, ":"},
		{TokenNumber, "1.1.0.2"},
		{TokenSemicolon, ";"},
		{TokenEOF, ""},
	}

	for i, exp := range expected {
		token := lexer.NextToken()
		if token.Type != exp.tokenType {
			t.Errorf("token %d type = %v, want %v", i, token.Type, exp.tokenType)
		}
		if token.Value != exp.value {
			t.Errorf("token %d value = %q, want %q", i, token.Value, exp.value)
		}
	}
}

func TestLexerDeltaEntry(t *testing.T) {
	input := "1.1\ndate 2024.1.15.12.30.0; author johndoe; state Exp;"
	lexer := NewRCSLexer(strings.NewReader(input))

	expected := []struct {
		tokenType TokenType
		value     string
	}{
		{TokenNumber, "1.1"},
		{TokenIdent, "date"},
		{TokenNumber, "2024.1.15.12.30.0"},
		{TokenSemicolon, ";"},
		{TokenIdent, "author"},
		{TokenIdent, "johndoe"},
		{TokenSemicolon, ";"},
		{TokenIdent, "state"},
		{TokenIdent, "Exp"},
		{TokenSemicolon, ";"},
		{TokenEOF, ""},
	}

	for i, exp := range expected {
		token := lexer.NextToken()
		if token.Type != exp.tokenType {
			t.Errorf("token %d type = %v, want %v (value=%q)", i, token.Type, exp.tokenType, token.Value)
		}
		if token.Value != exp.value {
			t.Errorf("token %d value = %q, want %q", i, token.Value, exp.value)
		}
	}
}

func TestLexerUnknownCharacters(t *testing.T) {
	// Unknown characters should be skipped
	lexer := NewRCSLexer(strings.NewReader("head!1.0"))

	token := lexer.NextToken()
	if token.Type != TokenIdent {
		t.Errorf("token type = %v, want TokenIdent", token.Type)
	}
	if token.Value != "head" {
		t.Errorf("token value = %q, want %q", token.Value, "head")
	}

	// ! is skipped, next is number
	token = lexer.NextToken()
	if token.Type != TokenNumber {
		t.Errorf("token type = %v, want TokenNumber", token.Type)
	}
}

func TestLexerPeekChar(t *testing.T) {
	lexer := NewRCSLexer(strings.NewReader("head"))

	// Peek should return 'h' without consuming
	char := lexer.peekChar()
	if char != 'h' {
		t.Errorf("peekChar = %q, want %q", char, 'h')
	}

	// Next token should still work
	token := lexer.NextToken()
	if token.Value != "head" {
		t.Errorf("token value = %q, want %q", token.Value, "head")
	}
}

func TestLexerPeekCharEmpty(t *testing.T) {
	lexer := NewRCSLexer(strings.NewReader(""))

	char := lexer.peekChar()
	if char != 0 {
		t.Errorf("peekChar on empty = %q, want 0", char)
	}
}

func TestLexerStringWithSpecialChars(t *testing.T) {
	lexer := NewRCSLexer(strings.NewReader("@tab\there\nand@"))

	token := lexer.NextToken()
	if token.Type != TokenString {
		t.Errorf("token type = %v, want TokenString", token.Type)
	}
	if token.Value != "tab\there\nand" {
		t.Errorf("token value = %q, want %q", token.Value, "tab\there\nand")
	}
}

func TestLexerStringWithUnicode(t *testing.T) {
	lexer := NewRCSLexer(strings.NewReader("@hello 世界@"))

	token := lexer.NextToken()
	if token.Type != TokenString {
		t.Errorf("token type = %v, want TokenString", token.Type)
	}
	if token.Value != "hello 世界" {
		t.Errorf("token value = %q, want %q", token.Value, "hello 世界")
	}
}

func TestLexerHelperFunctions(t *testing.T) {
	// Test isWhitespace
	if !isWhitespace(' ') {
		t.Error("space should be whitespace")
	}
	if !isWhitespace('\t') {
		t.Error("tab should be whitespace")
	}
	if !isWhitespace('\r') {
		t.Error("carriage return should be whitespace")
	}
	if isWhitespace('a') {
		t.Error("'a' should not be whitespace")
	}
	if isWhitespace('\n') {
		t.Error("newline should not be whitespace (handled separately)")
	}

	// Test isAlpha
	if !isAlpha('a') {
		t.Error("'a' should be alpha")
	}
	if !isAlpha('Z') {
		t.Error("'Z' should be alpha")
	}
	if isAlpha('1') {
		t.Error("'1' should not be alpha")
	}

	// Test isDigit
	if !isDigit('0') {
		t.Error("'0' should be digit")
	}
	if !isDigit('9') {
		t.Error("'9' should be digit")
	}
	if isDigit('a') {
		t.Error("'a' should not be digit")
	}
}

func TestLexerMultipleEOF(t *testing.T) {
	lexer := NewRCSLexer(strings.NewReader("head"))

	_ = lexer.NextToken() // head
	_ = lexer.NextToken() // EOF

	// Multiple EOF calls should be safe
	for i := 0; i < 3; i++ {
		token := lexer.NextToken()
		if token.Type != TokenEOF {
			t.Errorf("EOF call %d: type = %v, want TokenEOF", i, token.Type)
		}
	}
}

func TestLexerComplexRCSFragment(t *testing.T) {
	input := `head 1.5;
access;
symbols
	REL_1_0:1.4
	DEV:1.4.0.2;
locks; strict;
comment @# @;

1.5
date 2024.1.15.12.30.0; author johndoe; state Exp;
branches 1.4.0.2;
next 1.4;

desc
@@Initial revision@@
`

	lexer := NewRCSLexer(strings.NewReader(input))

	// Just verify we can tokenize the whole thing without errors
	tokens := 0
	for {
		token := lexer.NextToken()
		if token.Type == TokenEOF {
			break
		}
		tokens++
	}

	if tokens == 0 {
		t.Error("Expected to tokenize some tokens")
	}
}

func TestLexerStringStartingWithAt(t *testing.T) {
	// @@@@start@ parses as:
	// - First @ starts string
	// - Second @ is escaped by third @, so we get @ in result
	// - Fourth @ ends string (next char is 's', not @)
	// Result: string with value "@"
	lexer := NewRCSLexer(strings.NewReader("@@@@start@"))

	token := lexer.NextToken()
	if token.Type != TokenString {
		t.Errorf("token type = %v, want TokenString", token.Type)
	}
	// The value is "@" because @@ is an escaped @ character
	if token.Value != "@" {
		t.Errorf("token value = %q, want %q", token.Value, "@")
	}
}

func TestLexerMixedContent(t *testing.T) {
	input := "head\t1.0;\nbranch 1.0.1;\n@log\nmessage@ 1.1"
	lexer := NewRCSLexer(strings.NewReader(input))

	expected := []TokenType{
		TokenIdent,  // head
		TokenNumber, // 1.0
		TokenSemicolon,
		TokenIdent,  // branch
		TokenNumber, // 1.0.1
		TokenSemicolon,
		TokenString, // log\nmessage
		TokenNumber, // 1.1
		TokenEOF,
	}

	for i, expType := range expected {
		token := lexer.NextToken()
		if token.Type != expType {
			t.Errorf("token %d type = %v, want %v (value=%q)", i, token.Type, expType, token.Value)
		}
	}
}
