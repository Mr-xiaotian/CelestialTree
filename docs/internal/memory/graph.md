# `graph.go`

## 文件整体描述

`graph.go` 是 **CelestialTree** 项目内存存储引擎中负责**图拓扑关系查询**的实现文件，位于 `internal/memory` 包中。该文件实现了四个与 DAG 局部结构和全局概览相关的读取方法：

- `Children(id)` —— 查询某事件的直接子事件
- `Ancestors(id)` —— 查询某事件的所有根祖先
- `Heads()` —— 查询当前所有叶子事件
- `Roots()` —— 查询当前所有创世事件

这些方法为 HTTP API 的 `/children/`、`/ancestors/`、`/heads`、`/roots` 端点提供底层数据支持。

## 函数说明

### `(*Store) Children`

```go
func (s *Store) Children(id uint64) ([]uint64, bool)
```

查询指定事件的**直接子事件 ID 列表**。

| 参数 | 类型 | 说明 |
|-----|------|------|
| `id` | `uint64` | 父事件 ID。 |

**返回值**：

- `[]uint64`：子事件 ID 列表。若事件无子事件，返回空数组 `[]`（而非 `nil`）。
- `bool`：`true` 表示父事件存在；`false` 表示父事件不存在。

**实现细节**：在 `s.mu` 保护下，先通过 `isEventIDValid` 检查父事件是否存在，再读取 `s.children[id]`，将列表拷贝到新切片后返回。返回的是内部数据的**副本**，调用方可安全修改而不影响内部状态。

### `(*Store) Ancestors`

```go
func (s *Store) Ancestors(id uint64) ([]uint64, bool)
```

查询指定事件的**所有根祖先（Roots）**。从该事件出发，沿着 `Parents` 链向上 DFS 遍历，收集所有到达的终极根节点（即 `Parents` 为空的事件）。

| 参数 | 类型 | 说明 |
|-----|------|------|
| `id` | `uint64` | 起始事件 ID。 |

**返回值**：

- `[]uint64`：根祖先 ID 列表，按升序排序。
- `bool`：`true` 表示遍历成功；`false` 表示起始事件不存在或遍历中发现父事件缺失（数据不一致）。

**实现细节**：

1. 在 `s.mu` 保护下执行。
2. 使用内部闭包函数 `dfs(cur uint64) bool` 递归遍历：
   - `visited` 集合防止重复访问（处理 DAG 中多路径汇聚到同一祖先的情况）。
   - `roots` 集合收集所有无父事件的起始点。
   - 通过 `isEventIDValid` 检查每个节点的存在性，若某 `cur` 或其父事件不存在，返回 `false`，整个 `Ancestors` 调用失败。
3. 对收集到的根 ID 列表使用 `slices.Sort` 升序排序后返回。

**时间复杂度**：O(V+E) 在最坏情况下，其中 V 为访问的节点数，E 为访问的边数。实际中由于 DAG 通常不深，开销很小。

### `(*Store) Heads`

```go
func (s *Store) Heads() []uint64
```

查询当前 DAG 中所有**无子事件的叶子事件（Heads）**的 ID 列表。

**返回值**：`[]uint64` —— Head 事件 ID 列表。不保证顺序。

**实现细节**：在 `s.mu` 保护下遍历 `s.heads` 集合，收集所有 ID 到切片中返回。

**时间复杂度**：O(|heads|)。

### `(*Store) Roots`

```go
func (s *Store) Roots() []uint64
```

查询当前 DAG 中所有**无父事件的创世事件（Roots）**的 ID 列表。

**返回值**：`[]uint64` —— Root 事件 ID 列表。不保证顺序。

**实现细节**：在 `s.mu` 保护下遍历 `s.roots` 集合，收集所有 ID 到切片中返回。

**时间复杂度**：O(|roots|)。

## 与其他文件的关系

| 依赖方向 | 文件/包 | 关系说明 |
|---------|--------|---------|
| 标准库 | `slices` | `Ancestors` 中使用 `slices.Sort` 对根 ID 列表排序。 |
| 同包协作 | `internal/memory/store.go` | 读取 `Store.events`、`Store.children`、`Store.heads`、`Store.roots`。 |
| 被调用 | `internal/httpapi/graph.go` | HTTP Handler 调用 `store.Children`、`store.Ancestors`、`store.Heads`、`store.Roots` 构造响应。 |

## 设计说明

- **Heads/Roots 的实时性**：`heads` 与 `roots` 是 `Store` 在 `Emit` 时实时维护的集合，而非惰性计算。这使得 `Heads()` 与 `Roots()` 的查询非常快，仅需遍历少量元素。
- **Ancestors 的防御性编程**：`dfs` 闭包在遍历中若发现某父事件不存在，会立即返回 `false`。这能在数据不一致（如索引损坏）时及时暴露问题，而不是静默返回错误结果。
- **排序策略**：`Ancestors` 对结果排序是为了给调用方提供稳定、可测试的输出；`Children` 返回的是内部 slice 的拷贝，顺序与插入顺序一致（即 `Emit` 写入的顺序）。若需要确定性升序，可在外部排序。
