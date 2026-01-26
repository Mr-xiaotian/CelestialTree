package httpapi

import (
	"celestialtree/internal/tree"
	"net/http"
	"strconv"
	"strings"
)

func handleGetEvent(store *tree.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSON(w, 405, tree.ResponseError{Error: "method not allowed"})
			return
		}

		idStr := strings.TrimPrefix(r.URL.Path, "/event/")
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil || id == 0 {
			writeJSON(w, 400, tree.ResponseError{Error: "bad id"})
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
