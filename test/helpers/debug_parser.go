package main

import (
	"fmt"
	"strings"

	"github.com/adamf123git/git-migrator/internal/vcs/cvs"
)

func main() {
	input := `head 1.1;
desc
@This is the file description.
It can span multiple lines.@`

	parser := cvs.NewRCSParser(strings.NewReader(input))
	rcsFile, err := parser.Parse()
	if err != nil {
		fmt.Printf("Parse error: %v\n", err)
		return
	}

	fmt.Printf("Head: %s\n", rcsFile.Head)
	fmt.Printf("Description: %q\n", rcsFile.Description)
}
