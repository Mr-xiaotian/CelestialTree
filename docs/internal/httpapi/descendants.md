# `descendants.go`

## 文件整体描述

`descendants.go` 是 **CelestialTree** 项目 HTTP API 中负责**后代树查询**的处理器文件，位于 `internal/httpapi` 包中。该文件实现了两个端点：

- `GET /descendants/{id}` —— 单事件后代树查询
- `POST /descendants` —— 批量后代树查询

后代树（Descendants Tree）是指从某个事件出发，向下遍历其所有子事件、子事件的子事件……直到叶子节点，所形成的树形结构。该接口支持两种视图（View）：

- `"struct"`（默认）：仅返回事件 ID 与树形骨架。
- `"meta"`：在树形骨架基础上，额外携带每个节点的时间戳、类型、消息、载荷等完整元数据。

## 函数说明

### `handleDescendants`

```go
func handleDescendants(store *memory.Store) http.HandlerFunc
```

处理 `GET /descendants/{id}` 端点，查询单个事件的后代树。

**Handler 内部逻辑**：

1. **方法校验**：仅接受 `GET`。
2. **路径解析**：调用 `parsePathUint64(w, r.URL.Path, "/descendants/")` 提取根事件 ID。
3. **视图参数解析**：调用 `normalizeView(r.URL.Query().Get("view"))` 规范化 `view` 查询参数。
4. **分支处理**：
   - `view` 为空或 `"struct"`：调用 `store.DescendantsTree(id)`，返回 `tree.DescendantsTree`。
   - `view` 为 `"meta"`：调用 `store.DescendantsTreeMeta(id)`，返回 `tree.DescendantsTreeMeta`。
   - 其他值：返回 `400 Bad Request`，提示 `unknown view`。
5. **错误处理**：若根事件不存在或遍历失败，返回 `404 Not Found`。
6. **响应**：成功时返回 `200 OK` 与对应树形结构的 JSON。

**请求示例**：

```
GET /descendants/42?view=meta
```

### `handleDescendantsBatch`

```go
func handleDescendantsBatch(store *memory.Store) http.HandlerFunc
```

处理 `POST /descendants` 端点，支持一次性查询**多个事件**的后代树，返回一片“森林”（Forest）。

**Handler 内部逻辑**：

1. **方法校验**：仅接受 `POST`。
2. **请求体解析**：调用 `readJSON(r, &req)` 将请求体反序列化为 `tree.TreeBatchRequest`。
   - 若 JSON 非法，返回 `400 Bad Request`。
   - 若 `req.IDs` 为空数组，返回 `400 Bad Request`，提示 `ids is required`。
3. **视图参数解析**：调用 `normalizeView(req.View)` 规范化视图参数。
4. **分支处理**：
   - `view` 为空或 `"struct"`：调用 `store.DescendantsForest(req.IDs)`，返回 `[]tree.DescendantsTree`。
   - `view` 为 `"meta"`：调用 `store.DescendantsForestMeta(req.IDs)`，返回 `[]tree.DescendantsTreeMeta`。
   - 其他值：返回 `400 Bad Request`。
5. **错误处理**：若任一 ID 无效，返回 `404 Not Found`。
6. **响应**：成功时返回 `200 OK` 与森林结构的 JSON 数组。

**请求示例**：

```json
POST /descendants
Content-Type: application/json

{
  "ids": [42, 43],
  "view": "struct"
}
```

**响应示例**（`struct` 视图）：

```json
[
  {
    "id": 42,
    "is_ref": false,
    "children": [
      {"id": 50, "is_ref": false, "children": []}
    ]
  },
  {
    "id": 43,
    "is_ref": false,
    "children": []
  }
]
```

## 与其他文件的关系

| 依赖方向 | 文件/包 | 关系说明 |
|---------|--------|---------|
| 导入 | `internal/memory` | 调用 `memory.Store.DescendantsTree`、`DescendantsTreeMeta`、`DescendantsForest`、`DescendantsForestMeta`。 |
| 导入 | `internal/tree` | 使用 `tree.TreeBatchRequest` 接收批量请求体，`tree.ResponseError` 构造错误响应。 |
| 同包协作 | `internal/httpapi/common.go` | 调用 `requireMethod`、`parsePathUint64`、`normalizeView`、`readJSON`、`writeJSON`。 |
| 同包协作 | `internal/httpapi/routes.go` | `RegisterRoutes` 中将 `/descendants/` 与 `/descendants` 注册到对应 Handler。 |

## 设计说明

- **视图分离**：通过 `view` 参数在同一端点上提供轻量与详细两种输出模式，避免为不同需求维护两套几乎重复的端点。
- **批量接口**：`POST /descendants` 使用请求体而非查询字符串传递 ID 列表，是因为 ID 数量可能很多，URL 查询字符串存在长度限制（不同浏览器/代理限制不同，通常 2KB~8KB）。
- **环检测**：`memory` 层的后代树构建已实现环检测（`visited` 集合 + `IsRef` 标记），即使未来 DAG 中意外出现循环引用，也不会导致无限递归或栈溢出。
