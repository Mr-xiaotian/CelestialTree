package memory

import "slices"

func (s *Store) Children(id uint64) ([]uint64, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.events[id]; !ok {
		return nil, false
	}

	set := s.children[id]
	if set == nil {
		return []uint64{}, true
	}

	out := make([]uint64, 0, len(set))
	for c := range set {
		out = append(out, c)
	}
	return out, true
}

func (s *Store) Ancestors(id uint64) ([]uint64, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.events[id]; !ok {
		return nil, false
	}

	visited := make(map[uint64]struct{}, 64)
	roots := make(map[uint64]struct{}, 8)

	var dfs func(uint64) bool
	dfs = func(cur uint64) bool {
		if _, seen := visited[cur]; seen {
			return true
		}
		visited[cur] = struct{}{}

		ev, ok := s.events[cur]
		if !ok {
			return false
		}

		// root：没有 parents
		if len(ev.Parents) == 0 {
			roots[cur] = struct{}{}
			return true
		}

		for _, p := range ev.Parents {
			if _, ok := s.events[p]; !ok {
				return false
			}
			if !dfs(p) {
				return false
			}
		}
		return true
	}

	if !dfs(id) {
		return nil, false
	}

	out := make([]uint64, 0, len(roots))
	for rid := range roots {
		out = append(out, rid)
	}

	slices.Sort(out)

	return out, true
}

func (s *Store) Heads() []uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]uint64, 0, len(s.heads))
	for id := range s.heads {
		out = append(out, id)
	}
	return out
}
