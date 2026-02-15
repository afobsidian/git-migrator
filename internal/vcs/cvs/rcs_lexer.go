// Package cvs provides CVS repository reading and RCS file parsing capabilities.
package cvs

import (
	"bufio"
	"io"
)

// TokenType represents the type of a token in RCS format
type TokenType int

const (
	TokenEOF TokenType = iota
	TokenIdent    // identifier
	TokenNumber   // revision number like 1.2.3.4
	TokenString   // @-delimited string
	TokenSemicolon
	TokenColon
	TokenNewline
)

// Token represents a lexical token
type Token struct {
	Type  TokenType
	Value string
	Line  int
}

// RCSLexer tokenizes RCS file format
type RCSLexer struct {
	reader *bufio.Reader
	line   int
}

// NewRCSLexer creates a new RCS lexer
func NewRCSLexer(r io.Reader) *RCSLexer {
	return &RCSLexer{
		reader: bufio.NewReader(r),
		line:   1,
	}
}

// NextToken returns the next token from the input
func (l *RCSLexer) NextToken() Token {
	l.skipWhitespace()

	char, _, err := l.reader.ReadRune()
	if err != nil {
		return Token{Type: TokenEOF, Line: l.line}
	}

	switch char {
	case ';':
		return Token{Type: TokenSemicolon, Value: ";", Line: l.line}
	case ':':
		return Token{Type: TokenColon, Value: ":", Line: l.line}
	case '@':
		return l.readString()
	default:
		if isDigit(char) || (char == '.' && isDigit(l.peekChar())) {
			l.reader.UnreadRune()
			return l.readNumber()
		}
		if isAlpha(char) || char == '_' {
			l.reader.UnreadRune()
			return l.readIdent()
		}
		// Skip unknown characters
		return l.NextToken()
	}
}

func (l *RCSLexer) peekChar() rune {
	char, _, err := l.reader.ReadRune()
	if err != nil {
		return 0
	}
	l.reader.UnreadRune()
	return char
}

func (l *RCSLexer) skipWhitespace() {
	for {
		char, _, err := l.reader.ReadRune()
		if err != nil {
			return
		}
		if char == '\n' {
			l.line++
		}
		if !isWhitespace(char) && char != '\n' {
			l.reader.UnreadRune()
			return
		}
	}
}

func (l *RCSLexer) readString() Token {
	var result []rune

	for {
		char, _, err := l.reader.ReadRune()
		if err != nil {
			break
		}

		if char == '@' {
			// Check for escaped @@
			next, _, err := l.reader.ReadRune()
			if err != nil {
				break
			}
			if next == '@' {
				result = append(result, '@')
			} else {
				l.reader.UnreadRune()
				break
			}
		} else {
			if char == '\n' {
				l.line++
			}
			result = append(result, char)
		}
	}

	return Token{Type: TokenString, Value: string(result), Line: l.line}
}

func (l *RCSLexer) readNumber() Token {
	var result []rune

	for {
		char, _, err := l.reader.ReadRune()
		if err != nil {
			break
		}
		if isDigit(char) || char == '.' {
			result = append(result, char)
		} else {
			l.reader.UnreadRune()
			break
		}
	}

	return Token{Type: TokenNumber, Value: string(result), Line: l.line}
}

func (l *RCSLexer) readIdent() Token {
	var result []rune

	for {
		char, _, err := l.reader.ReadRune()
		if err != nil {
			break
		}
		if isAlpha(char) || isDigit(char) || char == '_' || char == '-' {
			result = append(result, char)
		} else {
			l.reader.UnreadRune()
			break
		}
	}

	return Token{Type: TokenIdent, Value: string(result), Line: l.line}
}

func isWhitespace(c rune) bool {
	return c == ' ' || c == '\t' || c == '\r'
}

func isAlpha(c rune) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isDigit(c rune) bool {
	return c >= '0' && c <= '9'
}
