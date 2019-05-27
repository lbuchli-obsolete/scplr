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
			last := pop()

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

	previous.Accepting = true

	return result
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
		var next *DFAState

		if nextLabeled, ok := current.Next[runes[index]]; ok {
			next = nextLabeled
			index++
		} else if nextUnlabeled, ok := current.Next['\x00']; ok {
			next = nextUnlabeled
		} else {
			break
		}

		if next.Accepting {
			lastaccepted = index
		}

		current = next
	}

	if lastaccepted >= 0 {
		return true, string(runes[:lastaccepted])
	} else {
		return false, ""
	}
}
