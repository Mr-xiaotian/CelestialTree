package httpapi

import (
	"net/http"
	"time"

	"github.com/Mr-xiaotian/CelestialTree/internal/version"
)

// handleHealthz 处理 GET /healthz，返回健康检查状态。
func handleHealthz() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]any{"ok": true, "ts": time.Now().Unix()})
	}
}

// handleVersion 处理 GET /version，返回版本、commit 和构建时间。
func handleVersion() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]any{
			"name":    version.Name,
			"version": version.Version,
			"commit":  version.GitCommit,
			"build":   version.BuildTime,
		})
	}
}
