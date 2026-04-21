# `provenance.go`

## 文件整体描述

`provenance.go` 是 **CelestialTree** 项目 HTTP API 中负责**溯源树查询**的处理器文件，位于 `internal/httpapi` 包中。该文件实现了两个端点：

- `GET /provenance/{id}` —— 单事件溯源树查询
- `POST /provenance` —— 批量溯源树查询

溯源树（Provenance Tree）是指从某个事件出发，**向上**遍历其所有父事件、父事件的父事件……直到根节点，所形成的树形结构。它与后代树（Descendants Tree）方向相反：前者追溯“从哪里来”，后者探索“到哪里去”。

同样支持两种视图：

- `"struct"`（默认）：仅返回事件 ID 与树形骨架。
- `"meta"`：额外携带每个节点的完整元数据。

## 函数说明

### `handleProvenance`

```go
func handleProvenance(store *memory.Store) http.HandlerFunc
```

处理 `GET /provenance/{id}` 端点，查询单个事件的溯源树。

**Handler 内部逻辑**：

1. **方法校验**：仅接受 `GET`。
2. **路径解析**：调用 `parsePathUint64(w, r.URL.Path, "/provenance/")` 提取根事件 ID。
3. **视图参数解析**：调用 `normalizeView(r.URL.Query().Get("view"))` 规范化 `view` 查询参数。
4. **分支处理**：
   - `view` 为空或 `"struct"`：调用 `store.ProvenanceTree(id)`，返回 `tree.ProvenanceTree`。
   - `view` 为 `"meta"`：调用 `store.ProvenanceTreeMeta(id)`，返回 `tree.ProvenanceTreeMeta`。
   - 其他值：返回 `400 Bad Request`，提示 `unknown view`。
5. **错误处理**：若根事件不存在，返回 `404 Not Found`。
6. **响应**：成功时返回 `200 OK` 与对应树形结构的 JSON。

**请求示例**：

```
GET /provenance/42?view=meta
```

### `handleProvenanceBatch`

```go
func handleProvenanceBatch(store *memory.Store) http.HandlerFunc
```

处理 `POST /provenance` 端点，支持一次性查询**多个事件**的溯源树，返回一片“森林”。

**Handler 内部逻辑**：

1. **方法校验**：仅接受 `POST`。
2. **请求体解析**：调用 `readJSON(r, &req)` 将请求体反序列化为 `tree.TreeBatchRequest`。
   - 若 JSON 非法，返回 `400 Bad Request`。
   - 若 `req.IDs` 为空数组，返回 `400 Bad Request`，提示 `ids is required`。
3. **视图参数解析**：调用 `normalizeView(req.View)` 规范化视图参数。
4. **分支处理**：
   - `view` 为空或 `"struct"`：调用 `store.ProvenanceForest(req.IDs)`，返回 `[]tree.ProvenanceTree`。
   - `view` 为 `"meta"`：调用 `store.ProvenanceForestMeta(req.IDs)`，返回 `[]tree.ProvenanceTreeMeta`。
   - 其他值：返回 `400 Bad Request`。
5. **错误处理**：若任一 ID 无效，返回 `404 Not Found`。
6. **响应**：成功时返回 `200 OK` 与森林结构的 JSON 数组。

**请求示例**：

```json
POST /provenance
Content-Type: application/json

{
  "ids": [42, 43],
  "view": "meta"
}
```

**响应示例**（`meta` 视图）：

```json
[
  {
    "id": 42,
    "time_unix_nano": 1713709263000000000,
    "type": "user.signup",
    "is_ref": false,
    "parents": [
      {
        "id": 1,
        "time_unix_nano": 1713709200000000000,
        "type": "genesis",
        "is_ref": false,
        "parents": []
      }
    ]
  }
]
```

## 与其他文件的关系

| 依赖方向 | 文件/包 | 关系说明 |
|---------|--------|---------|
| 导入 | `internal/memory` | 调用 `memory.Store.ProvenanceTree`、`ProvenanceTreeMeta`、`ProvenanceForest`、`ProvenanceForestMeta`。 |
| 导入 | `internal/tree` | 使用 `tree.TreeBatchRequest` 接收批量请求体，`tree.ResponseError` 构造错误响应。 |
| 同包协作 | `internal/httpapi/common.go` | 调用 `requireMethod`、`parsePathUint64`、`normalizeView`、`readJSON`、`writeJSON`。 |
| 同包协作 | `internal/httpapi/routes.go` | `RegisterRoutes` 中将 `/provenance/` 与 `/provenance` 注册到对应 Handler。 |

## 设计说明

- **方向对称性**：`provenance.go` 与 `descendants.go` 在代码结构上高度对称（单条查询 + 批量查询、`view` 参数、错误处理模式），这是有意为之，降低维护者心智负担。
- **溯源 vs 祖先**：`/provenance/{id}` 返回完整的树形结构（包含中间节点），而 `/ancestors/{id}` 仅返回终极根节点集合。前者适合可视化血缘，后者适合快速定位源头。
