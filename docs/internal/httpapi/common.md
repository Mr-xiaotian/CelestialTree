# `common.go`

## 文件整体描述

`common.go` 是 **CelestialTree** 项目 HTTP API 层的公共工具文件，位于 `internal/httpapi` 包中。该文件不包含任何 HTTP 路由处理器，而是提供一组被多个 Handler 复用的**通用辅助函数**，涵盖 HTTP 方法校验、路径参数解析、查询参数规范化、JSON 读写等基础能力。

将这些通用逻辑抽离到独立文件，可以避免在各个 Handler 文件中重复编写样板代码，同时统一错误响应格式与解析行为。

## 函数说明

### `requireMethod`

```go
func requireMethod(w http.ResponseWriter, r *http.Request, method string) bool
```

校验当前 HTTP 请求的方法是否匹配预期。

| 参数 | 类型 | 说明 |
|-----|------|------|
| `w` | `http.ResponseWriter` | 响应写入器，校验不通过时直接写入 405 错误。 |
| `r` | `*http.Request` | 当前请求对象。 |
| `method` | `string` | 预期的方法（如 `http.MethodGet`、`http.MethodPost`）。 |

**返回值**：`bool` —— `true` 表示方法匹配，调用方可继续处理；`false` 表示不匹配，此时函数已向客户端写入 `405 Method Not Allowed` 及 JSON 错误体，调用方应直接 `return`。

**错误响应格式**：

```json
{"error": "method not allowed"}
```

### `parsePathUint64`

```go
func parsePathUint64(w http.ResponseWriter, path, prefix string) (uint64, bool)
```

从 URL 路径中解析出 `uint64` 类型的 ID 参数。常用于 `/event/{id}`、`/children/{id}` 等路径模式。

| 参数 | 类型 | 说明 |
|-----|------|------|
| `w` | `http.ResponseWriter` | 解析失败时直接写入 400 错误。 |
| `path` | `string` | 完整 URL 路径（通常取 `r.URL.Path`）。 |
| `prefix` | `string` | 路径前缀，解析时会将其从 `path` 中去除，剩余部分即为 ID 字符串。 |

**返回值**：

- `uint64`：解析成功的 ID。
- `bool`：`true` 表示解析成功；`false` 表示失败，函数已写入 400 错误响应。

**错误响应格式**：

```json
{"error": "bad id"}
```

**校验规则**：

- 去除前缀后的字符串必须能被 `strconv.ParseUint` 解析为 10 进制 `uint64`。
- 解析结果不能为 `0`（系统中 `0` 不是合法事件 ID）。

### `normalizeView`

```go
func normalizeView(s string) string
```

规范化视图查询参数。对 `descendants` 与 `provenance` 接口的 `view` 参数做统一预处理：去除首尾空白并转为小写，保证 `"Struct"`、`"STRUCT"`、 `" struct "` 等输入都被统一识别。

| 参数 | 类型 | 说明 |
|-----|------|------|
| `s` | `string` | 原始查询字符串（通常取 `r.URL.Query().Get("view")` 或请求 JSON 中的 `View` 字段）。 |

**返回值**：`string` —— 规范化后的小写字符串。

### `writeJSON`

```go
func writeJSON(w http.ResponseWriter, status int, v any)
```

向客户端写入 JSON 响应的统一封装。自动设置 `Content-Type: application/json; charset=utf-8` 头，并写入指定的 HTTP 状态码，最后将 `v` 序列化为 JSON 写入响应体。

| 参数 | 类型 | 说明 |
|-----|------|------|
| `w` | `http.ResponseWriter` | 响应写入器。 |
| `status` | `int` | HTTP 状态码（如 `200`、`400`、`404`）。 |
| `v` | `any` | 待序列化的响应体，通常为 `tree` 包中的结构体或 `map[string]any`。 |

**注意**：编码错误被忽略（`_ = json.NewEncoder(w).Encode(v)`），因为在写入响应头之后已难以返回新的 HTTP 错误，且 JSON 编码失败通常意味着程序内部存在严重类型不匹配。

### `readJSON`

```go
func readJSON(r *http.Request, v any) error
```

从请求体中读取并反序列化 JSON 的统一封装。使用 `json.NewDecoder` 并启用 `DisallowUnknownFields`，即客户端若传入结构体定义中不存在的字段，会返回错误，防止因拼写错误或版本差异导致的数据静默丢失。

| 参数 | 类型 | 说明 |
|-----|------|------|
| `r` | `*http.Request` | 当前请求对象，从 `r.Body` 中读取数据。 |
| `v` | `any` | 接收反序列化结果的目标指针。 |

**返回值**：`error` —— 反序列化失败时返回具体错误（如 JSON 语法错误、未知字段、类型不匹配）。

## 与其他文件的关系

| 依赖方向 | 文件/包 | 关系说明 |
|---------|--------|---------|
| 导入 | `internal/tree` | 使用 `tree.ResponseError` 作为方法不匹配时的错误响应体。 |
| 被调用 | `internal/httpapi/emit.go` | `handleEmit` 调用 `requireMethod`、`readJSON`、`writeJSON`。 |
| 被调用 | `internal/httpapi/event.go` | `handleGetEvent` 调用 `requireMethod`、`parsePathUint64`、`writeJSON`。 |
| 被调用 | `internal/httpapi/graph.go` | `handleChildren`、`handleAncestors`、`handleHeads`、`handleRoots` 调用 `requireMethod`、`parsePathUint64`、`writeJSON`。 |
| 被调用 | `internal/httpapi/descendants.go` | `handleDescendants`、`handleDescendantsBatch` 调用 `requireMethod`、`parsePathUint64`、`normalizeView`、`readJSON`、`writeJSON`。 |
| 被调用 | `internal/httpapi/provenance.go` | `handleProvenance`、`handleProvenanceBatch` 调用 `requireMethod`、`parsePathUint64`、`normalizeView`、`readJSON`、`writeJSON`。 |
| 被调用 | `internal/httpapi/snapshot.go` | `handleSnapshot` 调用 `requireMethod`、`writeJSON`。 |
| 被调用 | `internal/httpapi/sse.go` | `handleSubscribe` 调用 `writeJSON` 用于返回非 SSE 的错误响应。 |

## 设计说明

- **集中式错误处理**：所有辅助函数在出错时都直接向 `http.ResponseWriter` 写入响应，调用方只需检查 `bool` 返回值决定是否提前 `return`。这种“快速失败”模式减少了 Handler 中的 `if err != nil` 嵌套深度。
- **严格输入校验**：`readJSON` 启用 `DisallowUnknownFields`，体现了“早失败、显式失败”的设计哲学，有助于在 API 演化过程中及时发现客户端版本不匹配问题。
