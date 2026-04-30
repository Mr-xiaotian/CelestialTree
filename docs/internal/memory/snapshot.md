# `snapshot.go`

## 文件整体描述

`snapshot.go` 是 **CelestialTree** 项目内存存储引擎中负责**运行时统计快照**的实现文件，位于 `internal/memory` 包中。该文件实现了 `Store.Snapshot` 方法，用于在不阻塞写入的前提下，快速采集内存 DAG 的核心指标（事件数、边数、Head 数、订阅者数、下一个事件 ID）。

该接口为 HTTP API 的 `/snapshot` 端点、监控系统、容量评估提供数据基础。

## 函数说明

### `(*Store) Snapshot`

```go
func (s *Store) Snapshot() tree.Snapshot
```

采集并返回内存存储的当前运行时统计信息。

**返回值**：`tree.Snapshot` —— 包含以下字段的结构体：

| 字段 | 类型 | 说明 |
|------|------|------|
| `TS` | `int64` | 快照采集时间的 Unix 时间戳（秒）。 |
| `GoRoutines` | `int` | 当前 goroutine 数量，通过 `runtime.NumGoroutine()` 获取。 |
| `Edges` | `int` | DAG 中的边总数。通过遍历 `s.children` 中所有子列表的长度累加得到。 |
| `Roots` | `int` | 当前 Root（无父节点的事件）数量，即 `len(s.roots)`。 |
| `Heads` | `int` | 当前 Head（无子节点的事件）数量，即 `len(s.heads)`。 |
| `Subscribers` | `int` | 当前活跃的 SSE 订阅者数量，即 `len(s.subs)`。 |
| `NextEventID` | `uint64` | 下一个将被分配的事件 ID，即 `s.nextID`。可用于推算系统中事件的大致规模。 |

**实现细节**：

1. **获取 DAG 统计**：
   - 加 `s.mu` 锁。
   - 拷贝 `len(s.roots)`、`len(s.heads)`、`s.nextID`。
   - 遍历 `s.children`，累加每个父节点对应的子列表长度得到 `edges`。
   - 释放 `s.mu` 锁。
2. **获取订阅统计**：
   - 加 `s.subsMu` 锁。
   - 拷贝 `len(s.subs)`。
   - 释放 `s.subsMu` 锁。
3. 添加 `runtime.NumGoroutine()` 和 `time.Now().Unix()` 到快照中。
4. 构造并返回 `tree.Snapshot`。

**设计要点**：

- **最小锁持有时间**：在 `s.mu` 锁内仅做简单的长度读取和整数累加，不做任何内存分配或复杂计算，因此对并发写入的阻塞时间极短。
- **双锁分离**：订阅者数量在独立的 `subsMu` 下读取，避免订阅操作与快照采集相互干扰。
- **非原子性**：`Roots`、`Heads`、`NextEventID` 的读取与 `Subscribers` 的读取不在同一个临界区内，因此返回的快照不是全局某一时刻的严格一致性视图。但对于监控与运维场景，这种”最终一致性”级别的统计完全可接受。

## 与其他文件的关系

| 依赖方向 | 文件/包 | 关系说明 |
|---------|--------|---------|
| 导入 | `internal/tree` | 返回 `tree.Snapshot` 类型。 |
| 同包协作 | `internal/memory/store.go` | 读取 `Store.children`、`Store.roots`、`Store.heads`、`Store.nextID`、`Store.subs`。 |
| 被调用 | `internal/httpapi/snapshot.go` | HTTP Handler 调用 `store.Snapshot()` 直接返回快照。 |

## 使用场景

- **监控大盘**：定时调用 `/snapshot` 采集 `edges`、`roots`、`heads` 指标，绘制系统增长曲线。
- **容量预警**：当 `next_event_id` 接近内存上限阈值时触发告警，提示需要扩容或归档。
- **调试诊断**：在排查问题时，通过 `subscribers` 判断 SSE 连接是否泄漏；通过 `goroutines` 判断是否有 goroutine 泄漏。
