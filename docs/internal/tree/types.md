# `types.go`

## 文件整体描述

`types.go` 是 **CelestialTree** 项目的核心数据模型定义文件，位于 `internal/tree` 包中。该文件承担整个系统的“契约层”职责，定义了所有跨模块共享的数据结构、请求/响应体以及错误类型。任何对事件（Event）的读写、对 DAG 树的遍历、对外的 API 交互，都会依赖此文件中声明的类型。

由于 `internal/tree` 不依赖项目内部其他业务包（只依赖标准库 `encoding/json` 和 `fmt`），它是整个依赖图中最底层、最稳定的模块，适合作为后续演化的基础契约。

## 数据结构与类型说明

### `Event`

```go
type Event struct {
    ID           uint64          `json:"id"`
    TimeUnixNano int64           `json:"time_unix_nano"`
    Type         string          `json:"type"`
    Message      string          `json:"message,omitempty"`
    Payload      json.RawMessage `json:"payload,omitempty"`
    Parents      []uint64        `json:"parents"`
}
```

CelestialTree 的**最小历史原子**。每一条 `Event` 代表 DAG 中的一个节点，包含：

- `ID`: 全局唯一递增标识，由 `memory.Store` 在写入时分配。
- `TimeUnixNano`: 事件生成时间戳（纳秒级 Unix 时间）。
- `Type`: 事件类型，必填，用于业务分类与过滤。
- `Message`: 可选的人类可读文本描述。
- `Payload`: 可选的 JSON 载荷，使用 `json.RawMessage` 延迟解析，提升性能与灵活性。
- `Parents`: 父事件 ID 列表，用于构建 DAG 的边关系；空列表表示该事件为**创世根节点（Root）**。

### `EmitRequest`

```go
type EmitRequest struct {
    Type    string          `json:"type"`
    Message string          `json:"message,omitempty"`
    Payload json.RawMessage `json:"payload,omitempty"`
    Parents []uint64        `json:"parents"`
}
```

客户端请求**写入事件**时的请求体结构。被 HTTP API (`/emit`) 和 gRPC API (`Emit`) 共同消费，并最终传入 `memory.Store.Emit`。

### `TreeBatchRequest`

```go
type TreeBatchRequest struct {
    IDs  []uint64 `json:"ids"`
    View string   `json:"view,omitempty"`
}
```

批量查询后代树（descendants）或溯源树（provenance）时的请求体。`View` 字段用于控制返回结构：空值/`"struct"` 返回树形结构，`"meta"` 返回带完整元数据的树形结构。

### `EmitResponse`

```go
type EmitResponse struct {
    ID uint64 `json:"id"`
}
```

`/emit` 接口成功后返回的响应体，仅包含新创建事件的 ID。

### `DescendantsTree`

```go
type DescendantsTree struct {
    ID       uint64            `json:"id"`
    IsRef    bool              `json:"is_ref"`
    Children []DescendantsTree `json:"children"`
}
```

描述某个事件及其**所有后代**的树形结构，自顶向下展开。`IsRef` 为 `true` 表示该节点在当前树中因环检测或重复访问而以引用形式出现，其 `Children` 不再展开（避免循环引用导致的无限递归）。

### `DescendantsTreeMeta`

```go
type DescendantsTreeMeta struct {
    ID           uint64                `json:"id"`
    TimeUnixNano int64                 `json:"time_unix_nano"`
    Type         string                `json:"type"`
    IsRef        bool                  `json:"is_ref"`
    Message      string                `json:"message,omitempty"`
    Payload      json.RawMessage       `json:"payload,omitempty"`
    Children     []DescendantsTreeMeta `json:"children"`
}
```

与 `DescendantsTree` 结构相同，但额外携带事件的完整元数据（时间戳、类型、消息、载荷），适用于需要在前端直接展示详情而不二次请求的场景。

### `ProvenanceTree`

```go
type ProvenanceTree struct {
    ID      uint64           `json:"id"`
    IsRef   bool             `json:"is_ref"`
    Parents []ProvenanceTree `json:"parents"`
}
```

描述某个事件及其**所有祖先**的树形结构，自底向上追溯。与 `DescendantsTree` 方向相反，用于溯源（Provenance）场景。

### `ProvenanceTreeMeta`

```go
type ProvenanceTreeMeta struct {
    ID           uint64               `json:"id"`
    TimeUnixNano int64                `json:"time_unix_nano"`
    Type         string               `json:"type"`
    IsRef        bool                 `json:"is_ref"`
    Message      string               `json:"message,omitempty"`
    Payload      json.RawMessage      `json:"payload,omitempty"`
    Parents      []ProvenanceTreeMeta `json:"parents"`
}
```

带完整元数据的溯源树结构，语义同 `DescendantsTreeMeta`。

### `Snapshot`

```go
type Snapshot struct {
    Events      int    `json:"events"`
    Edges       int    `json:"edges"`
    Heads       int    `json:"heads"`
    Subscribers int    `json:"subscribers"`
    NextEventID uint64 `json:"next_event_id"`
}
```

系统运行时快照，暴露当前内存存储的核心统计指标：事件总数、边总数、当前 Head（无子节点的叶子事件）数量、SSE 订阅者数量以及下一个即将分配的事件 ID。

### `ResponseError`

```go
type ResponseError struct {
    Error  string `json:"error"`
    Detail string `json:"detail,omitempty"`
}
```

HTTP API 统一错误响应体。`Error` 为简短错误码/描述，`Detail` 可携带具体调试信息。

### `RootIDError`

```go
type RootIDError struct {
    ID     uint64
    Reason string
}
```

表示查询时传入的根 ID 无效。实现了 `error` 接口：

```go
func (e *RootIDError) Error() string
```

当 ID 为 0 或对应事件不存在时，`memory` 包内部会构造并返回此错误。

## 与其他文件的关系

| 依赖方向 | 文件/包 | 关系说明 |
|---------|--------|---------|
| 被导入 | `internal/memory/*` | `memory` 包所有操作均以 `tree.Event` / `tree.EmitRequest` 等为基础类型。 |
| 被导入 | `internal/httpapi/*` | HTTP Handler 读取 `tree.EmitRequest`、`tree.TreeBatchRequest`，返回 `tree.EmitResponse`、`tree.ResponseError` 等。 |
| 被导入 | `internal/grpcapi/*` | gRPC `Emit` 方法将 `pb.EmitRequest` 转换为 `tree.EmitRequest` 后调用存储层。 |
| 被导入 | `cmd/celestialtree/main.go` | 启动时创建创世事件 `tree.EmitRequest{Type: "genesis", ...}`。 |
| 标准库依赖 | `encoding/json`, `fmt` | 仅依赖标准库，保持最小耦合。 |

## 设计说明

- **延迟解析（Lazy Parsing）**：`Payload` 使用 `json.RawMessage` 而非 `map[string]any` 或具体结构体，避免在存储层做不必要的反序列化，同时保留原始 JSON 字节流的完整性。
- **方向分离**：Descendants（向下）与 Provenance（向上）分别定义树形结构，虽然字段相似，但独立类型可在未来演化中保持接口清晰，避免混淆。
- **引用标记（`IsRef`）**：在 DAG 中，同一个事件可能通过多条路径被访问到。`IsRef` 机制既支持树形序列化，又避免循环与爆炸性膨胀。
