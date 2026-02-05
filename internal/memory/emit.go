package memory

import (
	"celestialtree/internal/tree"
	"fmt"
	"strings"
	"sync/atomic"
	"time"
)

// Emit 追加一个事件到 DAG 中。
func (s *Store) Emit(req tree.EmitRequest) (tree.Event, error) {
	if strings.TrimSpace(req.Type) == "" {
		return tree.Event{}, fmt.Errorf("type is required")
	}

	// parents 去重 + 过滤 0
	parents := make([]uint64, 0, len(req.Parents))
	seen := map[uint64]struct{}{}
	for _, p := range req.Parents {
		if p == 0 {
			continue
		}
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		parents = append(parents, p)
	}

	now := time.Now().UnixNano()
	id := atomic.AddUint64(&s.nextID, 1)

	ev := tree.Event{
		ID:           id,
		TimeUnixNano: now,
		Type:         req.Type,
		Message:      req.Message,
		Payload:      req.Payload,
		Parents:      parents,
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 父事件必须存在：否则历史图会断裂
	for _, p := range parents {
		if _, ok := s.events[p]; !ok {
			return tree.Event{}, fmt.Errorf("parent %d not found", p)
		}
	}

	// 写入事件
	s.events[id] = ev

	// 新事件默认是 head
	s.heads[id] = struct{}{}

	// 有 parents -> parents 不再是 head；同时建立 parent -> child 边
	for _, p := range parents {
		if s.children[p] == nil {
			s.children[p] = make(map[uint64]struct{})
		}
		s.children[p][id] = struct{}{}
		delete(s.heads, p)
	}

	// 广播给订阅者（非阻塞，慢订阅者可能丢事件：v0 的取舍）
	s.broadcast(ev)

	return ev, nil
}
