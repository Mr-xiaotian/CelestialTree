# `emit.go`

## 文件整体描述

`emit.go` 是 **CelestialTree** 项目内存存储引擎中负责**事件写入**的核心实现文件，位于 `internal/memory` 包中。该文件实现了 `Store.Emit` 方法，是系统中**唯一的事件写入入口**。所有新事件（无论是通过 HTTP `/emit`、gRPC `Emit` 还是内部创世事件初始化）最终都会调用此方法，将事件追加到内存 DAG 中。

`Emit` 方法承担了参数校验、ID 分配、DAG 拓扑维护、索引更新与订阅广播的完整职责，是存储层最复杂、最关键的操作。

## 函数说明

### `(*Store) Emit`

```go
func (s *Store) Emit(req tree.EmitRequest) (tree.Event, error)
```

将一个新事件写入内存 DAG，并返回已创建的事件实体。

| 参数 | 类型 | 说明 |
|-----|------|------|
| `req` | `tree.EmitRequest` | 客户端请求，包含 `Type`、`Message`、`Payload`、`Parents`。 |

**返回值**：

- `tree.Event`：成功创建的事件，包含系统分配的 `ID`、生成时间戳 `TimeUnixNano` 以及处理后的 `Parents` 列表。
- `error`：写入失败时返回具体错误。

**处理流程**：

1. **Type 校验**：若 `req.Type` 为空或仅含空白字符，返回 `fmt.Errorf("type is required")`。
2. **Parents 预处理**：
   - 去重：通过 `map[uint64]struct{}` 剔除重复的父 ID。
   - 过滤 0：跳过值为 `0` 的父 ID（`0` 在系统中表示无效 ID）。
3. **ID 与时间戳分配**：
   - `id := atomic.AddUint64(&s.nextID, 1)` —— 原子递增，保证全局唯一且线程安全。
   - `now := time.Now().UnixNano()` —— 纳秒级时间戳。
4. **构造事件实体**：

```go
ev := tree.Event{
    ID:           id,
    TimeUnixNano: now,
    Type:         req.Type,
    Message:      req.Message,
    Payload:      req.Payload,
    Parents:      parents,
}
```

5. **加锁并写入 DAG**（`s.mu.Lock()`，手动 `s.mu.Unlock()`——不使用 `defer`，以便在锁外执行广播）：
   - **父事件存在性校验**：遍历所有 `parents`，若任一父 ID 通过 `isEventIDValid` 校验失败，先 `s.mu.Unlock()` 再返回 `fmt.Errorf("parent %d not found", p)`。此规则确保 DAG 不会断裂。
   - **扩展 events slice**：通过 `for uint64(len(s.events)) <= id` 循环追加零值 `tree.Event{}`，将稀疏 slice 扩展到足以容纳新 ID 的长度。
   - **写入事件**：`s.events[id] = ev`。
   - **更新 Head 集合**：新事件默认是 Head，`s.heads[id] = struct{}{}`。
   - **更新 Root 集合**：若 `parents` 为空，该事件为 Root，`s.roots[id] = struct{}{}`。
   - **更新父子关系索引**：遍历所有 `parents`：
     - `s.children[p] = append(s.children[p], id)` —— 将新事件 ID 追加到父事件的子 ID 列表。
     - 将父事件从 `s.heads` 中移除（因为它现在有了子事件，不再是叶子）。
   - **释放锁**：`s.mu.Unlock()`。
6. **广播订阅者**（锁外调用，`broadcast` 内部使用 `subsMu`）：
   - 调用 `s.broadcast(ev)`，将新事件推送给所有活跃的 SSE 订阅者。
   - 广播为非阻塞：慢消费者的通道若已满，事件会被丢弃。
   - 广播在 `mu` 释放后执行，避免广播期间阻塞其他读写操作。

## 与其他文件的关系

| 依赖方向 | 文件/包 | 关系说明 |
|---------|--------|---------|
| 导入 | `internal/tree` | 消费 `tree.EmitRequest`，生产 `tree.Event`。 |
| 同包协作 | `internal/memory/store.go` | 操作 `Store` 的 `events`、`children`、`roots`、`heads`、`nextID` 字段。 |
| 同包协作 | `internal/memory/sse.go` | 调用 `broadcast(ev)` 触发 SSE 推送。 |
| 被调用 | `internal/httpapi/emit.go` | HTTP Handler 将客户端请求转换为 `tree.EmitRequest` 后调用 `store.Emit`。 |
| 被调用 | `internal/grpcapi/emit.go` | gRPC Handler 将 `pb.EmitRequest` 转换为 `tree.EmitRequest` 后调用 `s.store.Emit`。 |
| 被调用 | `cmd/celestialtree/main.go` | 启动时调用 `store.Emit` 写入 Genesis 创世事件。 |

## 设计说明

- **父事件强制存在**：系统不允许“悬空事件”（即引用不存在父事件的事件）。这一设计保证了 DAG 的完整性，任何事件的血缘链都可以完整追溯。
- **无环保证的缺失与补偿**：当前 `Emit` 方法**不检测循环引用**（即 A -> B -> A）。这是因为循环需要在写入时检测所有祖先，时间复杂度较高。系统的假设是调用方（业务层）不会构造循环；若未来需要严格保证无环，可在 `Emit` 的父事件校验阶段增加一次向上的 BFS/DFS 检测。
- **Head 集合的实时维护**：`heads` 不是惰性计算的，而是在每次 `Emit` 时实时更新。这使得 `Heads()` 查询为 O(|heads|) 的简单遍历，无需遍历全图。
- **原子性**：从 ID 分配到 DAG 写入再到 Head/Root/Children 索引更新，全部在 `s.mu` 临界区内完成，保证了操作的原子性与一致性。
