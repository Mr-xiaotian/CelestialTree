package httpapi

import (
	"net/http"
	"runtime"
	"time"

	"github.com/Mr-xiaotian/CelestialTree/internal/memory"
)

func handleSnapshot(store *memory.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !requireMethod(w, r, http.MethodGet) {
			return
		}
		load := store.Snapshot()
		writeJSON(w, 200, map[string]any{
			"ts":            time.Now().Unix(),
			"goroutines":    runtime.NumGoroutine(),
			"events":        load.Events,
			"edges":         load.Edges,
			"heads":         load.Heads,
			"subscribers":   load.Subscribers,
			"next_event_id": load.NextEventID,
		})
	}
}
