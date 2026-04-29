package memory

import "github.com/Mr-xiaotian/CelestialTree/internal/tree"

func (s *Store) Get(id uint64) (tree.Event, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isEventIDValid(id) {
		return tree.Event{}, false
	}
	ev := s.events[id]
	return ev, true
}
