# `server.go`

## 文件整体描述

`server.go` 是 **CelestialTree** 项目 gRPC 服务端的入口定义文件，位于 `internal/grpcapi` 包中。该文件负责声明 gRPC 服务结构体 `Server`，并提供其构造函数 `New`。`Server` 实现了由 Protobuf 编译生成的 `pb.CelestialTreeServiceServer` 接口，是 gRPC 层与业务存储层之间的唯一接合点。

当前 gRPC 服务仅暴露 `Emit` 一个 RPC（在 `emit.go` 中实现），未来可在此包中继续扩展其他方法（如查询事件、订阅流等）。

## 实体说明

### `Server`

```go
type Server struct {
    pb.UnimplementedCelestialTreeServiceServer
    store *memory.Store
}
```

gRPC 服务实现结构体。

- **`pb.UnimplementedCelestialTreeServiceServer`**：由 `protoc-gen-go-grpc` 生成的默认实现嵌入，确保在接口新增方法时服务仍能编译通过，避免破坏向前兼容。
- **`store *memory.Store`**：指向内存存储引擎的指针。所有 gRPC 请求最终都会委托给 `memory.Store` 的相应方法处理。`Server` 本身不持有业务状态，状态完全下沉到存储层。

### `New`

```go
func New(store *memory.Store) *Server
```

`Server` 的构造函数。

| 参数 | 类型 | 说明 |
|-----|------|------|
| `store` | `*memory.Store` | 已初始化好的内存存储实例，通常由 `cmd/celestialtree/main.go` 在启动时创建并注入。 |

**返回值**：`*Server` —— 可直接注册到 `grpc.Server` 上的服务实例。

## 与其他文件的关系

| 依赖方向 | 文件/包 | 关系说明 |
|---------|--------|---------|
| 导入 | `internal/memory` | 依赖 `memory.Store` 作为底层数据持久化（内存中）与业务逻辑执行者。 |
| 导入 | `proto`（`pb`） | 依赖由 `celestialtree.proto` 编译生成的 Go gRPC 接口与类型。 |
| 被调用 | `cmd/celestialtree/main.go` | `main.go` 通过 `grpcapi.New(store)` 创建服务实例，并注册到 gRPC 服务器：`pb.RegisterCelestialTreeServiceServer(srv, grpcapi.New(store))`。 |
| 同包协作 | `internal/grpcapi/emit.go` | `emit.go` 中为 `*Server` 实现了 `Emit` 方法，补全了 `pb.CelestialTreeServiceServer` 接口。 |

## 扩展建议

当需要新增 gRPC 方法时：

1. 更新 `proto/celestialtree.proto`，定义新的 RPC 与 Message；
2. 重新生成 Go 代码：`protoc --go_out=. --go-grpc_out=. celestialtree.proto`；
3. 在 `internal/grpcapi/` 下新建文件（如 `get_event.go`），为 `*Server` 新增对应方法；
4. 如需复用现有逻辑，确保 `memory.Store` 已提供相应能力。
