package main

import (
	"errors"
)

type RegexType uint

const (
	GROUP = RegexType(iota)
	ANYOF
	QUANTITY
	OR
	CHAR
)

type RegexSymbol struct {
	Type  RegexType
	Value string
}

// A Regex is a (somewhat) human-readable NFA.
type Regex string

// NFA stands for Nondeterministic Finite Automata
// The NULL byte is used as epsilon (\x00)
// The starting state is 0. It also has a
// labeled out-transition from the last state (len(NFA)).
type NFA struct {
	Transitions [][][]rune
	Out         rune
}

// DFA stands for Deterministic Finite Automata
// It's a NFA without epsilon transitions and no
// identically labelled transitions out of the same
// state.
type DFA struct {
	Transitions [][][]rune
	Accepting   []int
}

// NFA converts a regular expression to it's NFA
// equivalent.
func (r Regex) NFA() (nfa NFA, err error) {

	if len(r) == 1 {
		return NFA{
			Transitions: [][][]rune{},
			Out:         []rune(r)[0],
		}, nil
	}

	stack := []NFA{}

	// anonymous stack functions
	push := func(nfa NFA) {
		stack = append(stack, nfa)
	}
	pop := func() NFA {
		last := len(stack) - 1
		val := stack[last]
		stack = stack[:last]
		return val
	}

	var lastWasOr bool
	var sym RegexSymbol

	for len(r) > 0 {
		sym, r, err = r.nextSymbol()
		if err != nil {
			return nfa, err
		}

		switch sym.Type {
		case QUANTITY:
			switch sym.Value {
			case "+":
				push(pop().OneOrMany())
			case "*":
				push(pop().ZeroOrMany())
			case "?":
				push(pop().ZeroOrOne())
			default:
				//TODO
			}
		case ANYOF:
		case OR:
			// let the next symbol handle the 'beside' composition
			lastWasOr = true
		default:
			nfa, err := Regex(sym.Value).NFA()
			if err != nil {
				return nfa, err
			}

			// if the last symbol was an or, put this symbol beside
			// the last one
			if lastWasOr {
				push(pop().Beside(nfa))
				lastWasOr = false
			} else {
				push(nfa)
			}
		}
	}

	// Compose all loose items
	nfa = newNFA(0)
	for _, subnfa := range stack {
		nfa = nfa.Append(subnfa)
	}

	return
}

func (r Regex) nextSymbol() (symbol RegexSymbol, cut Regex, err error) {
	regexRunes := []rune(r)
	if len(regexRunes) > 0 {
		switch regexRunes[0] {
		case '(': // Groups
			length, err := inParens(regexRunes)
			if err != nil {
				return symbol, r, err
			}

			return RegexSymbol{
				Type:  GROUP,
				Value: string(regexRunes[1 : length+1]),
			}, r[length+1:], nil
		case '[': // Any of chars
			length, err := inParens(regexRunes)
			if err != nil {
				return symbol, r, err
			}

			return RegexSymbol{
				Type:  ANYOF,
				Value: string(regexRunes[1 : length+1]),
			}, r[length+1:], nil
		case '.': // Any character
			return RegexSymbol{
				Type:  ANYOF,
				Value: ".",
			}, r[1:], nil
		case '\\': // Escape
			return RegexSymbol{
				Type:  CHAR,
				Value: string(regexRunes[1]),
			}, r[2:], nil
		case '{': // Quantity
			length, err := inParens(regexRunes)
			if err != nil {
				return symbol, r, err
			}

			return RegexSymbol{
				Type:  QUANTITY,
				Value: string(regexRunes[1 : length+1]),
			}, r[length:], nil

		case '+', '?', '*': // Quantity shorthands
			return RegexSymbol{
				Type:  QUANTITY,
				Value: string(regexRunes[0]),
			}, r[1:], nil
		case '|':
			return RegexSymbol{
				Type:  OR,
				Value: "|",
			}, r[1:], nil
		default:
			return RegexSymbol{
				Type:  CHAR,
				Value: string(regexRunes[0]),
			}, r[1:], nil
		}
	}

	return symbol, r, errors.New("Requested next symbol when there was no symbol in regex")
}

