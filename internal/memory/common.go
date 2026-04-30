package memory

import (
	"slices"

	"github.com/Mr-xiaotian/CelestialTree/internal/tree"
)

// validateRootIDLocked 校验单个根 ID 是否合法（需在持锁状态调用）。
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

// validateRootIDsLocked 批量校验根 ID 列表（需在持锁状态调用）。
func (s *Store) validateRootIDsLocked(rootIDs []uint64) error {
	for _, id := range rootIDs {
		err := s.validateRootIDLocked(id)
		if err != nil {
			return err
		}
	}
	return nil
}

// isEventIDValid 检查 ID 对应的 events 槽位是否有效（非越界且非零值）。
func (s *Store) isEventIDValid(id uint64) bool {
	if id >= uint64(len(s.events)) || s.events[id].ID == 0 {
		return false
	}
	return true
}

// sortedChildIDs 返回 children 列表的排序副本，不修改原 slice。
func sortedChildIDs(sli []uint64) []uint64 {
	if len(sli) == 0 {
		return []uint64{}
	}
	out := make([]uint64, len(sli))
	copy(out, sli)
	slices.Sort(out)
	return out
}
