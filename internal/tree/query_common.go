package tree

import "slices"

func (s *Store) validateRootIDsLocked(rootIDs []uint64) bool {
	for _, id := range rootIDs {
		if id == 0 {
			return false
		}
		if _, ok := s.events[id]; !ok {
			return false
		}
	}
	return true
}

func sortedChildIDs(set map[uint64]struct{}) []uint64 {
	if len(set) == 0 {
		return []uint64{}
	}
	out := make([]uint64, 0, len(set))
	for id := range set {
		out = append(out, id)
	}
	slices.Sort(out)
	return out
}
