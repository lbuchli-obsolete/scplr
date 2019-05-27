package lexer

import "testing"

var mTestRegex = map[string]map[string]string{
	"a": map[string]string{
		"a": "a",
		"b": "",
	},
	"ab": map[string]string{
		"ab":  "ab",
		"abc": "ab",
		"ba":  "",
	},
	"a?": map[string]string{
		"a": "a",
		"":  "",
		"b": "",
	},
	"a+": map[string]string{
		"":   "",
		"a":  "a",
		"aa": "aa",
	},
	"a*": map[string]string{
		"":   "",
		"a":  "a",
		"aa": "aa",
	},
	"a|b": map[string]string{
		"a":  "a",
		"b":  "b",
		"ab": "a",
	},
}

func TestRegex(t *testing.T) {
	for regex, tests := range mTestRegex {
		dfa := FromRegex(regex)
		for str, match := range tests {
			_, result := dfa.Match(str)
			if result != match {
				t.Fatalf("Regex '%s' matched '%s' wrongly as '%s'. Correct would be: '%s'",
					regex, str, result, match)
			}
		}
	}
}
