package httpapi

import (
	"net/http"

	"github.com/Mr-xiaotian/CelestialTree/internal/memory"
)

// handleSnapshot 处理 GET /snapshot，返回系统状态快照。
func handleSnapshot(store *memory.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !requireMethod(w, r, http.MethodGet) {
			return
		}
		load := store.Snapshot()
		writeJSON(w, 200, load)
	}
}
