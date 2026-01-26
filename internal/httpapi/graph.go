package httpapi

import (
	"celestialtree/internal/tree"
	"net/http"
	"strconv"
	"strings"
)

func handleChildren(store *tree.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSON(w, 405, tree.ResponseError{Error: "method not allowed"})
			return
		}

		idStr := strings.TrimPrefix(r.URL.Path, "/children/")
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil || id == 0 {
			writeJSON(w, 400, tree.ResponseError{Error: "bad id"})
			return
		}

		children, ok := store.Children(id)
		if !ok {
			writeJSON(w, 404, tree.ResponseError{Error: "not found"})
			return
		}
		writeJSON(w, 200, children)
	}
}

func handleAncestors(store *tree.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSON(w, 405, tree.ResponseError{Error: "method not allowed"})
			return
		}

		idStr := strings.TrimPrefix(r.URL.Path, "/ancestors/")
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil || id == 0 {
			writeJSON(w, 400, tree.ResponseError{Error: "bad id"})
		}

		ancestors, ok := store.Ancestors(id)
		if !ok {
			writeJSON(w, 404, tree.ResponseError{Error: "not found"})
			return
		}

		writeJSON(w, 200, ancestors)
	}
}

func handleHeads(store *tree.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSON(w, 405, tree.ResponseError{Error: "method not allowed"})
			return
		}
		writeJSON(w, 200, store.Heads())
	}
}
