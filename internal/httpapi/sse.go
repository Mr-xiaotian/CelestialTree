package httpapi

import (
	"celestialtree/internal/memory"
	"celestialtree/internal/tree"
	"encoding/json"
	"fmt"
	"net/http"
)

// 为了避免 handlers.go 过长，SSE 的 handler 放到 sse_handler.go 也行。
// 这里先留在一个文件里。
func handleSubscribe(store *memory.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSON(w, 405, tree.ResponseError{Error: "method not allowed"})
			return
		}

		flusher, ok := w.(http.Flusher)
		if !ok {
			writeJSON(w, 500, tree.ResponseError{Error: "streaming not supported"})
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
				b, _ := json.Marshal(ev)
				fmt.Fprintf(w, "event: emit\ndata: %s\n\n", string(b))
				flusher.Flush()
			}
		}
	}
}
