# `provenance.go`

## 文件整体描述

`provenance.go` 是 **CelestialTree** 项目内存存储引擎中负责**溯源树构建**的实现文件，位于 `internal/memory` 包中。该文件实现了从指定事件出发，**向上**遍历所有父事件、父事件的父事件……直至根节点，构建完整溯源树的算法。

提供四种公开方法：

- `ProvenanceTree(id)` —— 单事件溯源树（结构视图）
- `ProvenanceTreeMeta(id)` —— 单事件溯源树（元数据视图）
- `ProvenanceForest(ids)` —— 批量溯源树（结构视图）
- `ProvenanceForestMeta(ids)` —— 批量溯源树（元数据视图）

这些方法为 HTTP API 的 `/provenance/{id}` 与 `POST /provenance` 端点提供底层支持。

## 函数说明

### `(*Store) provenanceTreeLocked`

```go
func (s *Store) provenanceTreeLocked(rootID uint64, visited map[uint64]struct{}) tree.ProvenanceTree
```

递归构建溯源树的内部方法（结构视图）。调用方**必须已持有 `s.mu` 锁**。

| 参数 | 类型 | 说明 |
|-----|------|------|
| `rootID` | `uint64` | 当前递归节点的事件 ID。 |
| `visited` | `map[uint64]struct{}` | 已访问节点集合，用于检测循环引用。 |

**返回值**：`tree.ProvenanceTree` —— 以 `rootID` 为根的溯源子树。

**递归规则**：

- 若 `rootID` 已在 `visited` 中，返回 `tree.ProvenanceTree{ID: rootID, IsRef: true, Parents: nil}`。
- 否则将 `rootID` 加入 `visited`，创建新节点，`IsRef: false`，`Parents` 初始为空数组。
- 读取 `s.events[rootID].Parents`，遍历所有父 ID：
  - 若父事件在 `s.events` 中不存在，跳过（防御性处理，容忍索引与主数据的不一致）。
  - 否则递归调用自身，将结果追加到 `Parents` 中。

### `(*Store) provenanceTreeMetaLocked`

```go
func (s *Store) provenanceTreeMetaLocked(rootID uint64, visited map[uint64]struct{}) tree.ProvenanceTreeMeta
```

递归构建溯源树的内部方法（元数据视图）。逻辑与 `provenanceTreeLocked` 一致，但节点类型为 `tree.ProvenanceTreeMeta`，额外携带完整元数据。即使 `IsRef: true`，也会填充元数据字段。

### `(*Store) ProvenanceTree`

```go
func (s *Store) ProvenanceTree(rootID uint64) (tree.ProvenanceTree, error)
```

公开方法，查询单个事件的溯源树（结构视图）。

**处理流程**：

1. 获取 `s.mu` 锁。
2. 调用 `validateRootIDLocked(rootID)` 校验根 ID。
3. 初始化新的 `visited` 映射。
4. 调用 `provenanceTreeLocked(rootID, visited)` 构建树。
5. 解锁并返回结果。

### `(*Store) ProvenanceTreeMeta`

```go
func (s *Store) ProvenanceTreeMeta(rootID uint64) (tree.ProvenanceTreeMeta, error)
```

公开方法，查询单个事件的溯源树（元数据视图）。

### `(*Store) ProvenanceForest`

```go
func (s *Store) ProvenanceForest(rootIDs []uint64) ([]tree.ProvenanceTree, error)
```

批量查询多个事件的溯源树，返回森林。

**处理流程**：

1. 获取 `s.mu` 锁。
2. 调用 `validateRootIDsLocked(rootIDs)` 批量校验。
3. 对每个根 ID，独立初始化新的 `visited` 映射。
4. 调用 `provenanceTreeLocked` 构建每棵树。
5. 解锁并返回森林。

### `(*Store) ProvenanceForestMeta`

```go
func (s *Store) ProvenanceForestMeta(rootIDs []uint64) ([]tree.ProvenanceTreeMeta, error)
```

批量查询多个事件的溯源树（元数据视图）。

## 与其他文件的关系

| 依赖方向 | 文件/包 | 关系说明 |
|---------|--------|---------|
| 导入 | `internal/tree` | 使用 `tree.ProvenanceTree`、`tree.ProvenanceTreeMeta`、`tree.RootIDError`。 |
| 同包协作 | `internal/memory/store.go` | 读取 `Store.events`。 |
| 同包协作 | `internal/memory/common.go` | 调用 `validateRootIDLocked`、`validateRootIDsLocked`。 |
| 被调用 | `internal/httpapi/provenance.go` | HTTP Handler 调用 `store.ProvenanceTree`、`ProvenanceTreeMeta`、`ProvenanceForest`、`ProvenanceForestMeta`。 |

## 设计说明

- **与后代树的对称性**：`provenance.go` 与 `descendants.go` 在代码结构、递归模式、视图分离策略上高度对称。这种对称性是有意为之，便于维护者一次理解、两处适用。
- **父事件缺失的宽容处理**：在 `provenanceTreeLocked` 中，若某父事件在 `s.events` 中不存在，选择**跳过**而非报错。这是因为在溯源场景中，历史数据可能部分缺失（如归档、清理），跳过可让遍历继续，返回不完整的但尽可能有用的树。相比之下，`descendantsTreeLocked` 不面临此问题（子事件缺失意味着索引损坏，但当前实现也未显式报错）。
- **Parents 顺序**：溯源树中 `Parents` 的顺序与 `tree.Event.Parents` 存储顺序一致，即事件写入时的原始顺序。
