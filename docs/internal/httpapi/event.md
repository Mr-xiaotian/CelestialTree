# `event.go`

## 文件整体描述

`event.go` 是 **CelestialTree** 项目 HTTP API 中负责**单事件查询**的处理器文件，位于 `internal/httpapi` 包中。该文件实现了 `/event/{id}` 端点，允许客户端根据事件 ID 精确检索单个事件的完整信息。

此接口是系统最基础的读取能力之一，与 `/children/`、`/ancestors/`、`/descendants/`、`/provenance/` 等树形查询接口形成互补：前者返回扁平化的单条记录，后者返回结构化的关系网络。

## 函数说明

### `handleGetEvent`

```go
func handleGetEvent(store *memory.Store) http.HandlerFunc
```

返回一个 HTTP Handler 函数，用于处理 `/event/{id}` 端点的 GET 请求。

| 参数 | 类型 | 说明 |
|-----|------|------|
| `store` | `*memory.Store` | 内存存储实例，通过闭包捕获。 |

**返回值**：`http.HandlerFunc` —— 标准 HTTP 处理函数。

**Handler 内部逻辑**：

1. **方法校验**：仅接受 `GET` 请求，其他方法返回 `405 Method Not Allowed`。
2. **路径参数解析**：调用 `parsePathUint64(w, r.URL.Path, "/event/")` 从 URL 中提取事件 ID。
   - 若 ID 缺失、非数字或为 `0`，返回 `400 Bad Request`。
3. **存储查询**：调用 `store.Get(id)`，在内存索引中检索事件。
   - 若事件不存在，返回 `404 Not Found`。
4. **响应**：成功时返回 `200 OK`，响应体为完整的 `tree.Event` JSON。

**请求示例**：

```
GET /event/42
```

**响应示例**：

```json
{
  "id": 42,
  "time_unix_nano": 1713709263000000000,
  "type": "user.signup",
  "message": "User alice signed up",
  "payload": {"email": "alice@example.com"},
  "parents": [1]
}
```

**错误响应示例**：

```json
{"error": "not found"}
```

## 与其他文件的关系

| 依赖方向 | 文件/包 | 关系说明 |
|---------|--------|---------|
| 导入 | `internal/memory` | 调用 `memory.Store.Get` 执行单条事件查询。 |
| 导入 | `internal/tree` | 使用 `tree.Event` 作为成功响应体，`tree.ResponseError` 作为错误响应体。 |
| 同包协作 | `internal/httpapi/common.go` | 调用 `requireMethod`、`parsePathUint64`、`writeJSON`。 |
| 同包协作 | `internal/httpapi/routes.go` | `RegisterRoutes` 中将 `/event/` 路径注册到此 Handler。 |

## 性能特征

`store.Get(id)` 底层为 `map[uint64]tree.Event` 的直接索引，时间复杂度 **O(1)**，无递归或遍历操作，是系统中响应速度最快的读取接口之一。
