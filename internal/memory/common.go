package memory

import (
	"celestialtree/internal/tree"
	"slices"
)

func (s *Store) validateRootIDLocked(id uint64) error {
	if id == 0 {
		return &tree.RootIDError{
			ID:     id,
			Reason: "id must be non-zero",
		}
	}
	if _, ok := s.events[id]; !ok {
		return &tree.RootIDError{
			ID:     id,
			Reason: "event not found",
		}
	}

	return nil
}

func (s *Store) validateRootIDsLocked(rootIDs []uint64) error {
	for _, id := range rootIDs {
		err := s.validateRootIDLocked(id)
		if err != nil {
			return err
		}
	}
	return nil
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
