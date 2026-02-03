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

func (s *Store) DescendantsTree(rootID uint64) (DescendantsTree, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.validateRootIDLocked(rootID)
	if err != nil {
		return DescendantsTree{}, err
	}

	visited := make(map[uint64]struct{})
	return s.descendantsTreeLocked(rootID, visited), nil
}

func (s *Store) DescendantsTreeMeta(rootID uint64) (DescendantsTreeMeta, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.validateRootIDLocked(rootID)
	if err != nil {
		return DescendantsTreeMeta{}, err
	}

	visited := make(map[uint64]struct{})
	return s.descendantsTreeMetaLocked(rootID, visited), nil
}

func (s *Store) DescendantsForest(rootIDs []uint64) ([]DescendantsTree, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.validateRootIDsLocked(rootIDs)
	if err != nil {
		return nil, err
	}

	out := make([]DescendantsTree, 0, len(rootIDs))
	for _, id := range rootIDs {
		visited := make(map[uint64]struct{})
		out = append(out, s.descendantsTreeLocked(id, visited))
	}
	return out, nil
}

func (s *Store) DescendantsForestMeta(rootIDs []uint64) ([]DescendantsTreeMeta, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.validateRootIDsLocked(rootIDs)
	if err != nil {
		return nil, err
	}

	out := make([]DescendantsTreeMeta, 0, len(rootIDs))
	for _, id := range rootIDs {
		visited := make(map[uint64]struct{})
		out = append(out, s.descendantsTreeMetaLocked(id, visited))
	}
	return out, nil
}
