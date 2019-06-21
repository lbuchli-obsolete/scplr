package parser

import (
	"github.com/lbuchli/scplr/lexer"
)

type SyntaxNode interface {
	Check(input string) (matchindex int, fullmatch bool)
}

type SyntaxParent struct {
	Children []SyntaxNode
}

func (sp SyntaxParent) Check(input string) (matchindex int, fullmatch bool) {
	for _, node := range sp.Children {
		mindex, fullmatch := node.Check(input[matchindex:])
		matchindex += mindex

		if !fullmatch {
			break
		}
	}

	return
}

type SyntaxRegex struct {
	Value lexer.Regex
}

func (sr SyntaxRegex) Check(input string) (matchindex int, fullmatch bool) {
	sr.Value.Match(input)
	return
}
