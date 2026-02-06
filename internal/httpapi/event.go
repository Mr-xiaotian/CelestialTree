package httpapi

import (
	"net/http"

	"github.com/Mr-xiaotian/CelestialTree/internal/memory"
	"github.com/Mr-xiaotian/CelestialTree/internal/tree"
)

func handleGetEvent(store *memory.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !requireMethod(w, r, http.MethodGet) {
			return
		}
		id, ok := parsePathUint64(w, r.URL.Path, "/event/")
		if !ok {
			return
		}

		ev, ok := store.Get(id)
		if !ok {
			writeJSON(w, 404, tree.ResponseError{Error: "not found"})
			return
		}
		writeJSON(w, 200, ev)
	}
}
