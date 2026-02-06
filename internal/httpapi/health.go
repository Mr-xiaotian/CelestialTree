package httpapi

import (
	"net/http"
	"time"

	"github.com/Mr-xiaotian/CelestialTree/internal/version"
)

func handleHealthz() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]any{"ok": true, "ts": time.Now().Unix()})
	}
}

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
