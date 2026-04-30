package memory

import "github.com/Mr-xiaotian/CelestialTree/internal/tree"

// descendantsTreeLocked 递归构建以 rootID 为根的后代树（仅 ID），visited 用于环检测。
func (s *Store) descendantsTreeLocked(rootID uint64, visited map[uint64]struct{}) tree.DescendantsTree {
	if _, seen := visited[rootID]; seen {
		return tree.DescendantsTree{ID: rootID, IsRef: true, Children: nil}
	}
	visited[rootID] = struct{}{}

	node := tree.DescendantsTree{ID: rootID, Children: []tree.DescendantsTree{}}

	childSli := s.children[rootID]
	for _, childID := range sortedChildIDs(childSli) {
		node.Children = append(node.Children, s.descendantsTreeLocked(childID, visited))
	}
	return node
}

// descendantsTreeMetaLocked 递归构建以 rootID 为根的后代树（含元数据），visited 用于环检测。
func (s *Store) descendantsTreeMetaLocked(rootID uint64, visited map[uint64]struct{}) tree.DescendantsTreeMeta {
	ev := s.events[rootID]

	if _, seen := visited[rootID]; seen {
		return tree.DescendantsTreeMeta{
			ID:           rootID,
			TimeUnixNano: ev.TimeUnixNano,
			Type:         ev.Type,
			Message:      ev.Message,
			Payload:      ev.Payload,
			IsRef:        true,
		}
	}
	visited[rootID] = struct{}{}

	node := tree.DescendantsTreeMeta{
		ID:           rootID,
		TimeUnixNano: ev.TimeUnixNano,
		Type:         ev.Type,
		Message:      ev.Message,
		Payload:      ev.Payload,
		IsRef:        false,
		Children:     []tree.DescendantsTreeMeta{},
	}

	childSet := s.children[rootID]
	for _, childID := range sortedChildIDs(childSet) {
		node.Children = append(node.Children, s.descendantsTreeMetaLocked(childID, visited))
	}
	return node
}

// DescendantsTree 返回以 rootID 为根的后代树（仅包含 ID 和结构）。
func (s *Store) DescendantsTree(rootID uint64) (tree.DescendantsTree, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.validateRootIDLocked(rootID)
	if err != nil {
		return tree.DescendantsTree{}, err
	}

	visited := make(map[uint64]struct{})
	return s.descendantsTreeLocked(rootID, visited), nil
}

// DescendantsTreeMeta 返回以 rootID 为根的后代树（含事件元数据）。
func (s *Store) DescendantsTreeMeta(rootID uint64) (tree.DescendantsTreeMeta, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.validateRootIDLocked(rootID)
	if err != nil {
		return tree.DescendantsTreeMeta{}, err
	}

	visited := make(map[uint64]struct{})
	return s.descendantsTreeMetaLocked(rootID, visited), nil
}

// DescendantsForest 批量返回多个根节点的后代树（仅 ID）。
func (s *Store) DescendantsForest(rootIDs []uint64) ([]tree.DescendantsTree, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.validateRootIDsLocked(rootIDs)
	if err != nil {
		return nil, err
	}

	out := make([]tree.DescendantsTree, 0, len(rootIDs))
	for _, id := range rootIDs {
		visited := make(map[uint64]struct{})
		out = append(out, s.descendantsTreeLocked(id, visited))
	}
	return out, nil
}

// DescendantsForestMeta 批量返回多个根节点的后代树（含元数据）。
func (s *Store) DescendantsForestMeta(rootIDs []uint64) ([]tree.DescendantsTreeMeta, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.validateRootIDsLocked(rootIDs)
	if err != nil {
		return nil, err
	}

	out := make([]tree.DescendantsTreeMeta, 0, len(rootIDs))
	for _, id := range rootIDs {
		visited := make(map[uint64]struct{})
		out = append(out, s.descendantsTreeMetaLocked(id, visited))
	}
	return out, nil
}
