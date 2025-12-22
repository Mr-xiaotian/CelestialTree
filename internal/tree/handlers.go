package tree

import (
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
	mux.HandleFunc("/heads", handleHeads(store))
	mux.HandleFunc("/subscribe", handleSubscribe(store))
	mux.HandleFunc("/healthz", handleHealthz())
}

func handleEmit(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, 405, map[string]any{"error": "method not allowed"})
			return
		}
		var req EmitRequest
		if err := readJSON(r, &req); err != nil {
			writeJSON(w, 400, map[string]any{"error": "invalid json", "detail": err.Error()})
			return
		}
		ev, err := store.Emit(req)
		if err != nil {
			writeJSON(w, 400, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, 200, EmitResponse{Event: ev})
	}
}

func handleGetEvent(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSON(w, 405, map[string]any{"error": "method not allowed"})
			return
		}

		idStr := strings.TrimPrefix(r.URL.Path, "/event/")
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil || id == 0 {
			writeJSON(w, 400, map[string]any{"error": "bad id"})
			return
		}

		ev, ok := store.Get(id)
		if !ok {
			writeJSON(w, 404, map[string]any{"error": "not found"})
			return
		}
		writeJSON(w, 200, ev)
	}
}

func handleChildren(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSON(w, 405, map[string]any{"error": "method not allowed"})
			return
		}

		idStr := strings.TrimPrefix(r.URL.Path, "/children/")
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil || id == 0 {
			writeJSON(w, 400, map[string]any{"error": "bad id"})
			return
		}

		children, ok := store.Children(id)
		if !ok {
			writeJSON(w, 404, map[string]any{"error": "not found"})
			return
		}
		writeJSON(w, 200, map[string]any{"id": id, "children": children})
	}
}

func handleDescendants(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSON(w, 405, map[string]any{"error": "method not allowed"})
			return
		}

		idStr := strings.TrimPrefix(r.URL.Path, "/descendants/")
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil || id == 0 {
			writeJSON(w, 400, map[string]any{"error": "bad id"})
			return
		}

		tree, ok := store.DescendantsTree(id)
		if !ok {
			writeJSON(w, 404, map[string]any{"error": "not found"})
			return
		}

		writeJSON(w, 200, tree)
	}
}

func handleHeads(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSON(w, 405, map[string]any{"error": "method not allowed"})
			return
		}
		writeJSON(w, 200, map[string]any{"heads": store.Heads()})
	}
}

func handleHealthz() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]any{"ok": true, "ts": time.Now().Unix()})
	}
}

// 为了避免 handlers.go 过长，SSE 的 handler 放到 sse_handler.go 也行。
// 这里先留在一个文件里：你要更细拆我也可以继续拆。
func handleSubscribe(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSON(w, 405, map[string]any{"error": "method not allowed"})
			return
		}

		flusher, ok := w.(http.Flusher)
		if !ok {
			writeJSON(w, 500, map[string]any{"error": "streaming not supported"})
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		_, ch, cancel := store.Subscribe()
		defer cancel()

		fmt.Fprintf(w, "event: hello\ndata: %s\n\n", `{"msg":"subscribed"}`)
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
