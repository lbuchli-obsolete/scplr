package lexer

type DFAState struct {
	ToThis    rune
	Next      []*DFAState
	Accepting bool
}

type Regex DFAState

func FromRegex(regex string) Regex {
	runes := []rune(regex)
	index := 0
	length := len(runes)

	stack := []*DFAState{newDFAState()}

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
			last.link(next)
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
			last := peak()
			next := newDFAState()
			next.ToThis = char
			last.link(next)
			push(next)
		}

		index++
	}

	peak().Accepting = true

	return Regex(*stack[0])
}

func newDFAState() *DFAState {
	return &DFAState{
		Next: []*DFAState{},
	}
}

func (ds *DFAState) link(state *DFAState) {
	ds.Next = append(ds.Next, state)
}

func (r Regex) Match(str string) (match bool, strpart string) {
	runes := []rune(str)
	ds := DFAState(r)
	current := &ds
	index := 0
	length := len(runes)
	lastaccepted := -1
	nextIterationPossible := true
	for nextIterationPossible && index < length {
		nextIterationPossible = false
		if current.Accepting {
			lastaccepted = index
		}

		for _, next := range current.Next {
			if next.ToThis == runes[index] { // labelled
				current = next
				index++
				nextIterationPossible = true
				break
			} else if next.ToThis == '\x00' && next != current { // unlabeled
				current = next
				nextIterationPossible = true
				// don't break; instead look for other possible labelled transitions first
			}
		}
	}

	if lastaccepted >= 0 {
		return true, string(runes[:lastaccepted+1])
	} else {
		return false, ""
	}
}
