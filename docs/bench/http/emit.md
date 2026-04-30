# `emit.go`

## 文件整体描述

`emit.go` 是 **CelestialTree** 项目的 HTTP 协议性能基准测试工具，位于 `bench/http` 目录中。该工具通过并发 goroutine 向 `/emit` 端点发送大量请求，测量吞吐量（RPS）和延迟分布（p50/p90/p99/max）。

## 命令行参数

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `-base` | `http://127.0.0.1:7777` | CelestialTree HTTP 服务地址。 |
| `-n` | `10000` | 总请求数。 |
| `-c` | `20` | 并发 goroutine 数量。 |

## 实现细节

### `EmitReq`

```go
type EmitReq struct {
    Type    string   `json:"type"`
    Parents []uint64 `json:"parents"`
    Message string   `json:"message"`
    Payload []byte   `json:"payload"`
}
```

基准测试请求体结构体。默认构造一个 32 字节 Payload 的 `"bench"` 类型事件，`Parents` 为空。

### `main`

测试流程：

1. 解析命令行参数，配置 HTTP 连接池（最大 100 连接）。
2. 预生成 `n` 个任务到 channel。
3. 启动 `c` 个 worker goroutine，每个从 channel 取任务，发送 POST `/emit` 请求。
4. 记录每个请求的延迟，原子计数成功/失败。
5. 所有任务完成后，对延迟排序，输出统计结果。

**输出示例**：

```
[go-bench] total=10000 ok=10000 fail=0 rps=25000.5 lat_ms(p50=0.68 p90=1.20 p99=2.55 max=5.30)
```

## 与其他文件的关系

| 依赖方向 | 文件/包 | 关系说明 |
|---------|--------|---------|
| 目标服务 | `cmd/celestialtree/main.go` | 需要先启动 CelestialTree 服务，本工具向其发送请求。 |
| 对应端点 | `internal/httpapi/emit.go` | 测试 `POST /emit` 端点的吞吐量与延迟。 |
