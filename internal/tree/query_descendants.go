package tree

func (s *Store) descendantsTreeLocked(rootID uint64, visited map[uint64]struct{}) DescendantsTree {
	if _, seen := visited[rootID]; seen {
		return DescendantsTree{ID: rootID, IsRef: true, Children: nil}
	}
	visited[rootID] = struct{}{}

	node := DescendantsTree{ID: rootID, Children: []DescendantsTree{}}

	childSet := s.children[rootID]
	for _, childID := range sortedChildIDs(childSet) {
		node.Children = append(node.Children, s.descendantsTreeLocked(childID, visited))
	}
	return node
}

func (s *Store) descendantsTreeMetaLocked(rootID uint64, visited map[uint64]struct{}) DescendantsTreeMeta {
	ev := s.events[rootID]

	if _, seen := visited[rootID]; seen {
		return DescendantsTreeMeta{
			ID:           rootID,
			TimeUnixNano: ev.TimeUnixNano,
			Type:         ev.Type,
			Message:      ev.Message,
			Payload:      ev.Payload,
			IsRef:        true,
		}
	}
	visited[rootID] = struct{}{}

	node := DescendantsTreeMeta{
		ID:           rootID,
		TimeUnixNano: ev.TimeUnixNano,
		Type:         ev.Type,
		Message:      ev.Message,
		Payload:      ev.Payload,
		IsRef:        false,
		Children:     []DescendantsTreeMeta{},
	}

	childSet := s.children[rootID]
	for _, childID := range sortedChildIDs(childSet) {
		node.Children = append(node.Children, s.descendantsTreeMetaLocked(childID, visited))
	}
	return node
}

func (s *Store) DescendantsTree(rootID uint64) (DescendantsTree, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.events[rootID]; !ok {
		return DescendantsTree{}, false
	}

	visited := make(map[uint64]struct{})
	return s.descendantsTreeLocked(rootID, visited), true
}

func (s *Store) DescendantsTreeMeta(rootID uint64) (DescendantsTreeMeta, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.events[rootID]; !ok {
		return DescendantsTreeMeta{}, false
	}

	visited := make(map[uint64]struct{})
	return s.descendantsTreeMetaLocked(rootID, visited), true
}

func (s *Store) DescendantsForest(rootIDs []uint64) ([]DescendantsTree, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.validateRootIDsLocked(rootIDs) {
		return nil, false
	}

	out := make([]DescendantsTree, 0, len(rootIDs))
	for _, id := range rootIDs {
		visited := make(map[uint64]struct{})
		out = append(out, s.descendantsTreeLocked(id, visited))
	}
	return out, true
}

func (s *Store) DescendantsForestMeta(rootIDs []uint64) ([]DescendantsTreeMeta, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.validateRootIDsLocked(rootIDs) {
		return nil, false
	}

	out := make([]DescendantsTreeMeta, 0, len(rootIDs))
	for _, id := range rootIDs {
		visited := make(map[uint64]struct{})
		out = append(out, s.descendantsTreeMetaLocked(id, visited))
	}
	return out, true
}
