package lexer

type DFAState struct {
	ToThis    rune
	Next      []*DFAState
	Accepting bool
}

func FromRegex(regex string) *DFAState {
	runes := []rune(regex)
	index := 0
	length := len(runes)

	stack := []*DFAState{}

	// Anonymous stack functions
	push := func(state *DFAState) {
		stack = append(stack, state)
	}
	pop := func() *DFAState {
		last := len(stack) - 1
		val := stack[last]
		stack = stack[:last]
		return val
	}
	peak := func() *DFAState {
		return stack[len(stack)-1]
	}

	for index < length {
		char := runes[index]
		switch char {
		case '+':
			last := peak()
			next := newDFAState()
			next.ToThis = '\x00'
			next.link(last)
			push(next)
		case '*':
			last := pop()
			post := newDFAState()
			post.ToThis = '\x00'
			post.link(last)
			last.link(post)
			peak().link(post)
			push(last)
			push(post)
		case '?':
			last := pop()
			post := newDFAState()
			post.ToThis = '\x00'
			last.link(post)
			peak().link(post)
			push(last)
			push(post)
		case '(':
			// TODO
		case '|':
			//last := pop()
			// TODO
		case '\\':
			index++
			last := peak()
			next := newDFAState()
			next.ToThis = char
			last.link(next)
			push(next)
		default:
			if len(stack) > 0 {
				last := peak()
				next := newDFAState()
				next.ToThis = char
				last.link(next)
				push(next)
			} else {
				next := newDFAState()
				next.ToThis = char
				push(next)
			}
		}

		index++
	}

	peak().Accepting = true

	return stack[0]
}

func newDFAState() *DFAState {
	return &DFAState{
		Next: []*DFAState{},
	}
}

func (ds *DFAState) link(state *DFAState) {
	ds.Next = append(ds.Next, state)
}

func (ds *DFAState) Match(str string) (match bool, strpart string) {
	runes := []rune(str)
	current := ds
	index := 0
	length := len(runes)
	lastaccepted := -1
	for index < length {
		if current.Accepting {
			lastaccepted = index
		}

		for _, next := range current.Next {
			if next.ToThis == runes[index] { // labelled
				current = next
				index++
				break
			} else if next.ToThis == '\x00' { // unlabeled
				current = next
				// don't break; instead look for other possible labelled transitions first
			}
		}
	}

	if lastaccepted >= 0 {
		return true, string(runes[:lastaccepted])
	} else {
		return false, ""
	}
}
