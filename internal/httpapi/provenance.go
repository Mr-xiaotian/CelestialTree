package httpapi

import (
	"celestialtree/internal/memory"
	"celestialtree/internal/tree"
	"fmt"
	"net/http"
)

func handleProvenance(store *memory.Store) http.HandlerFunc {
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
			pt, err := store.ProvenanceTree(id)
			if err != nil {
				writeJSON(w, 404, tree.ResponseError{Error: "provenance process failed", Detail: err.Error()})
				return
			}
			writeJSON(w, 200, pt)
			return

		case "meta":
			pt, err := store.ProvenanceTreeMeta(id)
			if err != nil {
				writeJSON(w, 404, tree.ResponseError{Error: "provenance process failed", Detail: err.Error()})
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

func handleProvenanceBatch(store *memory.Store) http.HandlerFunc {
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
			forest, err := store.ProvenanceForest(req.IDs)
			if err != nil {
				writeJSON(w, 404, tree.ResponseError{Error: "provenance process failed", Detail: err.Error()})
				return
			}
			writeJSON(w, 200, forest)
			return

		case "meta":
			forest, err := store.ProvenanceForestMeta(req.IDs)
			if err != nil {
				writeJSON(w, 404, tree.ResponseError{Error: "provenance process failed", Detail: err.Error()})
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
