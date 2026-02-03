package memory

import (
	"celestialtree/internal/tree"
	"sync/atomic"
)

func (s *Store) Subscribe() (subID uint64, ch <-chan tree.Event, cancel func()) {
	s.subsMu.Lock()
	defer s.subsMu.Unlock()

	subID = atomic.AddUint64(&s.subSeq, 1)
	c := make(chan tree.Event, 64)
	s.subs[subID] = c

	cancel = func() {
		s.subsMu.Lock()
		defer s.subsMu.Unlock()
		if cc, ok := s.subs[subID]; ok {
			delete(s.subs, subID)
			close(cc)
		}
	}

	return subID, c, cancel
}

func (s *Store) broadcast(ev tree.Event) {
	s.subsMu.Lock()
	defer s.subsMu.Unlock()

	for _, ch := range s.subs {
		select {
		case ch <- ev:
		default:
			// v0：订阅者太慢就丢弃，保证 Emit 不被卡住
		}
	}
}
