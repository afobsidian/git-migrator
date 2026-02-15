package main

import (
	"fmt"
	"strings"

	"github.com/adamf123git/git-migrator/internal/vcs/cvs"
)

func main() {
	input := `desc
@This is the file description.
It can span multiple lines.@`

	lexer := cvs.NewRCSLexer(strings.NewReader(input))

	for {
		token := lexer.NextToken()
		fmt.Printf("Token: Type=%v, Value=%q, Line=%d\n", token.Type, token.Value, token.Line)
		if token.Type == cvs.TokenEOF {
			break
		}
	}
}
