package main

import (
	"fmt"

	"github.com/lbuchli/scplr/lexer"
)

func main() {
	matched, strpart := lexer.FromRegex("a?").Match("a")
	fmt.Printf("Matched: %v\n", matched)
	fmt.Printf("Part:    '%s'\n", strpart)
}
