# `main.go`

## 文件整体描述

`main.go` 是 **CelestialTree** 服务的入口文件，位于 `cmd/celestialtree` 目录中。该文件负责解析命令行参数、初始化内存存储引擎、启动 HTTP 与 gRPC 双协议服务器，并处理优雅关闭流程。

## 函数说明

### `Config`

```go
type Config struct {
    HTTPAddr string
    GRPCAddr string
}
```

HTTP 和 gRPC 服务的监听地址配置结构体。

### `parseConfig`

```go
func parseConfig() Config
```

从命令行参数解析服务配置。支持两种指定方式：
- **直接指定**：`-http_addr host:port`、`-grpc_addr host:port`（优先）。
- **组合指定**：`-host` + `-http_port` / `-grpc_port`（默认 `0.0.0.0:7777` / `0.0.0.0:7778`）。

### `newStoreWithGenesis`

```go
func newStoreWithGenesis() (*memory.Store, error)
```

创建 `memory.Store` 并写入创世事件（Genesis），作为 DAG 的起点。创世事件类型为 `"genesis"`，Message 为 `"CelestialTree begins."`。

### `newHTTPServer`

```go
func newHTTPServer(addr string, store *memory.Store) *http.Server
```

创建并配置 HTTP 服务器，注册所有 API 路由（通过 `httpapi.RegisterRoutes`）。配置包括 3 秒读取超时和 60 秒空闲超时。

### `newGRPCServer`

```go
func newGRPCServer(addr string, store *memory.Store) (*grpc.Server, net.Listener, error)
```

创建 gRPC 服务器并监听指定地址。注册 `CelestialTreeService` 实现和 reflection 服务（便于 `grpcurl` 调试）。

### `main`

程序主入口，执行流程：

1. 解析配置（`parseConfig`）。
2. 创建带创世事件的存储实例（`newStoreWithGenesis`）。
3. 启动 HTTP 和 gRPC 服务器（各自运行在独立 goroutine 中）。
4. 监听 `SIGINT` / `SIGTERM` 信号或服务器错误。
5. 收到信号后，先 `GracefulStop` gRPC，再带 5 秒超时 `Shutdown` HTTP。

## 与其他文件的关系

| 依赖方向 | 文件/包 | 关系说明 |
|---------|--------|---------|
| 导入 | `internal/memory` | 调用 `memory.NewStore()` 创建存储实例。 |
| 导入 | `internal/httpapi` | 调用 `httpapi.RegisterRoutes` 注册 HTTP 路由。 |
| 导入 | `internal/grpcapi` | 调用 `grpcapi.New(store)` 创建 gRPC 服务实现。 |
| 导入 | `internal/tree` | 使用 `tree.EmitRequest` 写入创世事件。 |
| 导入 | `internal/version` | 启动时输出版本信息。 |
| 导入 | `proto` | 注册 gRPC 服务接口。 |
