package main

import "testing"

var mregexToNFA = map[Regex]NFA{
	"": newNFA(0),
	"a": NFA{
		Transitions: [][][]rune{},
		Out:         'a',
	},
	"ab": NFA{
		Transitions: [][][]rune{
			[][]rune{[]rune{}, []rune{'a'}},
			[][]rune{[]rune{}, []rune{}},
		},
		Out: 'b',
	},
	"abc": NFA{
		Transitions: [][][]rune{
			[][]rune{[]rune{}, []rune{'a'}, []rune{}},
			[][]rune{[]rune{}, []rune{}, []rune{'b'}},
			[][]rune{[]rune{}, []rune{}, []rune{}},
		},
		Out: 'c',
	},
	"a|b": NFA{
		Transitions: [][][]rune{
			[][]rune{
				[]rune{}, []rune{'\x00'},
				[]rune{'\x00'}, []rune{},
			},
			[][]rune{
				[]rune{}, []rune{},
				[]rune{}, []rune{'a'},
			},
			[][]rune{
				[]rune{}, []rune{},
				[]rune{}, []rune{'b'},
			},
			[][]rune{
				[]rune{}, []rune{},
				[]rune{}, []rune{},
			},
		},
		Out: '\x00',
	},
	"a*": NFA{
		Transitions: [][][]rune{
			[][]rune{[]rune{}, []rune{}, []rune{'\x00'}},
			[][]rune{[]rune{}, []rune{}, []rune{'a'}},
			[][]rune{[]rune{}, []rune{'\x00'}, []rune{}},
		},
		Out: '\x00',
	},
	"a+": NFA{
		Transitions: [][][]rune{
			[][]rune{[]rune{}, []rune{'a'}},
			[][]rune{[]rune{'\x00'}, []rune{}},
		},
		Out: '\x00',
	},
	"a?": NFA{
		Transitions: [][][]rune{
			[][]rune{[]rune{}, []rune{'\x00'}, []rune{'\x00'}},
			[][]rune{[]rune{}, []rune{}, []rune{'a'}},
			[][]rune{[]rune{}, []rune{}, []rune{}},
		},
		Out: '\x00',
	},
}

var mnewNFA = map[int]NFA{
	0: NFA{
		Transitions: [][][]rune{},
		Out:         '\x00',
	},
	1: NFA{
		Transitions: [][][]rune{[][]rune{[]rune{}}},
		Out:         '\x00',
	},
}

func TestRegexToNFA(t *testing.T) {
	for regex, nfa := range mregexToNFA {
		result, err := regex.NFA()
		if err != nil {
			t.Fatalf("Error while calculating NFA of '%s': %v\n", regex, err)
		}

		if !testEq(result.Transitions, nfa.Transitions) || result.Out != nfa.Out {
			t.Fatalf("Unexpected result while calculating NFA of '%s':\nGot:\n%v\n\nExpected:\n%v\n",
				regex, prettySPrint(result), prettySPrint(nfa))
		}
	}
}

func TestNewNFA(t *testing.T) {
	for size, nfa := range mnewNFA {
		result := newNFA(size)

		if !testEq(result.Transitions, nfa.Transitions) || result.Out != nfa.Out {
			t.Fatalf("Unexpected result while creating empty NFA of size %d:\nGot: %v\nExpected: %v\n",
				size, result, nfa)
		}
	}
}

func testEq(a, b [][][]rune) bool {
	if len(a) != len(b) {
		return false
	}

	for x := 0; x < len(a); x++ {
		if len(a[x]) != len(b[x]) {
			return false
		}

		for y := 0; y < len(a[x]); y++ {
			if len(a[x][y]) != len(b[x][y]) {
				return false
			}

			for z := 0; z < len(a[x][y]); z++ {
				if a[x][y][z] != b[x][y][z] {
					return false
				}
			}
		}
	}

	return true
}

func prettySPrint(nfa NFA) string {
	result := ""
	for y := 0; y < len(nfa.Transitions); y++ {
		for x := 0; x < len(nfa.Transitions[y]); x++ {
			result += "'" + string(nfa.Transitions[x][y]) + "'\t"
		}
		result += "\n"
	}

	result += "-> " + string(nfa.Out)
	return result
}
