package memory

import "celestialtree/internal/tree"

func (s *Store) Get(id uint64) (tree.Event, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ev, ok := s.events[id]
	return ev, ok
}
