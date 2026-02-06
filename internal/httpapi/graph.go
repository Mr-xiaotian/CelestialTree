package httpapi

import (
	"net/http"

	"github.com/Mr-xiaotian/CelestialTree/internal/memory"
	"github.com/Mr-xiaotian/CelestialTree/internal/tree"
)

func handleChildren(store *memory.Store) http.HandlerFunc {
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

func handleAncestors(store *memory.Store) http.HandlerFunc {
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

func handleHeads(store *memory.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !requireMethod(w, r, http.MethodGet) {
			return
		}
		writeJSON(w, 200, store.Heads())
	}
}
