# `descendants.go`

## 文件整体描述

`descendants.go` 是 **CelestialTree** 项目内存存储引擎中负责**后代树构建**的实现文件，位于 `internal/memory` 包中。该文件实现了从指定事件出发，**向下**遍历所有子事件、子事件的子事件……直至叶子节点，构建完整后代树的算法。

提供四种公开方法：

- `DescendantsTree(id)` —— 单事件后代树（结构视图）
- `DescendantsTreeMeta(id)` —— 单事件后代树（元数据视图）
- `DescendantsForest(ids)` —— 批量后代树（结构视图）
- `DescendantsForestMeta(ids)` —— 批量后代树（元数据视图）

这些方法为 HTTP API 的 `/descendants/{id}` 与 `POST /descendants` 端点提供底层支持。

## 函数说明

### `(*Store) descendantsTreeLocked`

```go
func (s *Store) descendantsTreeLocked(rootID uint64, visited map[uint64]struct{}) tree.DescendantsTree
```

递归构建后代树的内部方法（结构视图）。调用方**必须已持有 `s.mu` 锁**。

| 参数 | 类型 | 说明 |
|-----|------|------|
| `rootID` | `uint64` | 当前递归节点的事件 ID。 |
| `visited` | `map[uint64]struct{}` | 已访问节点集合，用于检测循环引用，避免无限递归。 |

**返回值**：`tree.DescendantsTree` —— 以 `rootID` 为根的子树。

**递归规则**：

- 若 `rootID` 已在 `visited` 中，返回 `tree.DescendantsTree{ID: rootID, IsRef: true, Children: nil}`，标记为引用节点，不再展开子节点。
- 否则将 `rootID` 加入 `visited`，创建新节点，`IsRef: false`，`Children` 初始为空数组。
- 遍历 `s.children[rootID]` 中的所有子 ID（通过 `sortedChildIDs` 排序），对每个子 ID 递归调用自身，将结果追加到 `Children` 中。

### `(*Store) descendantsTreeMetaLocked`

```go
func (s *Store) descendantsTreeMetaLocked(rootID uint64, visited map[uint64]struct{}) tree.DescendantsTreeMeta
```

递归构建后代树的内部方法（元数据视图）。逻辑与 `descendantsTreeLocked` 完全一致，但节点类型为 `tree.DescendantsTreeMeta`，额外携带 `TimeUnixNano`、`Type`、`Message`、`Payload` 等字段。

**注意**：即使节点被标记为 `IsRef: true`，元数据视图仍会填充该节点的元数据（从 `s.events[rootID]` 读取），以便前端在展示引用节点时仍有基本信息可用。

### `(*Store) DescendantsTree`

```go
func (s *Store) DescendantsTree(rootID uint64) (tree.DescendantsTree, error)
```

公开方法，查询单个事件的后代树（结构视图）。

**处理流程**：

1. 获取 `s.mu` 锁。
2. 调用 `validateRootIDLocked(rootID)` 校验根 ID 有效性。
3. 初始化新的 `visited` 映射。
4. 调用 `descendantsTreeLocked(rootID, visited)` 构建树。
5. 解锁并返回结果。

### `(*Store) DescendantsTreeMeta`

```go
func (s *Store) DescendantsTreeMeta(rootID uint64) (tree.DescendantsTreeMeta, error)
```

公开方法，查询单个事件的后代树（元数据视图）。流程与 `DescendantsTree` 相同，但调用 `descendantsTreeMetaLocked`。

### `(*Store) DescendantsForest`

```go
func (s *Store) DescendantsForest(rootIDs []uint64) ([]tree.DescendantsTree, error)
```

批量查询多个事件的后代树，返回一片“森林”（`[]tree.DescendantsTree`）。

**处理流程**：

1. 获取 `s.mu` 锁。
2. 调用 `validateRootIDsLocked(rootIDs)` 批量校验所有根 ID。
3. 对每个根 ID，独立初始化新的 `visited` 映射（各树之间的 visited 不共享，避免跨树截断）。
4. 调用 `descendantsTreeLocked` 构建每棵树，收集到结果切片中。
5. 解锁并返回森林。

### `(*Store) DescendantsForestMeta`

```go
func (s *Store) DescendantsForestMeta(rootIDs []uint64) ([]tree.DescendantsTreeMeta, error)
```

批量查询多个事件的后代树（元数据视图）。流程与 `DescendantsForest` 相同，但调用 `descendantsTreeMetaLocked`。

## 与其他文件的关系

| 依赖方向 | 文件/包 | 关系说明 |
|---------|--------|---------|
| 导入 | `internal/tree` | 使用 `tree.DescendantsTree`、`tree.DescendantsTreeMeta`、`tree.RootIDError`。 |
| 同包协作 | `internal/memory/store.go` | 读取 `Store.events`、`Store.children`。 |
| 同包协作 | `internal/memory/common.go` | 调用 `validateRootIDLocked`、`validateRootIDsLocked`、`sortedChildIDs`。 |
| 被调用 | `internal/httpapi/descendants.go` | HTTP Handler 调用 `store.DescendantsTree`、`DescendantsTreeMeta`、`DescendantsForest`、`DescendantsForestMeta`。 |

## 设计说明

- **独立 visited 映射**：在批量查询（Forest）中，每棵树使用独立的 `visited` 映射。这意味着若多棵树共享某些子树，每棵树都会完整展开这些子树，而不会在第一棵树中标记后导致后续树截断。这种设计保证了每棵返回的树都是自包含、完整的。
- **排序保证**：`sortedChildIDs` 确保子节点按 ID 升序排列，使得输出具有确定性，便于测试与前端展示。
- **IsRef 的语义**：`IsRef: true` 表示“该节点在当前遍历路径中已被访问过，为避免循环与重复展开，此处以引用形式出现”。它**不表示**该节点在全局 DAG 中是重复节点，而是表示在当前树的构建上下文中遇到了循环或多路径汇聚。
