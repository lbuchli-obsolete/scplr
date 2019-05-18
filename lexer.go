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
	symbols, err := symbols(r)
	if err != nil {
		return nfa, err
	}

	if () {}

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
