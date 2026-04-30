package memory

import (
	"runtime"
	"time"

	"github.com/Mr-xiaotian/CelestialTree/internal/tree"
)

func (s *Store) Snapshot() tree.Snapshot {
	s.mu.Lock()
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
		TS:          time.Now().Unix(),
		GoRoutines:  runtime.NumGoroutine(),
		Edges:       edges,
		Roots:       roots,
		Heads:       heads,
		Subscribers: subscribers,
		NextEventID: nextEventID,
	}
}
