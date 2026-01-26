package httpapi

import (
	"celestialtree/internal/tree"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func handleProvenance(store *tree.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSON(w, 405, tree.ResponseError{Error: "method not allowed"})
			return
		}

		idStr := strings.TrimPrefix(r.URL.Path, "/provenance/")
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil || id == 0 {
			writeJSON(w, 400, tree.ResponseError{Error: "bad id"})
			return
		}

		view := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("view")))
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
		if r.Method != http.MethodPost {
			writeJSON(w, 405, tree.ResponseError{Error: "method not allowed"})
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

		view := strings.ToLower(strings.TrimSpace(req.View))
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
