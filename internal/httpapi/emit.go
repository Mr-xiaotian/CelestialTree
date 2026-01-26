package httpapi

import (
	"celestialtree/internal/tree"
	"net/http"
)

func handleEmit(store *tree.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, 405, tree.ResponseError{Error: "method not allowed"})
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
