# CelestialTree

**CelestialTree** 是一个用于记录、存储和查询 **事件因果关系（Causal Event DAG）** 的轻量级服务。

它的核心目标是：
👉 **提供可靠的“事件血缘 / 溯源 / 影响分析”能力。**

CelestialTree 专注于 **“发生了什么，以及它是由什么引起的”**，而不是任务本身如何执行。

## 设计动机

在复杂系统中，我们经常会遇到这些问题：

* 一个任务失败，**它是由哪个上游事件引起的？**
* 某个输入变化，**会影响到哪些下游结果？**
* 一个 DAG 执行完成后，**如何重建完整的执行因果链？**
* 如何把“日志”升级为**结构化、可查询、可回放的事件历史？**

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

* 🌐 **多接口支持**

  * HTTP API
  * Server-Sent Events（SSE）事件流
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

* `id`：事件唯一 ID
* `type`：事件类型
* `message`：人类可读描述
* `payload`：结构化数据（JSON）
* `parents`：父事件 ID 列表
* `timestamp`：事件发生时间

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

## 快速开始

### 启动服务

```bash
go run cmd/celestialtree/main.go
// or
make run
```

默认会启动 HTTP 服务与 SSE 接口。

### 使用 Python Client 写入事件

```python
from celestialtree import Client

client = Client(base_url="http://localhost:7777")

event_id = client.emit(
    event_type="task.success",
    parents=[123456],
    message="Task completed successfully",
    payload={
        "task_id": "A-001",
        "duration": 1.23
    }
)

print(event_id)
```

### 查询事件血缘

```python
tree = client.provenance(event_id)
```

```python
desc = client.descendants(event_id)
```

返回结果为结构化树，可直接用于：

* UI 可视化
* 调试分析
* 执行回放

## API 概览（HTTP）

| 接口                      | 说明        |
| ----------------------- | --------- |
| `POST /emit`            | 写入事件      |
| `GET /provenance/{id}`  | 查询祖先      |
| `GET /descendants/{id}` | 查询后代      |
| `GET /events/{id}`      | 查询事件详情    |
| `GET /stream`           | SSE 实时事件流 |
| `GET /health`           | 健康检查      |

## 与 CelestialFlow 的关系

* **CelestialFlow**：任务如何执行
* **CelestialTree**：任务为何如此执行

CelestialFlow 中的每个 Task / Stage / Node
都可以将关键状态变化 **emit** 到 CelestialTree，
从而获得完整的执行因果历史。

两者解耦，但天然互补。

## 项目结构（简述）

```text
cmd/
  celestialtree/    # 服务入口
internal/tree/      # 事件存储与 DAG 逻辑
internal/httpapi/   # HTTP API
internal/grpcapi/   # gRPC API（可选）
proto/              # Protobuf 定义
```

## 设计原则

* **事件不可变**
* **因果显式化**
* **写入简单、查询强大**
* **不绑定具体任务系统**
* **可作为基础设施长期运行**

## 未来规划（非承诺）

* gRPC + Protobuf 原生接口
* 存储后端抽象（内存 / RocksDB / Redis / SQLite）
* 事件快照与压缩
* 更强的图查询能力
* 官方前端可视化 UI

## 文件结构（File Structure）
```
📁 CelestialTree	(75MB 436KB 809B)
    📁 bench   	(40KB 25B)
        📁 grpc	(2KB 716B)
            🌀 emit.go	(2KB 716B)
        📁 http	(2KB 482B)
            🌀 emit.go	(2KB 482B)
        🐍 bench_celestialtree.py	(23KB 969B)
        🐍 bench_emit.py         	(7KB 912B)
        🐍 bench_redis_emit.py   	(3KB 18B)
    📁 bin     	(42MB 610KB)
        ❓ bench_emit_grpc	(14MB 775KB 512B)
        ❓ bench_emit_http	(8MB 148KB)
        ❓ celestialtree  	(17MB 382KB 512B)
        ❓ now            	(2MB 328KB)
    📁 cmd     	(3KB 246B)
        📁 celestialtree	(3KB 129B)
            🌀 main.go	(3KB 129B)
        📁 now          	(117B)
            🌀 main.go	(117B)
    📁 docs    	(0B)
    📁 internal	(26KB 710B)
        📁 grpcapi	(1KB 519B)
            🌀 emit.go  	(1KB 255B)
            🌀 server.go	(264B)
        📁 httpapi	(10KB 615B)
            🌀 common.go     	(1KB 39B)
            🌀 descendants.go	(2KB 221B)
            🌀 emit.go       	(677B)
            🌀 event.go      	(525B)
            🌀 graph.go      	(1KB 182B)
            🌀 health.go     	(547B)
            🌀 provenance.go 	(2KB 214B)
            🌀 routes.go     	(988B)
            🌀 sse.go        	(1KB 294B)
        📁 memory 	(11KB 273B)
            🌀 common.go     	(798B)
            🌀 descendants.go	(2KB 934B)
            🌀 emit.go       	(1KB 539B)
            🌀 event.go      	(193B)
            🌀 graph.go      	(1KB 451B)
            🌀 provenance.go 	(2KB 923B)
            🌀 sse.go        	(728B)
            🌀 store.go      	(827B)
        📁 tools  	(0B)
        📁 tree   	(3KB 200B)
            🌀 types.go	(3KB 200B)
        📁 version	(127B)
            🌀 version.go	(127B)
    📁 proto   	(11KB 388B)
        🌀 celestialtree.pb.go     	(6KB 364B)
        ❓ celestialtree.proto     	(418B)
        🌀 celestialtree_grpc.pb.go	(4KB 630B)
    📁 temp    	(5KB 673B)
        📝 protocols.md	(5KB 673B)
    📁 [2项排除的目录]	(32MB 755KB 973B)
    ❓ .gitignore    	(42B)
    ❓ go.mod        	(330B)
    ❓ go.sum        	(3KB 50B)
    ❓ Makefile      	(1KB 792B)
    📝 README.md     	(0B)
    📓 _preview.ipynb	(2KB 676B)
```

## Star 历史趋势（Star History）

如果对项目感兴趣的话，欢迎star。如果有问题或者建议的话, 欢迎提交[Issues](https://github.com/Mr-xiaotian/CelestialTree/issues)或者在[Discussion](https://github.com/Mr-xiaotian/CelestialTree/discussions)中告诉我。

[![Star History Chart](https://api.star-history.com/svg?repos=Mr-xiaotian/CelestialTree&type=Date)](https://star-history.com/#Mr-xiaotian/CelestialTree&Date)

## 许可（License）
This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 作者（Author）
Author: Mr-xiaotian
Email: mingxiaomingtian@gmail.com
Project Link: [https://github.com/Mr-xiaotian/CelestialTree](https://github.com/Mr-xiaotian/CelestialTree)
