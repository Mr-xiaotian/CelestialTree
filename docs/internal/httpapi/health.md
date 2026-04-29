# `health.go`

## 文件整体描述

`health.go` 是 **CelestialTree** 项目 HTTP API 中负责**健康检查与元数据暴露**的处理器文件，位于 `internal/httpapi` 包中。该文件实现了两个运维/诊断端点：

- `/healthz` —— 健康检查
- `/version` —— 版本信息查询

这两个端点通常被负载均衡器、Kubernetes Probe、监控系统等外部基础设施调用，不直接参与业务逻辑，但对生产环境的可观测性与稳定性至关重要。

## 函数说明

### `handleHealthz`

```go
func handleHealthz() http.HandlerFunc
```

返回 `/healthz` 端点的 Handler。该端点用于快速判断服务是否存活且可响应请求。

**Handler 内部逻辑**：

- 接受任意 HTTP 方法（通常由调用方使用 `GET`）。
- 直接返回 `200 OK`，响应体为 JSON：

```json
{"ok": true, "ts": 1713709263}
```

- `ok` 字段固定为 `true`，表示服务处于健康状态。
- `ts` 字段为当前 Unix 时间戳（秒级），便于调用方检测时钟漂移或响应新鲜度。

**注意**：当前实现为“乐观健康检查”，即只要 HTTP 层能响应，即认为健康。未来若引入外部依赖（如持久化数据库、消息队列），可在此 Handler 中增加对依赖状态的探测，并视情况返回 `503 Service Unavailable`。

### `handleVersion`

```go
func handleVersion() http.HandlerFunc
```

返回 `/version` 端点的 Handler。该端点暴露当前运行实例的版本元数据，便于排查问题、确认部署是否生效。

**Handler 内部逻辑**：

- 接受任意 HTTP 方法（通常由调用方使用 `GET`）。
- 返回 `200 OK`，响应体为 JSON：

```json
{
  "name": "CelestialTree",
  "version": "dev",
  "commit": "unknown",
  "build": "unknown"
}
```

- `name`、`version`、`commit`、`build` 的值均来自 `internal/version` 包中的变量。在生产构建时，这些值会被 `-ldflags` 注入为实际版本信息。

## 与其他文件的关系

| 依赖方向 | 文件/包 | 关系说明 |
|---------|--------|---------|
| 导入 | `internal/version` | 读取 `version.Name`、`version.Version`、`version.GitCommit`、`version.BuildTime` 构造 `/version` 响应。 |
| 同包协作 | `internal/httpapi/common.go` | 调用 `writeJSON` 统一写入 JSON 响应。 |
| 同包协作 | `internal/httpapi/routes.go` | `RegisterRoutes` 中将 `/healthz` 与 `/version` 注册到对应 Handler。 |
| 标准库 | `time` | `handleHealthz` 使用 `time.Now().Unix()` 生成时间戳。 |

## 运维建议

- **Kubernetes 配置**：建议将 `/healthz` 同时配置为 `livenessProbe` 与 `readinessProbe`：

```yaml
livenessProbe:
  httpGet:
    path: /healthz
    port: 7777
  initialDelaySeconds: 5
  periodSeconds: 10
readinessProbe:
  httpGet:
    path: /healthz
    port: 7777
  initialDelaySeconds: 2
  periodSeconds: 5
```

- **构建注入**：在 CI/CD 流水线中，使用 `-ldflags` 注入真实版本信息，确保 `/version` 返回的数据可用于问题回溯：

```bash
go build -ldflags "\
  -X github.com/Mr-xiaotian/CelestialTree/internal/version.Version=v1.2.3 \
  -X github.com/Mr-xiaotian/CelestialTree/internal/version.GitCommit=$(git rev-parse --short HEAD) \
  -X github.com/Mr-xiaotian/CelestialTree/internal/version.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ) \
" -o bin/celestialtree cmd/celestialtree/main.go
```
