package tree

import (
	"celestialtree/internal/version"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func RegisterRoutes(mux *http.ServeMux, store *Store) {
	mux.HandleFunc("/emit", handleEmit(store))
	mux.HandleFunc("/event/", handleGetEvent(store))
	mux.HandleFunc("/children/", handleChildren(store))
	mux.HandleFunc("/descendants/", handleDescendants(store))
	mux.HandleFunc("/ancestors/", handleAncestors(store))
	mux.HandleFunc("/heads", handleHeads(store))
	mux.HandleFunc("/subscribe", handleSubscribe(store))
	mux.HandleFunc("/healthz", handleHealthz())
	mux.HandleFunc("/version", handleVersion())
}

func handleEmit(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, 405, ResponseError{Error: "method not allowed"})
			return
		}

		var req EmitRequest
		if err := readJSON(r, &req); err != nil {
			writeJSON(w, 400, ResponseError{Error: "invalid json", Detail: err.Error()})
			return
		}

		ev, err := store.Emit(req)
		if err != nil {
			writeJSON(w, 400, ResponseError{Error: "emit failed", Detail: err.Error()})
			return
		}

		writeJSON(w, 200, EmitResponse{ID: ev.ID})
	}
}

func handleGetEvent(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSON(w, 405, ResponseError{Error: "method not allowed"})
			return
		}

		idStr := strings.TrimPrefix(r.URL.Path, "/event/")
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil || id == 0 {
			writeJSON(w, 400, ResponseError{Error: "bad id"})
			return
		}

		ev, ok := store.Get(id)
		if !ok {
			writeJSON(w, 404, ResponseError{Error: "not found"})
			return
		}
		writeJSON(w, 200, ev)
	}
}

func handleChildren(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSON(w, 405, ResponseError{Error: "method not allowed"})
			return
		}

		idStr := strings.TrimPrefix(r.URL.Path, "/children/")
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil || id == 0 {
			writeJSON(w, 400, ResponseError{Error: "bad id"})
			return
		}

		children, ok := store.Children(id)
		if !ok {
			writeJSON(w, 404, ResponseError{Error: "not found"})
			return
		}
		writeJSON(w, 200, children)
	}
}

func handleDescendants(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSON(w, 405, ResponseError{Error: "method not allowed"})
			return
		}

		idStr := strings.TrimPrefix(r.URL.Path, "/descendants/")
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil || id == 0 {
			writeJSON(w, 400, ResponseError{Error: "bad id"})
			return
		}

		view := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("view")))
		switch view {
		case "", "struct":
			tree, ok := store.DescendantsTree(id)
			if !ok {
				writeJSON(w, 404, ResponseError{Error: "not found"})
				return
			}
			writeJSON(w, 200, tree)
			return

		case "meta":
			tree, ok := store.DescendantsTreeView(id)
			if !ok {
				writeJSON(w, 404, ResponseError{Error: "not found"})
				return
			}
			writeJSON(w, 200, tree)
			return

		default:
			writeJSON(w, 400, ResponseError{Error: "bad view", Detail: fmt.Sprintf("unknown view: %s", view)})
			return
		}
	}
}

func handleAncestors(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSON(w, 405, ResponseError{Error: "method not allowed"})
			return
		}

		idStr := strings.TrimPrefix(r.URL.Path, "/ancestors/")
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil || id == 0 {
			writeJSON(w, 400, ResponseError{Error: "bad id"})
		}

		ancestors, ok := store.Ancestors(id)
		if !ok {
			writeJSON(w, 404, ResponseError{Error: "not found"})
			return
		}

		writeJSON(w, 200, ancestors)
	}
}

func handleHeads(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSON(w, 405, ResponseError{Error: "method not allowed"})
			return
		}
		writeJSON(w, 200, store.Heads())
	}
}

func handleHealthz() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]any{"ok": true, "ts": time.Now().Unix()})
	}
}

// 为了避免 handlers.go 过长，SSE 的 handler 放到 sse_handler.go 也行。
// 这里先留在一个文件里。
func handleSubscribe(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSON(w, 405, ResponseError{Error: "method not allowed"})
			return
		}

		flusher, ok := w.(http.Flusher)
		if !ok {
			writeJSON(w, 500, ResponseError{Error: "streaming not supported"})
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		_, ch, cancel := store.Subscribe()
		defer cancel()

		fmt.Fprintf(w, "event: hello\ndata: %s\n\n", `{"message":"subscribed"}`)
		flusher.Flush()

		ctx := r.Context()
		for {
			select {
			case <-ctx.Done():
				return
			case ev, ok := <-ch:
				if !ok {
					return
				}
				// 这里我们直接复用 writeJSON 是不行的，因为 SSE 不是 JSON 响应而是流。
				b, _ := jsonMarshal(ev)
				fmt.Fprintf(w, "event: emit\ndata: %s\n\n", string(b))
				flusher.Flush()
			}
		}
	}
}

func handleVersion() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]any{
			"name":    version.Name,
			"version": version.Version,
			"commit":  version.GitCommit,
			"build":   version.BuildTime,
		})
	}
}
