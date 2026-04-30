# CelestialTree

**CelestialTree** 是一个用于记录、存储和查询 **事件因果关系（Causal Event DAG）** 的轻量级服务。

它的核心目标是：
👉 **提供可靠的"事件血缘 / 溯源 / 影响分析"能力。**

CelestialTree 专注于 **"发生了什么，以及它是由什么引起的"**，而不是任务本身如何执行。

## 设计动机

在复杂系统中，我们经常会遇到这些问题：

* 一个任务失败，**它是由哪个上游事件引起的？**
* 某个输入变化，**会影响到哪些下游结果？**
* 一个 DAG 执行完成后，**如何重建完整的执行因果链？**
* 如何把"日志"升级为**结构化、可查询、可回放的事件历史？**

CelestialTree 的答案是：
**把所有重要状态变化建模为事件，并显式记录事件之间的父子关系。**

最终形成一棵（或多棵）**有向无环事件树（DAG）**。

## 核心特性

* 🌳 **事件因果 DAG 存储**

  * 每个事件可以有 0～N 个父事件
  * 自动形成可回溯、可下钻的因果结构

* 🚀 **高性能事件写入**

  * 事件写入为追加式（append-only）
  * 适合高频任务系统埋点

* 🔍 **血缘与影响分析**

  * 查询某事件的所有祖先（provenance）
  * 查询某事件的所有后代（descendants）

* 🌐 **多协议接口**

  * HTTP REST API
  * Server-Sent Events（SSE）实时事件流
  * gRPC / Protobuf

* 🐍 **Python Client**

  * 可直接嵌入 CelestialFlow 等任务系统
  * 无侵入式记录任务生命周期

## 适用场景

CelestialTree 特别适合以下场景：

* DAG / Workflow / Pipeline 系统
* 分布式任务调度与执行框架
* 数据处理与 ETL 血缘追踪
* AI / ML Pipeline 训练与推理溯源
* 复杂系统运行态调试与回放

## 核心概念

### 事件（Event）

一个事件代表系统中一次**不可变的事实**，例如：

* `task.created`
* `task.started`
* `task.success`
* `task.failed`
* `stage.split`
* `router.dispatch`

事件包含：

* `id`：事件唯一 ID（全局递增）
* `type`：事件类型（必填）
* `message`：人类可读描述
* `payload`：结构化数据（JSON）
* `parents`：父事件 ID 列表
* `time_unix_nano`：事件发生时间（纳秒级 Unix 时间戳）

### 因果关系（Parents）

事件之间通过 `parents` 建立因果关系：

```text
A ──▶ B ──▶ C
 \          ▲
  ─────▶ D ─┘
```

这不是一条简单链路，而是一个 **DAG**：

* 一个事件可以由多个父事件触发
* 一个事件也可以触发多个后续事件

### 关键拓扑概念

* **Root（根）**：无父事件的事件，代表因果链的起点
* **Head（头/叶子）**：无子事件的事件，代表 DAG 的当前前沿

## 快速开始

### 启动服务

```bash
go run cmd/celestialtree/main.go
# or
make run
```

默认启动：
* HTTP 服务：`http://localhost:7777`
* gRPC 服务：`grpc://localhost:7778`

启动时会自动写入一个 `genesis` 创世事件作为 DAG 的根节点。

### 写入事件（curl）

```bash
curl -X POST http://localhost:7777/emit \
  -H "Content-Type: application/json" \
  -d '{
    "type": "task.success",
    "message": "Task completed",
    "parents": [1],
    "payload": {"task_id": "A-001", "duration": 1.23}
  }'
```

返回：
```json
{"id": 2}
```

### 查询事件

```bash
# 查询单个事件
curl http://localhost:7777/event/2

# 查询溯源树（祖先）
curl http://localhost:7777/provenance/2

# 查询后代树
curl http://localhost:7777/descendants/2

# 查询运行时快照
curl http://localhost:7777/snapshot
# 返回示例：
# {"ts":1713709263,"goroutines":7,"edges":1480,"roots":1,"heads":43,"subscribers":5,"next_event_id":1524}
```

### 实时订阅（SSE）

```bash
curl -N http://localhost:7777/subscribe
```

或使用 JavaScript `EventSource`：

```javascript
const es = new EventSource('http://localhost:7777/subscribe');
es.addEventListener('emit', (e) => {
  console.log('New event:', JSON.parse(e.data));
});
```

### 使用 Python Client

```python
from celestialtree import Client

client = Client(base_url="http://localhost:7777")

event_id = client.emit(
    event_type="task.success",
    parents=[1],
    message="Task completed successfully",
    payload={
        "task_id": "A-001",
        "duration": 1.23
    }
)

print(event_id)

# 查询血缘
tree = client.provenance(event_id)
desc = client.descendants(event_id)
```

## HTTP API 概览

| 方法 | 接口 | 说明 |
|------|------|------|
| `POST` | `/emit` | 写入新事件 |
| `GET` | `/event/{id}` | 查询单个事件详情 |
| `GET` | `/children/{id}` | 查询某事件的直接子事件 |
| `GET` | `/ancestors/{id}` | 查询某事件的所有根祖先 |
| `GET` | `/descendants/{id}?view=struct\|meta` | 查询后代树 |
| `POST` | `/descendants` | 批量查询后代树（森林） |
| `GET` | `/provenance/{id}?view=struct\|meta` | 查询溯源树 |
| `POST` | `/provenance` | 批量查询溯源树（森林） |
| `GET` | `/heads` | 查询当前所有 Head（叶子事件） |
| `GET` | `/roots` | 查询当前所有 Root（创世事件） |
| `GET` | `/snapshot` | 查询运行时统计快照 |
| `GET` | `/subscribe` | SSE 实时事件流订阅 |
| `GET` | `/healthz` | 健康检查 |
| `GET` | `/version` | 查询应用版本信息 |

