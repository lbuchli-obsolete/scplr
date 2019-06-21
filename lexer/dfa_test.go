package lexer

import (
	"testing"
)

func TestRegexPlain(t *testing.T) {
	dfa := FromRegex("abc")
	matched, strpart := dfa.Match("abc")
	if strpart != "abc" || !matched {
		t.Fatalf("Regex 'abc' matched 'abc' wrongly as '%s'. Correct would be: 'abc'", strpart)
	}
}

func TestRegexOneOrMany(t *testing.T) {
	dfa := FromRegex("a+")
	matched, strpart := dfa.Match("aaaaa")
	if strpart != "aaaaa" || !matched {
		t.Fatalf("Regex 'a+' matched 'aaaaa' wrongly as '%s'. Correct would be: 'aaaaa'", strpart)
	}
}

func TestRegexZeroOrOne(t *testing.T) {
	dfa := FromRegex("a?")
	matched, strpart := dfa.Match("")
	if strpart != "" || !matched {
		t.Fatalf("Regex 'a?' matched '' wrongly as '%s'. Correct would be: ''", strpart)
	}
}

func TestRegexZeroOrMany(t *testing.T) {
	dfa := FromRegex("a*")
	matched, strpart := dfa.Match("aaaaaaa")
	if strpart != "aaaaaaa" || !matched {
		t.Fatalf("Regex 'a*' matched 'aaaaaaa' wrongly as '%s'. Correct would be: 'aaaaaaa'", strpart)
	}
}
