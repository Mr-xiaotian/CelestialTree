package memory

import (
	"sync"

	"github.com/Mr-xiaotian/CelestialTree/internal/tree"
)

// Store 是 CelestialTree 的内存存储实现：
// - events:    id -> tree.Event
// - children:  parent -> set(child)
// - heads:     当前没有子节点的事件集合（叶子集合）
// - subs:      订阅者集合（用于 SSE 广播）
type Store struct {
	mu sync.Mutex // Maybe use RWMutex future

	nextID uint64

	events   []tree.Event
	children map[uint64][]uint64
	roots    map[uint64]struct{}
	heads    map[uint64]struct{}

	subsMu sync.Mutex
	subs   map[uint64]chan tree.Event
	subSeq uint64
}

// NewStore 创建并返回一个空的 Store 实例，events 预分配 1024 容量。
func NewStore() *Store {
	return &Store{
		events:   make([]tree.Event, 0, 1024),
		children: make(map[uint64][]uint64),
		roots:    make(map[uint64]struct{}),
		heads:    make(map[uint64]struct{}),
		subs:     make(map[uint64]chan tree.Event),
	}
}
