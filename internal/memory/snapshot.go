package memory

import (
	"github.com/Mr-xiaotian/CelestialTree/internal/tree"
)

func (s *Store) Snapshot() tree.Snapshot {
	s.mu.Lock()
	events := len(s.events)
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
		Heads:       heads,
		Subscribers: subscribers,
		NextEventID: nextEventID,
	}
}
