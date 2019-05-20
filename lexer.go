package main

import (
	"errors"
	"strings"

	"strconv"
)

type RegexType uint

const (
	GROUP = RegexType(iota)
	ANYOF
	RANGE
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
// It's a NFA without epsilon transitions
type DFA NFA

// NFA converts a regular expression to it's NFA
// equivalent.
func (r Regex) NFA() (nfa NFA, err error) {
	nfa = NFA{
		Transitions: [][][]rune{[][]rune{[]rune{}}},
	}

	symbols, err := symbols(r)
	if err != nil {
		return nfa, err
	}

	if len(symbols) == 1 {
		symbol := symbols[0]
		if symbol.Type != CHAR {
			return nfa, errors.New("Single non-char regex")
		}

		nfa.Out = []rune(symbol.Value)[0]

		return nfa, nil
	}

	a, err := Regex(symbols[0].Value).NFA()
	if err != nil {
		return nfa, err
	}

	b, err := Regex(symbols[1].Value).NFA()
	if err != nil {
		return nfa, err
	}

	second := symbols[1]
	switch second.Type {
	case CHAR, ANYOF, GROUP, RANGE:
		// a -> b ->
		nfa = a.Append(b)
	case OR:
		//      a
		// 0 -<   >- 0 ->
		//	    b
		nfa = a.Beside(b)
	case QUANTITY:
		if strings.Contains(second.Value, "-") { // Range
			parts := strings.Split(second.Value, "-")
			if len(parts) != 2 {
				return nfa, errors.New("Invalid range: " + second.Value)
			}
			// TODO
		} else if _, err := strconv.Atoi(second.Value); err == nil { // Single max value
			// TODO
		} else {
			switch second.Value {
			case "*":
				nfa = a.ZeroOrMany()
			case "?":
				nfa = a.ZeroOrOne()
			case "+":
				nfa = a.OneOrMany()
			default:
				return nfa, errors.New("Invalid quantity: " + second.Value)
			}
		}
	}

	return nfa, err
}

func symbols(r Regex) (symbols []RegexSymbol, err error) {
	regexRunes := []rune(r)
	symbols = []RegexSymbol{}
	for len(regexRunes) > 0 {
		switch regexRunes[0] {
		case '(': // Groups
			length, err := inParens(regexRunes)
			if err != nil {
				return symbols, err
			}

			symbols = append(symbols, RegexSymbol{
				Type:  GROUP,
				Value: string(regexRunes[1 : length+1]),
			})

			regexRunes = regexRunes[length:]
		case '[': // Any of chars
			length, err := inParens(regexRunes)
			if err != nil {
				return symbols, err
			}

			symbols = append(symbols, RegexSymbol{
				Type:  ANYOF,
				Value: string(regexRunes[1 : length+1]),
			})

			regexRunes = regexRunes[length:]
		case '.': // Any character
			symbols = append(symbols, RegexSymbol{
				Type:  ANYOF,
				Value: ".",
			})

			regexRunes = regexRunes[1:]
		case '\\': // Escape
			symbols = append(symbols, RegexSymbol{
				Type:  CHAR,
				Value: string(regexRunes[1]),
			})
			regexRunes = regexRunes[2:]
		case '{': // Quantity
			length, err := inParens(regexRunes)
			if err != nil {
				return symbols, err
			}

			symbols = append(symbols, RegexSymbol{
				Type:  QUANTITY,
				Value: string(regexRunes[1 : length+1]),
			})

			regexRunes = regexRunes[length:]
		case '+', '?', '*': // Quantity shorthands
			symbols = append(symbols, RegexSymbol{
				Type:  QUANTITY,
				Value: string(regexRunes[0]),
			})

			regexRunes = regexRunes[1:]
		case '|':
			symbols = append(symbols, RegexSymbol{
				Type:  OR,
				Value: "|",
			})

			regexRunes = regexRunes[1:]
		default:
			symbols = append(symbols, RegexSymbol{
				Type:  CHAR,
				Value: string(regexRunes[0]),
			})
			regexRunes = regexRunes[1:]
		}
	}

	return
}

func inParens(str []rune) (length int, err error) {
	// use a simple stack to check when the parenthesis are closed
	parens := []rune{str[0]}
	length = 1
	maxlength := len(str)
	for len(parens) > 0 {
		if length > maxlength {
			return length, errors.New("Unclosed parenthesis")
		}

		char := str[length]
		switch char {
		case '(', '[', '{':
			parens = append(parens, char)
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
	csize := asize + bsize
	c = newNFA(csize)

	// copy in NFA a
	c.paste(a, 0, 0)
	c.Transitions[asize][asize+1] = append(c.Transitions[asize][asize+1], a.Out)

	// copy in NFA b
	c.paste(b, asize, asize)
	c.Out = b.Out

	return c
}

// Beside puts NFA a and b next to each other in NFA c.
func (a NFA) Beside(b NFA) (c NFA) {
	// make an empty graph of the right size
	asize := len(a.Transitions)
	bsize := len(b.Transitions)
	csize := asize + bsize + 2
	c = newNFA(csize)

	// copy in NFA a
	c.paste(a, 1, 1)
	c.Transitions[asize+1][csize] = append(c.Transitions[asize+1][csize], a.Out)

	// copy in NFA b
	bLoc := 1 + asize
	c.paste(b, bLoc, bLoc)
	c.Transitions[bLoc+bsize][csize] = append(c.Transitions[bLoc][csize], a.Out)

	return c
}

func (a NFA) ZeroOrMany() (c NFA) {
	// make an empty graph of the right size
	asize := len(a.Transitions)
	csize := asize + 2
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
	csize := asize
	c = newNFA(csize)

	// copy in NFA a
	c.paste(a, 0, 0)

	// make transitions
	c.Transitions[asize-1][0] = append(c.Transitions[asize-1][0], '\x00')

	return c
}

func (a NFA) ZeroOrOne() (c NFA) {
	// make an empty graph of the right size
	asize := len(a.Transitions)
	csize := asize
	c = newNFA(csize)

	// copy in NFA a
	c.paste(a, 0, 0)

	// make transitions
	c.Transitions[0][asize-1] = append(c.Transitions[0][asize-1], '\x00')

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
	for cx := 0; cx < bsize; x++ {
		for cy := 0; cy < bsize; y++ {
			a.Transitions[cx+x][cy+y] = b.Transitions[x][y]
		}
	}
}
