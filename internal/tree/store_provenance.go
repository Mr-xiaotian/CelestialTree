package tree

func (s *Store) provenanceTreeLocked(rootID uint64, visited map[uint64]struct{}) ProvenanceTree {
	if _, seen := visited[rootID]; seen {
		return ProvenanceTree{ID: rootID, IsRef: true, Parents: nil}
	}
	visited[rootID] = struct{}{}

	node := ProvenanceTree{ID: rootID, Parents: []ProvenanceTree{}}

	ev := s.events[rootID]
	for _, pid := range ev.Parents {
		if _, ok := s.events[pid]; !ok {
			continue
		}
		node.Parents = append(node.Parents, s.provenanceTreeLocked(pid, visited))
	}
	return node
}

func (s *Store) provenanceTreeMetaLocked(rootID uint64, visited map[uint64]struct{}) ProvenanceTreeMeta {
	ev := s.events[rootID]

	if _, seen := visited[rootID]; seen {
		return ProvenanceTreeMeta{
			ID:           rootID,
			TimeUnixNano: ev.TimeUnixNano,
			Type:         ev.Type,
			Message:      ev.Message,
			Payload:      ev.Payload,
			IsRef:        true,
		}
	}
	visited[rootID] = struct{}{}

	node := ProvenanceTreeMeta{
		ID:           rootID,
		TimeUnixNano: ev.TimeUnixNano,
		Type:         ev.Type,
		Message:      ev.Message,
		Payload:      ev.Payload,
		IsRef:        false,
		Parents:      []ProvenanceTreeMeta{},
	}

	for _, pid := range ev.Parents {
		if _, ok := s.events[pid]; !ok {
			continue
		}
		node.Parents = append(node.Parents, s.provenanceTreeMetaLocked(pid, visited))
	}
	return node
}

func (s *Store) ProvenanceTree(rootID uint64) (ProvenanceTree, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.validateRootIDLocked(rootID)
	if err != nil {
		return ProvenanceTree{}, err
	}

	visited := make(map[uint64]struct{})
	return s.provenanceTreeLocked(rootID, visited), nil
}

func (s *Store) ProvenanceTreeMeta(rootID uint64) (ProvenanceTreeMeta, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.validateRootIDLocked(rootID)
	if err != nil {
		return ProvenanceTreeMeta{}, err
	}

	visited := make(map[uint64]struct{})
	return s.provenanceTreeMetaLocked(rootID, visited), nil
}

func (s *Store) ProvenanceForest(rootIDs []uint64) ([]ProvenanceTree, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.validateRootIDsLocked(rootIDs)
	if err != nil {
		return nil, err
	}

	out := make([]ProvenanceTree, 0, len(rootIDs))
	for _, id := range rootIDs {
		visited := make(map[uint64]struct{})
		out = append(out, s.provenanceTreeLocked(id, visited))
	}
	return out, nil
}

func (s *Store) ProvenanceForestMeta(rootIDs []uint64) ([]ProvenanceTreeMeta, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.validateRootIDsLocked(rootIDs)
	if err != nil {
		return nil, err
	}

	out := make([]ProvenanceTreeMeta, 0, len(rootIDs))
	for _, id := range rootIDs {
		visited := make(map[uint64]struct{})
		out = append(out, s.provenanceTreeMetaLocked(id, visited))
	}
	return out, nil
}
