# `graph.go`

## 文件整体描述

`graph.go` 是 **CelestialTree** 项目 HTTP API 中负责**图关系查询**的处理器文件，位于 `internal/httpapi` 包中。该文件实现了四个与 DAG（有向无环图）拓扑结构直接相关的读取端点：

- `/children/{id}` —— 查询某事件的直接子事件
- `/ancestors/{id}` —— 查询某事件的所有根祖先
- `/heads` —— 查询当前所有叶子事件（Head）
- `/roots` —— 查询当前所有创世事件（Root）

这些接口共同构成了对 DAG 局部拓扑与全局概览的查询能力，是理解事件血缘关系与系统状态的核心入口。

## 函数说明

### `handleChildren`

```go
func handleChildren(store *memory.Store) http.HandlerFunc
```

处理 `/children/{id}` 端点，返回指定事件的**直接子事件 ID 列表**。

**Handler 内部逻辑**：

1. 方法校验：仅接受 `GET`。
2. 路径解析：从 `/children/{id}` 中提取 `uint64` 类型的 ID。
3. 存储查询：调用 `store.Children(id)`。
   - 若事件本身不存在，返回 `404 Not Found`。
   - 若事件存在但无子事件，返回 `200 OK` 与空数组 `[]`。
4. 响应：返回 `[]uint64` 的 JSON 数组。

### `handleAncestors`

```go
func handleAncestors(store *memory.Store) http.HandlerFunc
```

处理 `/ancestors/{id}` 端点，返回指定事件的**所有根祖先（Roots）**。注意：该接口返回的是沿着 DAG 向上追溯后到达的终极根节点集合，而非完整的祖先路径树。如需完整路径树，应使用 `/provenance/{id}`。

**Handler 内部逻辑**：

1. 方法校验：仅接受 `GET`。
2. 路径解析：从 `/ancestors/{id}` 中提取 ID。
3. 存储查询：调用 `store.Ancestors(id)`。
   - 若事件不存在或遍历过程中发现父事件缺失（数据不一致），返回 `404 Not Found`。
4. 响应：返回排序后的 `[]uint64` JSON 数组。

### `handleHeads`

```go
func handleHeads(store *memory.Store) http.HandlerFunc
```

处理 `/heads` 端点，返回当前 DAG 中所有**无子节点的叶子事件（Heads）**的 ID 列表。Head 集合代表了 DAG 的“前沿”，所有最新写入、尚未被后续事件引用的事件都会出现在此列表中。

**Handler 内部逻辑**：

1. 方法校验：仅接受 `GET`。
2. 存储查询：调用 `store.Heads()`。
3. 响应：返回 `[]uint64` JSON 数组。若当前无事件，返回空数组。

### `handleRoots`

```go
func handleRoots(store *memory.Store) http.HandlerFunc
```

处理 `/roots` 端点，返回当前 DAG 中所有**无父事件的创世事件（Roots）**的 ID 列表。Root 集合代表了 DAG 的起点，通常至少包含系统启动时自动生成的 Genesis 事件。

**Handler 内部逻辑**：

1. 方法校验：仅接受 `GET`。
2. 存储查询：调用 `store.Roots()`。
3. 响应：返回 `[]uint64` JSON 数组。

## 与其他文件的关系

| 依赖方向 | 文件/包 | 关系说明 |
|---------|--------|---------|
| 导入 | `internal/memory` | 调用 `memory.Store.Children`、`memory.Store.Ancestors`、`memory.Store.Heads`、`memory.Store.Roots`。 |
| 导入 | `internal/tree` | 使用 `tree.ResponseError` 构造错误响应。 |
| 同包协作 | `internal/httpapi/common.go` | 调用 `requireMethod`、`parsePathUint64`、`writeJSON`。 |
| 同包协作 | `internal/httpapi/routes.go` | `RegisterRoutes` 中将 `/children/`、`/ancestors/`、`/heads`、`/roots` 注册到对应 Handler。 |

## 设计说明

- **Heads vs Roots**：Head 是 DAG 的“终点”（无出边），Root 是 DAG 的“起点”（无入边）。理解这一区分对正确使用查询接口至关重要。
- **Ancestors 的返回值**：`store.Ancestors` 内部使用 DFS 向上遍历，但只返回最终到达的根节点集合，而非中间节点。这种设计适合快速定位事件的“源头”，但不适合展示完整血缘路径。完整路径请使用 `/provenance/{id}`。
