package cvs

import (
	"io"
	"strings"
	"strconv"
	"time"
)

// RCSParser parses RCS file format
type RCSParser struct {
	lexer *RCSLexer
	token Token
}

// NewRCSParser creates a new RCS parser
func NewRCSParser(r io.Reader) *RCSParser {
	lexer := NewRCSLexer(r)
	return &RCSParser{
		lexer: lexer,
		token: lexer.NextToken(),
	}
}

func (p *RCSParser) advance() {
	p.token = p.lexer.NextToken()
}

func parseRCSDate(s string) time.Time {
	parts := strings.Split(s, ".")
	if len(parts) != 6 {
		return time.Time{}
	}
	year, _ := strconv.Atoi(parts[0])
	month, _ := strconv.Atoi(parts[1])
	day, _ := strconv.Atoi(parts[2])
	hour, _ := strconv.Atoi(parts[3])
	minute, _ := strconv.Atoi(parts[4])
	second, _ := strconv.Atoi(parts[5])

	return time.Date(year, time.Month(month), day, hour, minute, second, 0, time.UTC)
}

// Parse executes the main parsing logic
func (p *RCSParser) Parse() (*RCSFile, error) {
	rcs := &RCSFile{
		Deltas: make(map[string]*Delta),
	}

	if err := p.parseDeltas(rcs); err != nil {
		return nil, err
	}

	return rcs, nil
}

// parseDeltas processes delta and description data
func (p *RCSParser) parseDeltas(rcs *RCSFile) error {
	for p.token.Type != TokenEOF {
		if p.token.Type == TokenIdent && p.token.Value == "desc" {
			p.advance()
			if p.token.Type == TokenString {
				rcs.Description = p.token.Value
			}
			break
		}
		p.advance()
	}
	return nil
}