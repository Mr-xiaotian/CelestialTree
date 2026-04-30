# `store.go`

## 文件整体描述

`store.go` 是 **CelestialTree** 项目内存存储引擎的核心定义文件，位于 `internal/memory` 包中。该文件声明了 `Store` 结构体——整个系统的**单一事实来源（Single Source of Truth）**——以及其构造函数 `NewStore`。`Store` 以纯内存哈希表的形式维护事件 DAG 的全量数据，不提供持久化能力，适用于高吞吐、低延迟、可容忍重启丢失的场景（如事件总线、实时计算中间层）。

所有对 DAG 的写入、读取、遍历、订阅操作，最终都会落到 `Store` 的方法上。它是 HTTP API、gRPC API 与 SSE 推送三者的共同底层依赖。

## 实体说明

### `Store`

```go
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
```

内存存储引擎结构体，维护 DAG 的完整状态与订阅者集合。

| 字段 | 类型 | 说明 |
|------|------|------|
| `mu` | `sync.Mutex` | 保护 `events`、`children`、`roots`、`heads`、`nextID` 的互斥锁。注释提示未来可能升级为 `sync.RWMutex` 以提升读并发。 |
| `nextID` | `uint64` | 下一个待分配的事件 ID，通过 `atomic.AddUint64` 在 `Emit` 中安全递增。 |
| `events` | `[]tree.Event` | 事件主存储，使用稀疏 slice，以事件 ID 为下标直接寻址。空洞位置为零值 `tree.Event{}`（`ID == 0`）。相比 `map[uint64]tree.Event`，省去了 map 的 bucket 元数据开销，在大规模场景下（1M+ 事件）显著降低内存占用。 |
| `children` | `map[uint64][]uint64` | 父子关系索引，parent ID -> child ID 列表。相比之前的 `map[uint64]map[uint64]struct{}`，省去了每个内层 map 的 ~200 字节 header 开销。 |
| `roots` | `map[uint64]struct{}` | 当前所有无父事件（创世事件）的 ID 集合。 |
| `heads` | `map[uint64]struct{}` | 当前所有无子事件（叶子事件）的 ID 集合。新事件默认加入此集合；一旦有子事件产生，父事件即从集合中移除。 |
| `subsMu` | `sync.Mutex` | 保护订阅者映射 `subs` 与序列号 `subSeq` 的互斥锁。与 `mu` 分离，避免订阅/取消订阅操作阻塞事件写入。 |
| `subs` | `map[uint64]chan tree.Event` | 活跃 SSE 订阅者集合，sub ID -> 事件通道。 |
| `subSeq` | `uint64` | 订阅者 ID 序列号，通过 `atomic.AddUint64` 安全递增。 |

### `NewStore`

```go
func NewStore() *Store
```

`Store` 的构造函数，初始化所有内部映射与切片并返回就绪的存储实例。`events` 预分配 1024 容量。

**返回值**：`*Store` —— 所有映射与切片已初始化完成，可直接用于 `Emit`、`Get` 等操作。

**注意**：返回的 `Store` 中**不包含任何事件**。通常由 `cmd/celestialtree/main.go` 在启动后先调用 `store.Emit(tree.EmitRequest{Type: "genesis", ...})` 写入创世事件，再注册到 HTTP/gRPC 服务器。

## 与其他文件的关系

| 依赖方向 | 文件/包 | 关系说明 |
|---------|--------|---------|
| 导入 | `internal/tree` | 使用 `tree.Event` 作为核心存储单元。 |
| 被调用 | `cmd/celestialtree/main.go` | `main.go` 调用 `memory.NewStore()` 创建存储实例，注入到 HTTP 与 gRPC 服务器中。 |
| 被消费 | `internal/httpapi/*` | 所有 HTTP Handler 通过闭包持有 `*memory.Store`，调用其公开方法。 |
| 被消费 | `internal/grpcapi/*` | gRPC `Server` 持有 `*memory.Store`，将 RPC 请求委托给存储层。 |
| 同包协作 | `internal/memory/emit.go` | `Emit` 方法写入 `events`、`children`、`roots`、`heads`，并调用 `broadcast`。 |
| 同包协作 | `internal/memory/event.go` | `Get` 方法读取 `events`。 |
| 同包协作 | `internal/memory/graph.go` | `Children`、`Ancestors`、`Heads`、`Roots` 读取图关系索引。 |
| 同包协作 | `internal/memory/descendants.go` | `DescendantsTree*`、`DescendantsForest*` 基于 `children` 索引递归构建后代树。 |
| 同包协作 | `internal/memory/provenance.go` | `ProvenanceTree*`、`ProvenanceForest*` 基于 `events` 中的 `Parents` 递归构建溯源树。 |
| 同包协作 | `internal/memory/snapshot.go` | `Snapshot` 汇总所有映射的计数信息。 |
| 同包协作 | `internal/memory/sse.go` | `Subscribe`、`broadcast` 管理 `subs` 映射。 |
| 同包协作 | `internal/memory/common.go` | `validateRootIDLocked`、`validateRootIDsLocked`、`sortedChildIDs` 为 `Store` 的私有辅助方法。 |

## 并发模型

- **双锁分离**：`mu` 保护 DAG 数据，`subsMu` 保护订阅者集合。这种分离使得订阅/取消订阅操作不会阻塞事件写入，反之亦然。
- **ID 分配无锁化**：`nextID` 与 `subSeq` 使用 `atomic.AddUint64` 在锁外安全递增，减少锁持有时间。
- **广播非阻塞**：`broadcast` 在 `mu` 锁释放**之后**调用（持有 `subsMu` 期间仅做 `select` + `default` 尝试发送），不会阻塞在慢消费者上，也不会在广播期间阻塞其他读写操作。

## 扩展建议

- **持久化（WAL/Snapshot）**：当前纯内存设计在进程重启后数据全部丢失。若需持久化，可在 `Emit` 中追加 Write-Ahead Log，或在定时任务中将 `events`/`children` 快照写入磁盘/对象存储。
- **读写锁升级**：若读多写少场景明显，可将 `sync.Mutex` 替换为 `sync.RWMutex`，所有只读方法（`Get`、`Children`、`Heads` 等）使用 `RLock()`/`RUnlock()`。