func inParens(str []rune) (length int, err error) {
	// use a simple stack to check when the parenthesis are closed
	// start with the closing version of the parenthesis
	parens := []rune{str[0] + 1}
	length = 1
	maxlength := len(str)
	for len(parens) > 0 {
		if length > maxlength {
			return length, errors.New("Unclosed parenthesis")
		}

		char := str[length]
		switch char {
		case '(', '[', '{':
			// append the closing version of the parenthesis
			parens = append(parens, char+1)
		case ')', ']', '}':
			if parens[len(parens)-1] == char {
				parens = parens[:len(parens)-1]
			} else {
				return length, errors.New("Parenthesis do not match")
			}
		}
		length++
	}

	return length - 2, nil
}

// Append appends NFA b to NFA a.
func (a NFA) Append(b NFA) (c NFA) {
	// make an empty graph of the right size
	asize := len(a.Transitions)
	bsize := len(b.Transitions)
	csize := asize + bsize + 2
	c = newNFA(csize)

	// copy in NFA a
	c.paste(a, 0, 0)
	c.Transitions[asize][asize+1] = append(c.Transitions[asize][asize+1], a.Out)

	// copy in NFA b
	c.paste(b, asize+1, asize+1)
	c.Out = b.Out

	return c
}

// Beside puts NFA a and b next to each other in NFA c.
func (a NFA) Beside(b NFA) (c NFA) {
	// make an empty graph of the right size
	asize := len(a.Transitions)
	bsize := len(b.Transitions)
	csize := asize + bsize + 4
	c = newNFA(csize)

	bLoc := 2 + asize
	c.Transitions[0][1] = []rune{'\x00'}
	c.Transitions[0][bLoc] = []rune{'\x00'}

	// copy in NFA a
	c.paste(a, 1, 1)
	c.Transitions[asize+1][csize-1] = append(c.Transitions[asize+1][csize-1], a.Out)

	// copy in NFA b
	c.paste(b, bLoc, bLoc)
	c.Transitions[bLoc+bsize][csize-1] = append(c.Transitions[bLoc+bsize][csize-1], b.Out)

	return c
}

func (a NFA) ZeroOrMany() (c NFA) {
	// make an empty graph of the right size
	asize := len(a.Transitions)
	csize := asize + 3
	c = newNFA(csize)

	// copy in NFA a
	c.paste(a, 1, 1)

	// make transitions
	c.Transitions[0][asize+2] = []rune{'\x00'}
	c.Transitions[asize+1][asize+2] = []rune{a.Out}
	c.Transitions[asize+2][1] = []rune{'\x00'}

	return c
}

func (a NFA) OneOrMany() (c NFA) {
	// make an empty graph of the right size
	asize := len(a.Transitions)
	csize := asize + 2
	c = newNFA(csize)

	// copy in NFA a
	c.paste(a, 0, 0)

	// make transitions
	c.Transitions[asize][asize+1] = append(c.Transitions[asize][0], a.Out)
	c.Transitions[asize+1][0] = []rune{'\x00'}

	return c
}

func (a NFA) ZeroOrOne() (c NFA) {
	// make an empty graph of the right size
	asize := len(a.Transitions)
	csize := asize + 1
	c = newNFA(csize)

	// copy in NFA a
	c.paste(a, 0, 0)
	c.Out = a.Out

	// make transitions
	c.Transitions[0][asize] = append(c.Transitions[0][asize], '\x00')

	return c
}

func newNFA(size int) NFA {
	result := NFA{
		Transitions: make([][][]rune, size),
		Out:         '\x00',
	}
	for i, _ := range result.Transitions {
		result.Transitions[i] = make([][]rune, size)
	}

	return result
}

// paste pastes an NFA into another. It does not check for
// out of bounds errors and will panic if one occurs.
// It also does not handle Out transitions.
func (a *NFA) paste(b NFA, x, y int) {
	bsize := len(b.Transitions)
	for bx := 0; bx < bsize; bx++ {
		for by := 0; by < bsize; by++ {
			a.Transitions[bx+x][by+y] = b.Transitions[bx][by]
		}
	}
}
