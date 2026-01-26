package httpapi

import (
	"celestialtree/internal/tree"
	"net/http"
)

func handleChildren(store *tree.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !requireMethod(w, r, http.MethodGet) {
			return
		}
		id, ok := parsePathUint64(w, r.URL.Path, "/children/")
		if !ok {
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
		if !requireMethod(w, r, http.MethodGet) {
			return
		}
		id, ok := parsePathUint64(w, r.URL.Path, "/ancestors/")
		if !ok {
			return
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
		if !requireMethod(w, r, http.MethodGet) {
			return
		}
		writeJSON(w, 200, store.Heads())
	}
}
