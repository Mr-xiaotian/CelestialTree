package httpapi

import (
	"net/http"

	"github.com/Mr-xiaotian/CelestialTree/internal/memory"
	"github.com/Mr-xiaotian/CelestialTree/internal/tree"
)

// handleChildren 处理 GET /children/{id}，返回指定事件的直接子事件 ID 列表。
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

// handleAncestors 处理 GET /ancestors/{id}，返回指定事件可达的所有根节点 ID。
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

// handleHeads 处理 GET /heads，返回 DAG 中所有叶子节点的 ID 列表。
func handleHeads(store *memory.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !requireMethod(w, r, http.MethodGet) {
			return
		}
		writeJSON(w, 200, store.Heads())
	}
}

// handleRoots 处理 GET /roots，返回 DAG 中所有根节点的 ID 列表。
func handleRoots(store *memory.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !requireMethod(w, r, http.MethodGet) {
			return
		}
		writeJSON(w, 200, store.Roots())
	}
}
