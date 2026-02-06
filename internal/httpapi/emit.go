package httpapi

import (
	"net/http"

	"github.com/Mr-xiaotian/CelestialTree/internal/memory"
	"github.com/Mr-xiaotian/CelestialTree/internal/tree"
)

func handleEmit(store *memory.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !requireMethod(w, r, http.MethodPost) {
			return
		}

		var req tree.EmitRequest
		if err := readJSON(r, &req); err != nil {
			writeJSON(w, 400, tree.ResponseError{Error: "invalid json", Detail: err.Error()})
			return
		}

		ev, err := store.Emit(req)
		if err != nil {
			writeJSON(w, 400, tree.ResponseError{Error: "emit failed", Detail: err.Error()})
			return
		}

		writeJSON(w, 200, tree.EmitResponse{ID: ev.ID})
	}
}
