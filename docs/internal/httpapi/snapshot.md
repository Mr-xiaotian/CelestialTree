# `snapshot.go`

## 文件整体描述

`snapshot.go` 是 **CelestialTree** 项目 HTTP API 中负责**运行时快照查询**的处理器文件，位于 `internal/httpapi` 包中。该文件实现了 `/snapshot` 端点，用于暴露内存存储引擎的当前运行状态统计信息，包括事件总量、边数量、Head 数量、SSE 订阅者数量以及下一个事件 ID。

该接口对监控、调试、容量评估非常有价值，可帮助运维人员在不深入日志的情况下快速了解系统负载与数据规模。

## 函数说明

### `handleSnapshot`

```go
func handleSnapshot(store *memory.Store) http.HandlerFunc
```

处理 `/snapshot` 端点，返回内存存储的实时统计快照。

| 参数 | 类型 | 说明 |
|-----|------|------|
| `store` | `*memory.Store` | 内存存储实例，通过闭包捕获。 |

**返回值**：`http.HandlerFunc` —— 标准 HTTP 处理函数。

**Handler 内部逻辑**：

1. **方法校验**：仅接受 `GET`。
2. **存储查询**：调用 `store.Snapshot()`，获取 `tree.Snapshot` 结构体（已包含时间戳、goroutine 数量等运行时信息）。
3. **响应**：返回 `200 OK`，响应体为 JSON 对象。

**响应示例**：

```json
{
  "ts": 1713709263,
  "goroutines": 7,
  "edges": 1480,
  "roots": 1,
  "heads": 43,
  "subscribers": 5,
  "next_event_id": 1524
}
```

**字段说明**：

| 字段 | 类型 | 来源 | 说明 |
|------|------|------|------|
| `ts` | `int64` | `time.Now().Unix()` | 快照生成时间戳（秒级）。 |
| `goroutines` | `int` | `runtime.NumGoroutine()` | 当前进程中的 goroutine 数量，反映并发负载。 |
| `edges` | `int` | `store.Snapshot().Edges` | DAG 中的边总数（即所有 parent->child 关系数）。 |
| `roots` | `int` | `store.Snapshot().Roots` | 当前 Root（无父节点的事件）数量。 |
| `heads` | `int` | `store.Snapshot().Heads` | 当前 Head（无子节点的事件）数量。 |
| `subscribers` | `int` | `store.Snapshot().Subscribers` | 当前活跃的 SSE 订阅者数量。 |
| `next_event_id` | `uint64` | `store.Snapshot().NextEventID` | 下一个将被分配的事件 ID，可用于推算事件规模。 |

## 与其他文件的关系

| 依赖方向 | 文件/包 | 关系说明 |
|---------|--------|---------|
| 导入 | `internal/memory` | 调用 `memory.Store.Snapshot()` 获取存储层统计信息。 |
| 同包协作 | `internal/httpapi/common.go` | 调用 `requireMethod`、`writeJSON`。 |
| 同包协作 | `internal/httpapi/routes.go` | `RegisterRoutes` 中将 `/snapshot` 注册到此 Handler。 |

## 设计说明

- **无侵入性**：`store.Snapshot()` 在内部使用最小粒度的锁策略（分别对事件锁与订阅锁加锁后快速拷贝计数），对并发写入的阻塞时间极短，因此高频调用 `/snapshot` 不会显著影响系统吞吐量。
- **聚合视角**：`edges` 字段的计数通过遍历 `children` 映射的所有子集汇总得出，虽然时间复杂度为 O(E)，但由于是在锁保护下对内存映射做简单计数，实际开销很小。
