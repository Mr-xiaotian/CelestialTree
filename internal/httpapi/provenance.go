package httpapi

import (
	"celestialtree/internal/tree"
	"fmt"
	"net/http"
)

func handleProvenance(store *tree.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !requireMethod(w, r, http.MethodGet) {
			return
		}
		id, ok := parsePathUint64(w, r.URL.Path, "/provenance/")
		if !ok {
			return
		}

		view := normalizeView(r.URL.Query().Get("view"))
		switch view {
		case "", "struct":
			pt, ok := store.ProvenanceTree(id)
			if !ok {
				writeJSON(w, 404, tree.ResponseError{Error: "not found"})
				return
			}
			writeJSON(w, 200, pt)
			return

		case "meta":
			pt, ok := store.ProvenanceTreeMeta(id)
			if !ok {
				writeJSON(w, 404, tree.ResponseError{Error: "not found"})
				return
			}
			writeJSON(w, 200, pt)
			return

		default:
			writeJSON(w, 400, tree.ResponseError{Error: "bad view", Detail: fmt.Sprintf("unknown view: %s", view)})
			return
		}
	}
}

func handleProvenanceBatch(store *tree.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !requireMethod(w, r, http.MethodPost) {
			return
		}

		var req tree.TreeBatchRequest
		if err := readJSON(r, &req); err != nil {
			writeJSON(w, 400, tree.ResponseError{Error: "invalid json", Detail: err.Error()})
			return
		}
		if len(req.IDs) == 0 {
			writeJSON(w, 400, tree.ResponseError{Error: "ids is required"})
			return
		}

		view := normalizeView(req.View)
		switch view {
		case "", "struct":
			forest, ok := store.ProvenanceForest(req.IDs)
			if !ok {
				writeJSON(w, 404, tree.ResponseError{Error: "not found"})
				return
			}
			writeJSON(w, 200, forest)
			return

		case "meta":
			forest, ok := store.ProvenanceForestMeta(req.IDs)
			if !ok {
				writeJSON(w, 404, tree.ResponseError{Error: "not found"})
				return
			}
			writeJSON(w, 200, forest)
			return

		default:
			writeJSON(w, 400, tree.ResponseError{Error: "bad view", Detail: fmt.Sprintf("unknown view: %s", view)})
			return
		}
	}
}
