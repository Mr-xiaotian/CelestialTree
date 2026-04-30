# `main.go`

## 文件整体描述

`main.go` 是一个极简的命令行工具，位于 `cmd/now` 目录中。它输出当前 UTC 时间（RFC3339 格式），通常在构建脚本或 Makefile 中用于获取构建时间戳。

## 函数说明

### `main`

```go
func main()
```

输出 `time.Now().UTC().Format(time.RFC3339)` 到标准输出，无换行。

**输出示例**：`2026-04-30T08:15:30Z`

## 与其他文件的关系

| 依赖方向 | 文件/包 | 关系说明 |
|---------|--------|---------|
| 被调用 | `Makefile` | 构建时通过 `go run cmd/now/main.go` 获取构建时间戳，注入到 `internal/version.BuildTime`。 |
