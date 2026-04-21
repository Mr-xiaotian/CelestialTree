# CelestialTree 内部模块文档索引

本文档为 `internal/` 目录下所有 Go 源文件的对应说明文档索引，按包（package）组织。

---

## `internal/tree` — 核心数据模型

| 源文件 | 文档 | 说明 |
|--------|------|------|
| `types.go` | [types.md](tree/types.md) | 定义 Event、请求/响应体、树形结构、错误类型等全部核心数据结构。 |

---

## `internal/version` — 版本信息

| 源文件 | 文档 | 说明 |
|--------|------|------|
| `version.go` | [version.md](version/version.md) | 应用名称、版本号、Git Commit、构建时间等编译期注入变量。 |

---

## `internal/memory` — 内存存储引擎

| 源文件 | 文档 | 说明 |
|--------|------|------|
| `store.go` | [store.md](memory/store.md) | `Store` 结构体定义与构造函数，系统的单一事实来源。 |
| `emit.go` | [emit.md](memory/emit.md) | 事件写入（`Emit`），DAG 拓扑维护与索引更新。 |
| `event.go` | [event.md](memory/event.md) | 单事件精确查询（`Get`）。 |
| `graph.go` | [graph.md](memory/graph.md) | 图拓扑查询：`Children`、`Ancestors`、`Heads`、`Roots`。 |
| `descendants.go` | [descendants.md](memory/descendants.md) | 后代树构建：单条/批量、结构/元数据四种视图。 |
| `provenance.go` | [provenance.md](memory/provenance.md) | 溯源树构建：单条/批量、结构/元数据四种视图。 |
| `snapshot.go` | [snapshot.md](memory/snapshot.md) | 运行时统计快照采集。 |
| `sse.go` | [sse.md](memory/sse.md) | SSE 订阅者管理与事件广播机制。 |
| `common.go` | [common.md](memory/common.md) | 内部辅助函数：根 ID 校验、子 ID 排序等。 |

---

## `internal/httpapi` — HTTP API 层

| 源文件 | 文档 | 说明 |
|--------|------|------|
| `routes.go` | [routes.md](httpapi/routes.md) | 路由注册中心，所有 HTTP 端点的统一挂载点。 |
| `common.go` | [common.md](httpapi/common.md) | 通用 HTTP 工具函数：方法校验、路径解析、JSON 读写等。 |
| `emit.go` | [emit.md](httpapi/emit.md) | `/emit` 端点，事件写入 Handler。 |
| `event.go` | [event.md](httpapi/event.md) | `/event/{id}` 端点，单事件查询 Handler。 |
| `graph.go` | [graph.md](httpapi/graph.md) | `/children/`、`/ancestors/`、`/heads`、`/roots` 端点。 |
| `descendants.go` | [descendants.md](httpapi/descendants.md) | `/descendants/{id}` 与 `POST /descendants` 端点。 |
| `provenance.go` | [provenance.md](httpapi/provenance.md) | `/provenance/{id}` 与 `POST /provenance` 端点。 |
| `snapshot.go` | [snapshot.md](httpapi/snapshot.md) | `/snapshot` 端点，运行时快照查询。 |
| `health.go` | [health.md](httpapi/health.md) | `/healthz` 与 `/version` 运维端点。 |
| `sse.go` | [sse.md](httpapi/sse.md) | `/subscribe` 端点，SSE 长连接订阅 Handler。 |

---

## `internal/grpcapi` — gRPC API 层

| 源文件 | 文档 | 说明 |
|--------|------|------|
| `server.go` | [server.md](grpcapi/server.md) | gRPC 服务结构体 `Server` 定义与构造函数。 |
| `emit.go` | [emit.md](grpcapi/emit.md) | gRPC `Emit` RPC 实现，Protobuf 与内部类型的协议转换。 |

---

## 架构概览

```
┌─────────────────────────────────────────────────────────────┐
│                        协议入口层                             │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐  │
│  │  HTTP API    │  │  gRPC API    │  │  SSE /subscribe  │  │
│  │  httpapi/*   │  │  grpcapi/*   │  │  httpapi/sse.go  │  │
│  └──────┬───────┘  └──────┬───────┘  └────────┬─────────┘  │
└─────────┼─────────────────┼───────────────────┼────────────┘
          │                 │                   │
          └─────────────────┴───────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                      核心存储引擎                             │
│                    internal/memory/*                          │
│                      (Store — 内存 DAG)                       │
└─────────────────────────────────────────────────────────────┘
                              │
          ┌───────────────────┼───────────────────┐
          ▼                   ▼                   ▼
┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐
│  tree/types.go  │  │ version/version │  │   proto/*.pb.go │
│  核心数据模型    │  │   版本信息       │  │  Protobuf 生成  │
└─────────────────┘  └─────────────────┘  └─────────────────┘
```
