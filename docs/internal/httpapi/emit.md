# `emit.go`

## 文件整体描述

`emit.go` 是 **CelestialTree** 项目 HTTP API 中负责**事件写入**的处理器文件，位于 `internal/httpapi` 包中。该文件实现了 `/emit` 端点，接收客户端 POST 请求，将事件数据写入内存存储引擎，并返回新创建事件的 ID。

`/emit` 是系统的核心写入接口，与 gRPC 的 `Emit` RPC 共同构成 CelestialTree 的双协议写入通道。

## 函数说明

### `handleEmit`

```go
func handleEmit(store *memory.Store) http.HandlerFunc
```

返回一个 HTTP Handler 函数，用于处理 `/emit` 端点的请求。

| 参数 | 类型 | 说明 |
|-----|------|------|
| `store` | `*memory.Store` | 内存存储实例，通过闭包捕获，在 Handler 内部调用其 `Emit` 方法。 |

**返回值**：`http.HandlerFunc` —— 可直接注册到 `http.ServeMux` 的标准 HTTP 处理函数。

**Handler 内部逻辑**：

1. **方法校验**：仅接受 `POST` 请求，其他方法返回 `405 Method Not Allowed`。
2. **请求体解析**：调用 `readJSON(r, &req)` 将请求体反序列化为 `tree.EmitRequest`。
   - 若 JSON 格式非法或包含未知字段，返回 `400 Bad Request`。
3. **存储写入**：调用 `store.Emit(req)`，将事件持久化到内存 DAG 中。
   - 若 `Type` 为空、父事件不存在等校验失败，返回 `400 Bad Request` 并携带具体错误详情。
4. **响应**：成功时返回 `200 OK`，响应体为 `tree.EmitResponse{ID: ev.ID}`。

**请求示例**：

```json
POST /emit
Content-Type: application/json

{
  "type": "user.signup",
  "message": "User alice signed up",
  "payload": {"email": "alice@example.com"},
  "parents": [1]
}
```

**响应示例**：

```json
{"id": 42}
```

**错误响应示例**：

```json
{"error": "emit failed", "detail": "parent 99 not found"}
```

## 与其他文件的关系

| 依赖方向 | 文件/包 | 关系说明 |
|---------|--------|---------|
| 导入 | `internal/memory` | 调用 `memory.Store.Emit` 执行实际写入。 |
| 导入 | `internal/tree` | 使用 `tree.EmitRequest` 接收请求体，`tree.EmitResponse` 构造成功响应，`tree.ResponseError` 构造错误响应。 |
| 同包协作 | `internal/httpapi/common.go` | 调用 `requireMethod`、`readJSON`、`writeJSON` 完成通用 HTTP 处理。 |
| 同包协作 | `internal/httpapi/routes.go` | `RegisterRoutes` 中将 `/emit` 路径注册到此 Handler。 |
| 被调用 | `cmd/celestialtree/main.go` | 启动流程中通过 `httpapi.RegisterRoutes` 间接挂载此端点。 |

## 设计说明

- **幂等性提示**：当前 `Emit` 接口不具备幂等性（客户端重复调用会产生多条独立事件）。若业务场景需要幂等写入，建议在请求体中增加 `client_request_id` 字段，并在 `memory.Store` 层实现去重逻辑。
- **同步写入**：HTTP `/emit` 为同步接口，事件写入后立即返回。对于高吞吐场景，未来可考虑在 `memory.Store` 层引入异步批处理或 WAL（Write-Ahead Log）机制，但 HTTP 接口本身可保持同步语义不变。
