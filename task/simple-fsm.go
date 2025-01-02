package task

type FSM[S comparable, E comparable] struct {
	transitionListing   map[S][]S
	nextMapping         map[S]map[E]S
	mappingMissingState S
}

func (t FSM[S, E]) Contains(src, dst S) bool {
	for _, nxtState := range t.transitionListing[src] {
		if nxtState == dst {
			return true
		}
	}
	return false
}

func (t FSM[S, E]) ValidStateTransition(src S, dst S) bool {
	return t.Contains(src, dst)
}

func (t FSM[S, E]) Next(src S, event E) (S, E, S) {
	eventMapping, ok := t.nextMapping[src]
	if !ok {
		return src, event, t.mappingMissingState
	}

	nextState, ok := eventMapping[event]
	if !ok {
		return src, event, t.mappingMissingState
	}

	return src, event, nextState
}
