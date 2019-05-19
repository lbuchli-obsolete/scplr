package main

import "errors"

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

	switch symbols[1].Type {
	case CHAR, ANYOF, GROUP, RANGE:
		// a -> b ->
		nfa = a.Append(b)
	case OR:
		//      a
		// 0 -<   >- 0 ->
		//	    b
		// TODO
	case QUANTITY:
		// TODO
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

// Append appends a NFA 'b' to a NFA 'a'.
// a -> b ->
func (a NFA) Append(b NFA) (c NFA) {
	// make an empty graph of the right size
	asize := len(a.Transitions)
	bsize := len(b.Transitions)
	csize := asize + bsize
	c = NFA{
		Transitions: make([][][]rune, csize),
		Out:         '\x00',
	}
	for i, _ := range c.Transitions {
		c.Transitions[i] = make([][]rune, csize)
	}

	// copy in NFA a
	for x := 0; x < asize; x++ {
		for y := 0; y < asize; y++ {
			c.Transitions[x][y] = a.Transitions[x][y]
		}
	}
	// set out transition
	c.Transitions[asize][asize+1] = []rune{a.Out}

	// copy in NFA b
	for x := 0; x < bsize; x++ {
		for y := 0; y < bsize; y++ {
			c.Transitions[x+asize][y+asize] = append(c.Transitions[x+asize][y+asize],
				b.Transitions[x][y]...)
		}
	}
	// set out transition
	c.Out = b.Out

	return c
}

func (a NFA) Beside(b NFA) (c NFA) {

}
