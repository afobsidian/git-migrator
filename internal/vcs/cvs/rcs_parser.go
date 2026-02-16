package cvs

import (
	"io"
	"strconv"
	"strings"
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
	// Format: YYYY.MM.DD.HH.MM.SS
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
		Deltas:  make(map[string]*Delta),
		Symbols: make(map[string]string),
		Locks:   make(map[string]string),
	}

	// Parse header
	p.parseHeader(rcs)

	// Parse deltas (revision metadata)
	p.parseDeltas(rcs)

	// Parse desc
	p.parseDesc(rcs)

	// Parse delta texts (log and text for each revision)
	p.parseDeltaTexts(rcs)

	return rcs, nil
}

// parseHeader parses the RCS header section
func (p *RCSParser) parseHeader(rcs *RCSFile) {
	for p.token.Type != TokenEOF {
		if p.token.Type != TokenIdent {
			break
		}

		switch p.token.Value {
		case "head":
			p.advance()
			if p.token.Type == TokenNumber {
				rcs.Head = p.token.Value
				p.advance()
			}
			p.skipSemicolon()

		case "branch":
			p.advance()
			if p.token.Type == TokenNumber {
				rcs.Branch = p.token.Value
				p.advance()
			}
			p.skipSemicolon()

		case "access":
			p.advance()
			for p.token.Type == TokenIdent {
				rcs.Access = append(rcs.Access, p.token.Value)
				p.advance()
			}
			p.skipSemicolon()

		case "symbols":
			p.advance()
			for p.token.Type == TokenIdent {
				sym := p.token.Value
				p.advance()
				if p.token.Type == TokenColon {
					p.advance()
					if p.token.Type == TokenNumber {
						rcs.Symbols[sym] = p.token.Value
						p.advance()
					}
				}
			}
			p.skipSemicolon()

		case "locks":
			p.advance()
			for p.token.Type == TokenIdent {
				lock := p.token.Value
				p.advance()
				if p.token.Type == TokenColon {
					p.advance()
					if p.token.Type == TokenNumber {
						rcs.Locks[lock] = p.token.Value
						p.advance()
					}
				}
			}
			p.skipSemicolon()

		case "strict":
			rcs.StrictLocks = true
			p.advance()
			p.skipSemicolon()

		case "comment":
			p.advance()
			if p.token.Type == TokenString {
				rcs.Comment = p.token.Value
				p.advance()
			}
			p.skipSemicolon()

		default:
			// Unknown field - could be start of deltas or desc
			// Don't consume the token, let outer loop handle it
			return
		}

		// Check if we've hit a revision number (start of deltas)
		if p.token.Type == TokenNumber {
			break
		}
	}
}

// skipSemicolon skips a semicolon if present
func (p *RCSParser) skipSemicolon() {
	if p.token.Type == TokenSemicolon {
		p.advance()
	}
}

// parseDeltas parses delta nodes (revision metadata)
func (p *RCSParser) parseDeltas(rcs *RCSFile) {
	for p.token.Type != TokenEOF {
		// Check for desc - end of deltas
		if p.token.Type == TokenIdent && p.token.Value == "desc" {
			break
		}

		// Must be a revision number
		if p.token.Type != TokenNumber {
			break
		}

		rev := p.token.Value
		p.advance()
		delta := &Delta{Revision: rev}

		// Parse delta fields until we hit another revision number or desc
		for p.token.Type != TokenEOF {
			// Check for end conditions
			if p.token.Type == TokenNumber {
				// Next revision
				break
			}
			if p.token.Type == TokenIdent && p.token.Value == "desc" {
				break
			}

			if p.token.Type == TokenIdent {
				switch p.token.Value {
				case "date":
					p.advance()
					if p.token.Type == TokenNumber {
						delta.Date = parseRCSDate(p.token.Value)
						p.advance()
					}
					p.skipSemicolon()

				case "author":
					p.advance()
					if p.token.Type == TokenIdent {
						delta.Author = p.token.Value
						p.advance()
					}
					p.skipSemicolon()

				case "state":
					p.advance()
					if p.token.Type == TokenIdent {
						delta.State = p.token.Value
						p.advance()
					}
					p.skipSemicolon()

				case "branches":
					p.advance()
					for p.token.Type == TokenNumber {
						delta.Branches = append(delta.Branches, p.token.Value)
						p.advance()
					}
					p.skipSemicolon()

				case "next":
					p.advance()
					if p.token.Type == TokenNumber {
						delta.Next = p.token.Value
						p.advance()
					}
					p.skipSemicolon()

				default:
					// Unknown field - skip it and its value
					p.advance()
					// Skip to semicolon
					for p.token.Type != TokenEOF && p.token.Type != TokenSemicolon {
						p.advance()
					}
					p.skipSemicolon()
				}
			} else {
				p.advance()
			}
		}

		rcs.Deltas[rev] = delta
		rcs.DeltaOrder = append(rcs.DeltaOrder, rev)
	}
}

// parseDesc parses the description
func (p *RCSParser) parseDesc(rcs *RCSFile) {
	if p.token.Type == TokenIdent && p.token.Value == "desc" {
		p.advance()
		if p.token.Type == TokenString {
			rcs.Description = p.token.Value
			p.advance()
		}
	}
}

// parseDeltaTexts parses log and text for each revision
func (p *RCSParser) parseDeltaTexts(rcs *RCSFile) {
	for p.token.Type != TokenEOF {
		// Revision number
		if p.token.Type != TokenNumber {
			// Not a revision number, skip this token and continue looking
			p.advance()
			continue
		}

		rev := p.token.Value
		p.advance()

		delta := rcs.Deltas[rev]
		if delta == nil {
			delta = &Delta{Revision: rev}
			rcs.Deltas[rev] = delta
		}

		// Parse log and text
		for p.token.Type != TokenEOF && p.token.Type != TokenNumber {
			if p.token.Type == TokenIdent {
				switch p.token.Value {
				case "log":
					p.advance()
					if p.token.Type == TokenString {
						delta.Log = p.token.Value
						p.advance()
					}

				case "text":
					p.advance()
					if p.token.Type == TokenString {
						delta.Text = p.token.Value
						p.advance()
					}

				default:
					p.advance()
				}
			} else {
				p.advance()
			}
		}
	}
}
