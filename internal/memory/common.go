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
	if !s.isEventIDValid(id) {
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

func (s *Store) isEventIDValid(id uint64) bool {
	if id >= uint64(len(s.events)) || s.events[id].ID == 0 {
		return false
	}
	return true
}

func (s *Store) countEvents() int {
	var count int = 0
	for _, ev := range s.events {
		if ev.ID != 0 {
			count++
		}
	}
	return count
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
