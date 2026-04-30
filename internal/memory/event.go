package memory

import "github.com/Mr-xiaotian/CelestialTree/internal/tree"

// Get 根据 ID 获取单个事件，返回事件和是否存在。
func (s *Store) Get(id uint64) (tree.Event, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isEventIDValid(id) {
		return tree.Event{}, false
	}
	ev := s.events[id]
	return ev, true
}
