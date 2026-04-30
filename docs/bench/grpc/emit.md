# `emit.go`

## 文件整体描述

`emit.go` 是 **CelestialTree** 项目的 gRPC 协议性能基准测试工具，位于 `bench/grpc` 目录中。该工具通过并发 goroutine 向 gRPC `Emit` RPC 发送大量请求，测量吞吐量（RPS）和延迟分布（p50/p90/p99/max）。

## 命令行参数

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `-addr` | `127.0.0.1:7778` | CelestialTree gRPC 服务地址。 |
| `-n` | `10000` | 总请求数。 |
| `-c` | `20` | 并发 goroutine 数量。 |
| `-timeout` | `10s` | 每个请求的超时时间。 |

## 实现细节

### `main`

测试流程：

1. 解析命令行参数，建立单个 gRPC 连接（`grpc.WithBlock`，不安全凭据）。
2. 构造 `pb.EmitRequest`，包含 32 字节 `google.protobuf.Struct` Payload。
3. 预生成 `n` 个任务到 channel。
4. 启动 `c` 个 worker goroutine，每个从 channel 取任务，调用 `client.Emit` RPC。
5. 每个请求独立创建 `context.WithTimeout`，记录延迟，原子计数成功/失败。
6. 所有任务完成后，对延迟排序，输出统计结果。

**输出示例**：

```
[go-bench-grpc] total=10000 ok=10000 fail=0 rps=35000.2 lat_ms(p50=0.45 p90=0.88 p99=1.90 max=4.10)
```

**设计要点**：

- **单连接复用**：gRPC 客户端是并发安全的，所有 worker 共享同一连接，模拟真实生产环境的连接复用模式。
- **与 HTTP bench 对比**：配合 `bench/http/emit.go` 可对比相同负载下 HTTP 与 gRPC 协议的性能差异。

## 与其他文件的关系

| 依赖方向 | 文件/包 | 关系说明 |
|---------|--------|---------|
| 导入 | `proto` | 使用 `pb.EmitRequest`、`pb.CelestialTreeServiceClient`。 |
| 目标服务 | `cmd/celestialtree/main.go` | 需要先启动 CelestialTree 服务，本工具向其发送请求。 |
| 对应端点 | `internal/grpcapi/emit.go` | 测试 gRPC `Emit` RPC 的吞吐量与延迟。 |
