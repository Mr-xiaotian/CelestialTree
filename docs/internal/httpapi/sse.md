# `sse.go`

## 文件整体描述

`sse.go` 是 **CelestialTree** 项目 HTTP API 中负责**服务端推送（Server-Sent Events）**的处理器文件，位于 `internal/httpapi` 包中。该文件实现了 `/subscribe` 端点，允许客户端通过 HTTP 长连接实时订阅新写入的事件流。

与 WebSocket 相比，SSE 基于纯 HTTP，无需协议升级，实现更轻量，且天然支持自动重连（通过 `EventSource` 的 `retry` 机制）。CelestialTree 选择 SSE 作为实时推送方案，是为了在保持简单性的同时满足事件订阅需求。

## 函数说明

### `handleSubscribe`

```go
func handleSubscribe(store *memory.Store) http.HandlerFunc
```

处理 `/subscribe` 端点，建立 SSE 长连接并持续向客户端推送新事件。

| 参数 | 类型 | 说明 |
|-----|------|------|
| `store` | `*memory.Store` | 内存存储实例，通过闭包捕获。 |

**返回值**：`http.HandlerFunc` —— 标准 HTTP 处理函数。

**Handler 内部逻辑**：

1. **方法校验**：仅接受 `GET`，其他方法返回 `405 Method Not Allowed`。
2. **流支持校验**：检查 `http.ResponseWriter` 是否实现了 `http.Flusher` 接口。若不支持（如某些中间件包装器），返回 `500 Internal Server Error`，提示 `streaming not supported`。
3. **响应头设置**：
   - `Content-Type: text/event-stream`
   - `Cache-Control: no-cache`
   - `Connection: keep-alive`
4. **订阅注册**：调用 `store.Subscribe()`，获取订阅 ID、事件通道 `ch` 与取消函数 `cancel`。
   - 使用 `defer cancel()` 确保连接断开时清理订阅资源，关闭通道。
5. **握手消息**：向客户端发送一个 `hello` 事件，表示订阅成功：

```
event: hello
data: {"message":"subscribed"}

```

6. **事件循环**：通过 `select` 监听两个通道：
   - `r.Context().Done()`：客户端断开连接或请求超时，退出循环，Handler 返回。
   - `ch`：收到新事件时，将其序列化为 JSON，以 SSE 格式写入响应并立即 Flush：

```
event: emit
data: {"id":42,"time_unix_nano":...,"type":"user.signup",...}

```

## 与其他文件的关系

| 依赖方向 | 文件/包 | 关系说明 |
|---------|--------|---------|
| 导入 | `internal/memory` | 调用 `memory.Store.Subscribe()` 注册订阅，接收实时事件流。 |
| 导入 | `internal/tree` | 使用 `tree.Event` 作为通道传输的数据类型，`tree.ResponseError` 构造非 SSE 的错误响应。 |
| 同包协作 | `internal/httpapi/common.go` | 调用 `writeJSON` 返回方法不匹配或流不支持时的 JSON 错误响应。 |
| 同包协作 | `internal/httpapi/routes.go` | `RegisterRoutes` 中将 `/subscribe` 注册到此 Handler。 |

## 设计说明

- **非阻塞广播**：`memory.Store.broadcast()` 在向订阅者通道发送事件时使用 `select` + `default`，若通道已满则直接丢弃事件。这是 v0 版本的取舍：优先保证 `Emit` 接口的写入性能不被慢消费者拖累，牺牲的是极端情况下订阅者可能丢失事件。
  - 若业务场景要求强一致性推送，未来可改为带超时的阻塞发送，或为每个订阅者维护独立的发送队列。
- **通道缓冲**：`store.Subscribe()` 为每个订阅者创建缓冲大小为 64 的通道，可短暂平滑写入突发流量，但不能应对长期消费落后。
- **资源清理**：`defer cancel()` 确保即使客户端异常断开（如网络抖动、浏览器关闭），订阅通道也会被关闭并从 `store.subs` 中移除，防止 goroutine 与内存泄漏。
- **无历史回放**：当前 SSE 连接仅推送订阅**之后**产生的新事件。若客户端需要历史事件，应在建立 SSE 连接前先调用 `/event/`、`/descendants/` 等接口补齐数据。

## 客户端使用示例（JavaScript）

```javascript
const es = new EventSource('/subscribe');

es.addEventListener('hello', (e) => {
  console.log('Connected:', JSON.parse(e.data));
});

es.addEventListener('emit', (e) => {
  const ev = JSON.parse(e.data);
  console.log('New event:', ev);
});

es.addEventListener('error', (e) => {
  console.error('SSE error, will auto-reconnect:', e);
});
```