### 视图参数（View）

`/descendants` 与 `/provenance` 接口支持 `view` 参数：

* **`struct`**（默认）：仅返回事件 ID 与树形骨架，体积最小
* **`meta`**：额外携带时间戳、类型、消息、载荷等完整元数据

### 批量查询

`POST /descendants` 与 `POST /provenance` 支持一次查询多棵树的森林：

```json
{
  "ids": [42, 43, 44],
  "view": "meta"
}
```

## gRPC API

服务定义位于 `proto/celestialtree.proto`，当前暴露接口：

| RPC | 请求 | 响应 | 说明 |
|-----|------|------|------|
| `Emit` | `EmitRequest` | `EmitResponse` | 写入新事件 |

使用 `grpcurl` 调试（需开启 reflection）：

```bash
grpcurl -plaintext -d '{
  "type": "task.success",
  "message": "done",
  "parents": [1]
}' localhost:7778 celestialtree.v1.CelestialTreeService/Emit
```

## 项目结构

```text
CelestialTree/
├── cmd/
│   ├── celestialtree/    # 服务主入口（HTTP + gRPC 双协议）
│   └── now/              # 小工具：输出当前 UTC 时间
├── internal/
│   ├── tree/             # 核心数据模型（Event、树结构、错误类型）
│   ├── memory/           # 内存存储引擎（稀疏 slice + DAG 索引 + SSE 广播）
│   ├── httpapi/          # HTTP REST API 处理器
│   ├── grpcapi/          # gRPC API 实现
│   └── version/          # 版本信息（编译期注入）
├── proto/                # Protobuf 定义与生成代码
├── bench/                # 性能基准测试工具（HTTP + gRPC，Go 实现）
├── bin/                  # 编译产物
├── docs/
│   ├── internal/         # 内部模块详细文档（与源码一一对应）
│   ├── cmd/              # cmd 模块文档
│   └── bench/            # 基准测试工具文档
└── README.md
```

### 内部模块文档

`docs/internal/` 下为每个 Go 源文件建立了对应的 Markdown 说明文档，包含文件职责、所有函数/结构体说明、与其他文件的关系、设计决策等。详见 [`docs/internal/README.md`](docs/internal/README.md)。

核心模块速查：

| 模块 | 关键文件 | 职责 |
|------|---------|------|
| `internal/tree` | [`types.go`](docs/internal/tree/types.md) | 全系统共享的数据契约 |
| `internal/memory` | [`store.go`](docs/internal/memory/store.md) | 内存 DAG 存储引擎 |
| `internal/memory` | [`emit.go`](docs/internal/memory/emit.md) | 事件写入与拓扑维护 |
| `internal/httpapi` | [`routes.go`](docs/internal/httpapi/routes.md) | HTTP 路由注册中心 |
| `internal/grpcapi` | [`server.go`](docs/internal/grpcapi/server.md) | gRPC 服务入口 |

## 与 CelestialFlow 的关系

* **CelestialFlow**：任务如何执行
* **CelestialTree**：任务为何如此执行

CelestialFlow 中的每个 Task / Stage / Node
都可以将关键状态变化 **emit** 到 CelestialTree，
从而获得完整的执行因果历史。

两者解耦，但天然互补。

## 设计原则

* **事件不可变**：事件一旦写入不可修改、不可删除，保证历史可信
* **因果显式化**：父子关系由调用方显式声明，而非隐式推断
* **写入简单、查询强大**：写入只需 `type` + `parents`，查询支持单条、树形、批量、流式
* **高性能内存优化**：事件以稀疏 slice 存储（ID 即下标），父子索引使用紧凑的 `[]uint64`，最大限度降低大规模场景下的内存开销
* **不绑定具体任务系统**：纯事件语义，可接入任何产生状态变化的系统
* **可作为基础设施长期运行**：轻量、无外部依赖、单二进制可部署

## 未来规划（非承诺）

* [ ] 存储后端抽象（内存 / RocksDB / Redis / SQLite）
* [ ] 事件快照与压缩
* [ ] 更强的图查询能力（路径搜索、子图匹配）
* [ ] 官方前端可视化 UI
* [ ] 事件过滤与索引（按 type、time、payload 字段查询）
* [ ] 集群模式与事件复制

## Star 历史趋势（Star History）

如果对项目感兴趣的话，欢迎 star。如果有问题或者建议的话，欢迎提交 [Issues](https://github.com/Mr-xiaotian/CelestialTree/issues) 或者在 [Discussion](https://github.com/Mr-xiaotian/CelestialTree/discussions) 中告诉我。

[![Star History Chart](https://api.star-history.com/svg?repos=Mr-xiaotian/CelestialTree&type=Date)](https://star-history.com/#Mr-xiaotian/CelestialTree&Date)

## 许可（License）

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 作者（Author）

Author: Mr-xiaotian  
Email: mingxiaomingtian@gmail.com  
Project Link: [https://github.com/Mr-xiaotian/CelestialTree](https://github.com/Mr-xiaotian/CelestialTree)
