# `sse.go`

## 文件整体描述

`sse.go` 是 **CelestialTree** 项目内存存储引擎中负责**服务端事件推送（Server-Sent Events）**的实现文件，位于 `internal/memory` 包中。该文件实现了订阅者管理与事件广播机制，为 HTTP API 的 `/subscribe` 端点提供底层支持。

核心职责包括：

- `Subscribe()` —— 注册新的 SSE 订阅者，返回订阅 ID、事件通道与取消函数。
- `broadcast()` —— 将新写入的事件推送给所有活跃订阅者。

## 函数说明

### `(*Store) Subscribe`

```go
func (s *Store) Subscribe() (subID uint64, ch <-chan tree.Event, cancel func())
```

注册一个新的 SSE 订阅者，建立从存储引擎到客户端的事件流通道。

**返回值**：

| 返回值 | 类型 | 说明 |
|-------|------|------|
| `subID` | `uint64` | 订阅者唯一标识，可用于调试或管理。 |
| `ch` | `<-chan tree.Event` | 只读事件通道，订阅者通过读取此通道接收新事件。通道缓冲大小为 64。 |
| `cancel` | `func()` | 取消函数，调用后将订阅者从 `s.subs` 中移除并关闭通道。 |

**实现细节**：

1. 获取 `s.subsMu` 锁。
2. `subID = atomic.AddUint64(&s.subSeq, 1)` —— 原子递增分配订阅 ID。
3. `c := make(chan tree.Event, 64)` —— 创建缓冲通道，缓冲大小 64 可短暂平滑写入突发。
4. `s.subs[subID] = c` —— 将通道注册到订阅者映射中。
5. 释放锁。
6. 构造 `cancel` 闭包：
   - 获取 `s.subsMu` 锁。
   - 若订阅者仍存在，从 `s.subs` 中删除并关闭通道。
   - 释放锁。
7. 返回 `subID`、只读通道引用 `c`、`cancel` 函数。

### `(*Store) broadcast`

```go
func (s *Store) broadcast(ev tree.Event)
```

将事件推送给所有当前活跃的 SSE 订阅者。该方法由 `Emit` 在事件写入成功后调用。

| 参数 | 类型 | 说明 |
|-----|------|------|
| `ev` | `tree.Event` | 待广播的新事件。 |

**实现细节**：

1. 获取 `s.subsMu` 锁。
2. 遍历 `s.subs` 中的所有通道。
3. 对每个通道使用 `select` 尝试发送：

```go
select {
case ch <- ev:
default:
    // v0：订阅者太慢就丢弃，保证 Emit 不被卡住
}
```

4. 释放锁。

**关键设计**：`default` 分支确保广播操作**不会阻塞**。若某订阅者的通道已满（消费速度慢于生产速度），事件将被静默丢弃。这是 v0 版本在“数据完整性”与“写入性能”之间的取舍。

## 与其他文件的关系

| 依赖方向 | 文件/包 | 关系说明 |
|---------|--------|---------|
| 导入 | `internal/tree` | 通道传输的数据类型为 `tree.Event`。 |
| 标准库 | `sync/atomic` | 使用 `atomic.AddUint64` 安全分配订阅 ID。 |
| 同包协作 | `internal/memory/store.go` | 操作 `Store.subs`、`Store.subSeq`，使用 `Store.subsMu`。 |
| 同包协作 | `internal/memory/emit.go` | `Emit` 在事件写入成功后调用 `broadcast(ev)` 触发推送。 |
| 被调用 | `internal/httpapi/sse.go` | HTTP Handler 调用 `store.Subscribe()` 注册订阅，并持有返回的通道与取消函数管理 SSE 连接生命周期。 |

## 设计说明

- **缓冲通道的容量选择**：缓冲大小 64 是一个经验值，可在不显著增加内存占用的前提下，短暂吸收写入突发（如短时间内连续 `Emit` 数十个事件）。若订阅者持续消费落后，缓冲将被填满，后续事件开始丢弃。
- **无历史回放**：`Subscribe()` 返回的通道只接收订阅**之后**产生的事件。新订阅者不会收到之前已写入的历史事件。这是 Go 通道的语义决定的，也是 SSE 在 CelestialTree 中的设计选择。需要历史数据的客户端应在订阅前先通过 `/event/`、`/descendants/` 等接口查询。
- **取消函数的资源安全**：`cancel` 函数内部先检查订阅者是否仍存在，再删除并关闭通道。这种“存在性检查”防止了重复调用 `cancel` 时对已关闭通道的二次 `close` 操作（Go 中对已关闭通道再次 `close` 会引发 panic）。
- **订阅者泄漏防护**：HTTP Handler 在 SSE 连接断开时必须调用 `cancel()`。当前 `internal/httpapi/sse.go` 的 `handleSubscribe` 使用 `defer cancel()` 确保即使发生 panic 或客户端异常断开，资源也会被释放。

## 扩展建议

- **可配置丢弃策略**：当前为硬编码的“丢弃”策略。未来可引入配置项，允许选择：
  - `drop`（默认，当前行为）
  - `block-with-timeout`（阻塞发送但带超时，超时后丢弃）
  - `expand-buffer`（动态扩展缓冲，内存允许时尽量不丢）
- **订阅者管理接口**：可增加管理端点（如 `/admin/subscribers`）暴露当前订阅者数量、每个订阅者的通道填充率、订阅时长等元数据，便于运维排查。
