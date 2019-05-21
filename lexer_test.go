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
			[][]rune{
				[]rune{},
				[]rune{'a'},
			},
			[][]rune{
				[]rune{},
				[]rune{},
			},
		},
		Out: 'b',
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
			[][]rune{[]rune{'\x00'}},
		},
		Out: 'a',
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
			t.Fatalf("Unexpected result while calculating NFA of '%s':\nGot: %v\nExpected: %v\n",
				regex, result, nfa)
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
