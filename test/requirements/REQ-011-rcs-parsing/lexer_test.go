package requirements

import (
	"strings"
	"testing"

	"github.com/adamf123git/git-migrator/internal/vcs/cvs"
)

// TestRCSLexerTokens tests basic tokenization
func TestRCSLexerTokens(t *testing.T) {
	input := `head 1.3;`
	lexer := cvs.NewRCSLexer(strings.NewReader(input))

	tokens := []struct {
		expectedType cvs.TokenType
		expectedVal  string
	}{
		{cvs.TokenIdent, "head"},
		{cvs.TokenNumber, "1.3"},
		{cvs.TokenSemicolon, ";"},
		{cvs.TokenEOF, ""},
	}

	for i, expected := range tokens {
		token := lexer.NextToken()
		if token.Type != expected.expectedType {
			t.Errorf("Token %d: expected type %v, got %v", i, expected.expectedType, token.Type)
		}
		if token.Value != expected.expectedVal {
			t.Errorf("Token %d: expected value %q, got %q", i, expected.expectedVal, token.Value)
		}
	}
}

// TestRCSLexerString tests string tokenization
func TestRCSLexerString(t *testing.T) {
	input := `desc @This is a description@;`
	lexer := cvs.NewRCSLexer(strings.NewReader(input))

	// Skip 'desc'
	lexer.NextToken()

	// Get string token
	strToken := lexer.NextToken()
	if strToken.Type != cvs.TokenString {
		t.Errorf("Expected TokenString, got %v", strToken.Type)
	}
	if strToken.Value != "This is a description" {
		t.Errorf("Expected string value, got %q", strToken.Value)
	}
}

// TestRCSLexerNumbers tests revision number tokenization
func TestRCSLexerNumbers(t *testing.T) {
	input := `1.1 1.2.2.1 2.3.4.5.6`
	lexer := cvs.NewRCSLexer(strings.NewReader(input))

	revisions := []string{"1.1", "1.2.2.1", "2.3.4.5.6"}
	for i, expected := range revisions {
		token := lexer.NextToken()
		if token.Type != cvs.TokenNumber {
			t.Errorf("Token %d: expected TokenNumber, got %v", i, token.Type)
		}
		if token.Value != expected {
			t.Errorf("Token %d: expected %q, got %q", i, expected, token.Value)
		}
	}
}

// TestRCSLexerSymbols tests symbol tokenization
func TestRCSLexerSymbols(t *testing.T) {
	input := `symbols RELEASE_1_0:1.3 BETA:1.2;`
	lexer := cvs.NewRCSLexer(strings.NewReader(input))

	// symbols
	token := lexer.NextToken()
	if token.Type != cvs.TokenIdent || token.Value != "symbols" {
		t.Errorf("Expected 'symbols' identifier, got %v %q", token.Type, token.Value)
	}

	// RELEASE_1_0
	token = lexer.NextToken()
	if token.Type != cvs.TokenIdent || token.Value != "RELEASE_1_0" {
		t.Errorf("Expected 'RELEASE_1_0', got %v %q", token.Type, token.Value)
	}

	// :
	token = lexer.NextToken()
	if token.Type != cvs.TokenColon {
		t.Errorf("Expected colon, got %v", token.Type)
	}
}

// TestRCSLexerMultilineString tests multiline strings
func TestRCSLexerMultilineString(t *testing.T) {
	input := `desc @Line 1
Line 2
Line 3@`
	lexer := cvs.NewRCSLexer(strings.NewReader(input))

	// Skip 'desc'
	lexer.NextToken()

	token := lexer.NextToken()
	if token.Type != cvs.TokenString {
		t.Errorf("Expected TokenString, got %v", token.Type)
	}
	expected := "Line 1\nLine 2\nLine 3"
	if token.Value != expected {
		t.Errorf("Expected %q, got %q", expected, token.Value)
	}
}

// TestRCSLexerEscapedAt tests escaped @ symbols
func TestRCSLexerEscapedAt(t *testing.T) {
	input := `text @This has @@ symbol@`
	lexer := cvs.NewRCSLexer(strings.NewReader(input))

	// Skip 'text'
	lexer.NextToken()

	token := lexer.NextToken()
	if token.Type != cvs.TokenString {
		t.Errorf("Expected TokenString, got %v", token.Type)
	}
	expected := "This has @ symbol"
	if token.Value != expected {
		t.Errorf("Expected %q, got %q", expected, token.Value)
	}
}
