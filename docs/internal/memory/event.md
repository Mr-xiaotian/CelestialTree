# `event.go`

## 文件整体描述

`event.go` 是 **CelestialTree** 项目内存存储引擎中负责**单事件精确查询**的实现文件，位于 `internal/memory` 包中。该文件仅包含一个极简的方法 `Get`，用于根据事件 ID 在内存索引中检索对应的 `tree.Event`。

虽然实现简单，但 `Get` 是存储层最基础的读取原语之一，被 HTTP API 的 `/event/{id}` 端点直接消费，也是其他更复杂查询（如溯源树、后代树）在内部递归时的基础操作。

## 函数说明

### `(*Store) Get`

```go
func (s *Store) Get(id uint64) (tree.Event, bool)
```

根据事件 ID 检索单个事件。

| 参数 | 类型 | 说明 |
|-----|------|------|
| `id` | `uint64` | 待查询的事件 ID。`0` 在系统中为无效 ID，不会命中任何记录。 |

**返回值**：

- `tree.Event`：查询到的事件实体。若未命中，返回零值 `tree.Event{}`。
- `bool`：`true` 表示事件存在；`false` 表示不存在。

**实现细节**：

1. 获取 `s.mu` 互斥锁（`Lock()` / `defer Unlock()`）。
2. 在 `s.events` 映射中查找 `id`。
3. 返回查找到的事件与是否存在标志。

**时间复杂度**：**O(1)** —— 纯哈希表查找。

## 与其他文件的关系

| 依赖方向 | 文件/包 | 关系说明 |
|---------|--------|---------|
| 导入 | `internal/tree` | 返回 `tree.Event` 类型。 |
| 同包协作 | `internal/memory/store.go` | 读取 `Store.events` 映射。 |
| 被调用 | `internal/httpapi/event.go` | HTTP Handler `/event/{id}` 直接调用 `store.Get(id)`，若不存在返回 `404`。 |
| 被间接使用 | `internal/memory/provenance.go` | 溯源树递归构建过程中通过 `s.events[rootID]` 直接访问事件（不经过 `Get` 方法，但在语义上等价）。 |

## 扩展建议

- **读锁优化**：若 `Store.mu` 升级为 `sync.RWMutex`，`Get` 方法应使用 `RLock()` / `RUnlock()`，允许多个读取操作并发执行。
- **缓存层**：当前纯内存存储本身已足够快（O(1) 哈希查找），无需额外缓存。若未来引入持久化后端（如 RocksDB、PostgreSQL），可在 `Store` 前增加 LRU 缓存，`Get` 作为缓存的首要入口。
