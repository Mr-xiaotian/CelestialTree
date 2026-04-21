# `version.go`

## 文件整体描述

`version.go` 是 **CelestialTree** 项目的版本信息源文件，位于 `internal/version` 包中。该文件仅包含一组在编译期通过 `-ldflags` 注入的变量默认值，用于在运行时暴露应用的元数据（名称、版本号、Git Commit、构建时间）。

这是一个高度稳定、几乎不会变更的模块，职责单一：为系统其他部分（HTTP API、启动日志）提供统一的版本信息入口。

## 变量说明

### `Name`

```go
var Name = "CelestialTree"
```

应用名称。在 `cmd/celestialtree/main.go` 的启动日志与 `internal/httpapi/health.go` 的 `/version` 接口中被引用。

### `Version`

```go
var Version = "dev"
```

当前应用版本号。默认值为 `"dev"`，在生产构建时通常通过 `-ldflags "-X github.com/Mr-xiaotian/CelestialTree/internal/version.Version=v1.x.x"` 注入正式版本号。

### `GitCommit`

```go
var GitCommit = "unknown"
```

构建时对应的 Git Commit SHA。默认 `"unknown"`，同样通过 `-ldflags` 注入，用于问题追踪与版本回溯。

### `BuildTime`

```go
var BuildTime = "unknown"
```

构建时间戳。默认 `"unknown"`，通过 `-ldflags` 注入，格式通常由构建脚本决定（如 RFC3339）。

## 与其他文件的关系

| 依赖方向 | 文件/包 | 关系说明 |
|---------|--------|---------|
| 被导入 | `cmd/celestialtree/main.go` | 启动时打印日志：`log.Printf("%s %s(%s) built at %s", version.Name, version.Version, ...)`。 |
| 被导入 | `internal/httpapi/health.go` | `/version` HTTP 接口返回 `version.Name`、`version.Version`、`version.GitCommit`、`version.BuildTime` 的 JSON 对象。 |
| 无内部依赖 | — | 本包不依赖项目内任何其他包，也不依赖任何第三方库。 |

## 构建注入示例

在 `Makefile` 或 CI 流水线中，通常使用如下方式覆盖默认值：

```bash
go build -ldflags "\
  -X github.com/Mr-xiaotian/CelestialTree/internal/version.Version=$(VERSION) \
  -X github.com/Mr-xiaotian/CelestialTree/internal/version.GitCommit=$(COMMIT) \
  -X github.com/Mr-xiaotian/CelestialTree/internal/version.BuildTime=$(BUILD_TIME) \
" -o bin/celestialtree cmd/celestialtree/main.go
```
