# `common.go`

## 文件整体描述

`common.go` 是 **CelestialTree** 项目内存存储引擎中的**内部辅助函数**文件，位于 `internal/memory` 包中。该文件不包含任何对外暴露的 `Store` 方法，而是提供一组被本包其他文件复用的私有工具函数，包括根 ID 校验、批量根 ID 校验、事件 ID 有效性检查以及子 ID 排序。

这些辅助函数的职责单一且通用，抽离到独立文件后，使得 `descendants.go`、`provenance.go` 等核心业务文件更加聚焦于算法实现。

## 函数说明

### `(*Store) validateRootIDLocked`

```go
func (s *Store) validateRootIDLocked(id uint64) error
```

校验单个根事件 ID 是否有效。调用方**必须已持有 `s.mu` 锁**（由函数名中的 `Locked` 后缀提示）。

| 参数 | 类型 | 说明 |
|-----|------|------|
| `id` | `uint64` | 待校验的事件 ID。 |

**返回值**：`error` —— 若 ID 无效，返回 `*tree.RootIDError`；否则返回 `nil`。

**校验规则**：

1. `id == 0`：返回 `tree.RootIDError{ID: id, Reason: "id must be non-zero"}`。
2. `id` 对应的 events 槽位无效（通过 `isEventIDValid` 检查）：返回 `tree.RootIDError{ID: id, Reason: "event not found"}`。

### `(*Store) isEventIDValid`

```go
func (s *Store) isEventIDValid(id uint64) bool
```

检查 ID 对应的 `events` 稀疏 slice 槽位是否有效。调用方**必须已持有 `s.mu` 锁**。

**校验规则**：

1. `id >= uint64(len(s.events))`：越界，返回 `false`。
2. `s.events[id].ID == 0`：槽位为零值（空洞），返回 `false`。
3. 否则返回 `true`。

**设计说明**：由于 `events` 从 `map[uint64]tree.Event` 改为了稀疏 `[]tree.Event`，原有的 map 存在性检查（`_, ok := s.events[id]`）不再适用。`isEventIDValid` 统一封装了边界检查与零值检测逻辑，被所有需要校验事件存在性的方法调用。

### `(*Store) validateRootIDsLocked`

```go
func (s *Store) validateRootIDsLocked(rootIDs []uint64) error
```

批量校验一组根事件 ID 是否全部有效。调用方**必须已持有 `s.mu` 锁**。

| 参数 | 类型 | 说明 |
|-----|------|------|
| `rootIDs` | `[]uint64` | 待校验的 ID 列表。 |

**返回值**：`error` —— 若任一 ID 无效，立即返回对应的 `*tree.RootIDError`；全部有效则返回 `nil`。

**实现细节**：遍历 `rootIDs`，对每一项调用 `validateRootIDLocked`，遇到第一个错误即提前返回。

### `sortedChildIDs`

```go
func sortedChildIDs(sli []uint64) []uint64
```

返回子事件 ID 列表的排序副本，不修改原 slice。

| 参数 | 类型 | 说明 |
|-----|------|------|
| `sli` | `[]uint64` | 子事件 ID 列表。 |

**返回值**：`[]uint64` —— 升序排序后的 ID 切片副本。若列表为空，返回空数组 `[]uint64{}`（非 `nil`）。

**实现细节**：

1. 若 `len(sli) == 0`，直接返回空数组。
2. 预分配容量为 `len(sli)` 的切片，`copy` 原始数据。
3. 调用 `slices.Sort` 升序排序后返回。

## 与其他文件的关系

| 依赖方向 | 文件/包 | 关系说明 |
|---------|--------|---------|
| 导入 | `internal/tree` | 使用 `tree.RootIDError` 构造校验错误。 |
| 标准库 | `slices` | `sortedChildIDs` 中使用 `slices.Sort` 排序。 |
| 被调用 | `internal/memory/descendants.go` | `DescendantsTree`、`DescendantsTreeMeta` 等在锁保护下调用 `validateRootIDLocked` 与 `sortedChildIDs`。 |
| 被调用 | `internal/memory/provenance.go` | `ProvenanceTree`、`ProvenanceTreeMeta` 等在锁保护下调用 `validateRootIDLocked`。 |
| 被调用 | `internal/memory/emit.go` | `Emit` 在校验父事件存在性时调用 `isEventIDValid`。 |
| 被调用 | `internal/memory/event.go` | `Get` 调用 `isEventIDValid` 检查事件存在性。 |
| 被调用 | `internal/memory/graph.go` | `Children`、`Ancestors` 调用 `isEventIDValid` 检查事件存在性。 |

## 设计说明

- **Locked 命名约定**：函数名中包含 `Locked` 是为了明确提示调用者必须在调用前获取锁。这是 Go 并发编程中常见的自文档化手法，虽然编译器不会强制检查，但能显著降低误用风险。
- **空集合返回空数组**：`sortedChildIDs` 对空列表返回 `[]uint64{}` 而非 `nil`。这使得 JSON 序列化后输出 `[]` 而不是 `null`，对前端更友好，也避免了 `nil` 切片在 JSON 中语义不同的问题。
