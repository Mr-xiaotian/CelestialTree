# `routes.go`

## 文件整体描述

`routes.go` 是 **CelestialTree** 项目 HTTP API 的**路由注册中心**，位于 `internal/httpapi` 包中。该文件唯一的职责是将 URL 路径与对应的 HTTP Handler 函数绑定到传入的 `http.ServeMux` 上。它是 HTTP 层对外暴露接口的“清单”，任何新增、删除或修改公开端点的操作，都应集中在此文件中体现。

通过将路由注册逻辑独立出来，`cmd/celestialtree/main.go` 只需调用 `httpapi.RegisterRoutes(mux, store)` 即可完成所有 HTTP 端点的挂载，无需感知具体有哪些 Handler。

## 函数说明

### `RegisterRoutes`

```go
func RegisterRoutes(mux *http.ServeMux, store *memory.Store)
```

将 CelestialTree 的所有 HTTP API 端点注册到给定的 `http.ServeMux` 上。

| 参数 | 类型 | 说明 |
|-----|------|------|
| `mux` | `*http.ServeMux` | Go 标准库的多路复用器，所有路由规则都会注册到此对象上。 |
| `store` | `*memory.Store` | 内存存储实例，通过闭包方式注入到各个 Handler 中。 |

**注册的端点清单**：

| 路径 | Handler | 方法 | 说明 |
|------|---------|------|------|
| `/emit` | `handleEmit(store)` | POST | 写入新事件。 |
| `/event/` | `handleGetEvent(store)` | GET | 根据 ID 查询单个事件。 |
| `/children/` | `handleChildren(store)` | GET | 查询某事件的直接子事件列表。 |
| `/ancestors/` | `handleAncestors(store)` | GET | 查询某事件的所有根祖先。 |
| `/heads` | `handleHeads(store)` | GET | 查询当前所有 Head（无子节点的叶子事件）。 |
| `/roots` | `handleRoots(store)` | GET | 查询当前所有 Root（无父事件的创世事件）。 |
| `/snapshot` | `handleSnapshot(store)` | GET | 查询存储层运行时统计快照。 |
| `/healthz` | `handleHealthz()` | GET | 健康检查端点。 |
| `/version` | `handleVersion()` | GET | 查询应用版本信息。 |
| `/descendants/` | `handleDescendants(store)` | GET | 查询某事件的后代树（支持 `?view=` 参数）。 |
| `/descendants` | `handleDescendantsBatch(store)` | POST | 批量查询多个事件的后代树。 |
| `/provenance/` | `handleProvenance(store)` | GET | 查询某事件的溯源树（支持 `?view=` 参数）。 |
| `/provenance` | `handleProvenanceBatch(store)` | POST | 批量查询多个事件的溯源树。 |
| `/subscribe` | `handleSubscribe(store)` | GET | SSE 长连接订阅新事件流。 |

**路由设计说明**：

- `/descendants/`（带斜杠）与 `/descendants`（不带斜杠）分别对应单条查询与批量查询；`http.ServeMux` 按最长前缀匹配，因此两者不会冲突。
- `/provenance/` 与 `/provenance` 同理。
- `/subscribe` 使用 SSE（Server-Sent Events）协议，而非 WebSocket，降低实现复杂度。

## 与其他文件的关系

| 依赖方向 | 文件/包 | 关系说明 |
|---------|--------|---------|
| 导入 | `internal/memory` | 将 `*memory.Store` 注入到所有需要访问存储的 Handler 中。 |
| 被调用 | `cmd/celestialtree/main.go` | `main.go` 创建 `http.NewServeMux()` 后调用 `httpapi.RegisterRoutes(mux, store)`，随后将 `mux` 作为 `http.Server.Handler` 启动服务。 |
| 同包协作 | `internal/httpapi/emit.go` | 调用 `handleEmit(store)`。 |
| 同包协作 | `internal/httpapi/event.go` | 调用 `handleGetEvent(store)`。 |
| 同包协作 | `internal/httpapi/graph.go` | 调用 `handleChildren(store)`、`handleAncestors(store)`、`handleHeads(store)`、`handleRoots(store)`。 |
| 同包协作 | `internal/httpapi/descendants.go` | 调用 `handleDescendants(store)`、`handleDescendantsBatch(store)`。 |
| 同包协作 | `internal/httpapi/provenance.go` | 调用 `handleProvenance(store)`、`handleProvenanceBatch(store)`。 |
| 同包协作 | `internal/httpapi/snapshot.go` | 调用 `handleSnapshot(store)`。 |
| 同包协作 | `internal/httpapi/health.go` | 调用 `handleHealthz()`、`handleVersion()`。 |
| 同包协作 | `internal/httpapi/sse.go` | 调用 `handleSubscribe(store)`。 |

## 扩展建议

若需新增公共 HTTP 端点：

1. 在 `internal/httpapi/` 下新建文件（如 `search.go`）并实现 `handleXxx(store *memory.Store) http.HandlerFunc`；
2. 在 `routes.go` 的 `RegisterRoutes` 中新增一行 `mux.HandleFunc("/xxx", handleXxx(store))`；
3. 确保在 `internal/memory/` 中已提供对应存储能力。
