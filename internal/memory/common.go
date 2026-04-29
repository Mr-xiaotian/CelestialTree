package memory

import (
	"slices"

	"github.com/Mr-xiaotian/CelestialTree/internal/tree"
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

func sortedChildIDs(sli []uint64) []uint64 {
	if len(sli) == 0 {
		return []uint64{}
	}
	out := make([]uint64, len(sli))
	copy(out, sli)
	slices.Sort(out)
	return out
}
