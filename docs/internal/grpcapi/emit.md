# `emit.go`

## 文件整体描述

`emit.go` 是 **CelestialTree** 项目 gRPC 服务中 `Emit` RPC 的实现文件，位于 `internal/grpcapi` 包中。该文件的核心职责是将外部 gRPC 请求（`pb.EmitRequest`，携带 `google.protobuf.Struct` 类型的 Payload）转换为内部存储层可理解的 `tree.EmitRequest`（Payload 为 `json.RawMessage`），并调用 `memory.Store.Emit` 完成事件写入，最后将结果封装为 `pb.EmitResponse` 返回。

此文件是 gRPC 层与业务存储层之间的**适配器（Adapter）**，承担了协议转换、参数校验与错误码映射的职责。

## 函数说明

### `(*Server) Emit`

```go
func (s *Server) Emit(ctx context.Context, req *pb.EmitRequest) (*pb.EmitResponse, error)
```

gRPC `CelestialTreeService.Emit` 方法的实现，用于接收客户端请求并将新事件写入 DAG。

| 参数 | 类型 | 说明 |
|-----|------|------|
| `ctx` | `context.Context` | gRPC 调用上下文，可用于超时控制、链路追踪等（当前未显式使用，但保留以符合接口签名）。 |
| `req` | `*pb.EmitRequest` | 客户端传入的请求，包含 `Type`、`Message`、`Payload`（`google.protobuf.Struct`）和 `Parents`。 |

**返回值**：

| 返回值 | 类型 | 说明 |
|-------|------|------|
| `resp` | `*pb.EmitResponse` | 成功时返回，仅包含新创建事件的 `ID`。 |
| `err` | `error` | 失败时返回，已映射为 gRPC 标准 `status.Error`，错误码包括 `codes.InvalidArgument`。 |

**处理流程**：

1. **空请求校验**：若 `req == nil`，返回 `codes.InvalidArgument` 错误。
2. **Payload 协议转换**：使用 `protojson.Marshal` 将 `google.protobuf.Struct` 序列化为标准 JSON 字节流，再封装为 `json.RawMessage`。
   - 若序列化失败，返回 `codes.InvalidArgument`，并附带原始错误详情。
3. **调用存储层**：构造 `tree.EmitRequest`，传入 `s.store.Emit`。存储层会进一步校验 `Type` 非空、所有 `Parents` 存在等规则。
   - 若存储层返回错误，当前统一映射为 `codes.InvalidArgument`（未来可按错误类型细分，如 `codes.NotFound` 用于父事件缺失）。
4. **构造响应**：提取返回的 `tree.Event.ID`，封装为 `pb.EmitResponse` 返回。

## 与其他文件的关系

| 依赖方向 | 文件/包 | 关系说明 |
|---------|--------|---------|
| 导入 | `internal/tree` | 使用 `tree.EmitRequest` 作为存储层输入契约。 |
| 导入 | `internal/memory` | 通过 `s.store`（在 `server.go` 中注入的 `*memory.Store`）执行实际写入。 |
| 导入 | `proto`（`pb`） | 消费 `pb.EmitRequest`，生产 `pb.EmitResponse`；实现 `pb.CelestialTreeServiceServer` 接口。 |
| 标准库/第三方 | `google.golang.org/grpc/codes`, `google.golang.org/grpc/status` | 将内部错误映射为 gRPC 标准状态码。 |
| 标准库/第三方 | `google.golang.org/protobuf/encoding/protojson` | 将 Protobuf Struct 转为 JSON 字节。 |
| 同包协作 | `internal/grpcapi/server.go` | `emit.go` 中为 `Server` 类型扩展了 `Emit` 方法；`Server.store` 字段在此被消费。 |

## 设计说明

- **协议中立性**：存储层使用 `json.RawMessage` 而不感知 Protobuf，使得 gRPC 层成为唯一的 Protobuf 依赖点。未来若新增其他协议（如 Thrift、MsgPack），只需在对应入口层做转换，存储层无需改动。
- **错误码策略（v0）**：当前所有存储层错误统一映射为 `InvalidArgument`，这是为了先跑通端到端链路。后续迭代建议根据 `memory.Store.Emit` 返回的具体错误类型（如父事件不存在、Type 为空）映射到更精确的错误码（`NotFound`、`FailedPrecondition` 等）。
