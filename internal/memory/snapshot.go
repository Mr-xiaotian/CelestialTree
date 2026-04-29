package memory

import (
	"github.com/Mr-xiaotian/CelestialTree/internal/tree"
)

func (s *Store) Snapshot() tree.Snapshot {
	s.mu.Lock()
	events := s.countEvents()
	roots := len(s.roots)
	heads := len(s.heads)
	nextEventID := s.nextID
	edges := 0
	for _, set := range s.children {
		edges += len(set)
	}
	s.mu.Unlock()

	s.subsMu.Lock()
	subscribers := len(s.subs)
	s.subsMu.Unlock()

	return tree.Snapshot{
		Events:      events,
		Edges:       edges,
		Roots:       roots,
		Heads:       heads,
		Subscribers: subscribers,
		NextEventID: nextEventID,
	}
}
