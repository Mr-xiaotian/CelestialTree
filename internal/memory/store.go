package memory

import (
	"celestialtree/internal/tree"
	"fmt"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Store 是 CelestialTree 的内存存储实现：
// - events:    id -> tree.Event
// - children:  parent -> set(child)
// - heads:     当前没有子节点的事件集合（叶子集合）
// - subs:      订阅者集合（用于 SSE 广播）
type Store struct {
	mu sync.Mutex // Maybe use RWMutex future

	nextID uint64

	events   map[uint64]tree.Event
	children map[uint64]map[uint64]struct{}
	heads    map[uint64]struct{}

	subsMu sync.Mutex
	subs   map[uint64]chan tree.Event
	subSeq uint64
}

func NewStore() *Store {
	return &Store{
		events:   make(map[uint64]tree.Event),
		children: make(map[uint64]map[uint64]struct{}),
		heads:    make(map[uint64]struct{}),
		subs:     make(map[uint64]chan tree.Event),
	}
}

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

func (s *Store) Get(id uint64) (tree.Event, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ev, ok := s.events[id]
	return ev, ok
}

func (s *Store) Children(id uint64) ([]uint64, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.events[id]; !ok {
		return nil, false
	}

	set := s.children[id]
	if set == nil {
		return []uint64{}, true
	}

	out := make([]uint64, 0, len(set))
	for c := range set {
		out = append(out, c)
	}
	return out, true
}

func (s *Store) Ancestors(id uint64) ([]uint64, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.events[id]; !ok {
		return nil, false
	}

	visited := make(map[uint64]struct{}, 64)
	roots := make(map[uint64]struct{}, 8)

	var dfs func(uint64) bool
	dfs = func(cur uint64) bool {
		if _, seen := visited[cur]; seen {
			return true
		}
		visited[cur] = struct{}{}

		ev, ok := s.events[cur]
		if !ok {
			return false
		}

		// root：没有 parents
		if len(ev.Parents) == 0 {
			roots[cur] = struct{}{}
			return true
		}

		for _, p := range ev.Parents {
			if _, ok := s.events[p]; !ok {
				return false
			}
			if !dfs(p) {
				return false
			}
		}
		return true
	}

	if !dfs(id) {
		return nil, false
	}

	out := make([]uint64, 0, len(roots))
	for rid := range roots {
		out = append(out, rid)
	}

	slices.Sort(out)

	return out, true
}

func (s *Store) Heads() []uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]uint64, 0, len(s.heads))
	for id := range s.heads {
		out = append(out, id)
	}
	return out
}
