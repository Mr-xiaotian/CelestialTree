package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"celestialtree/internal/tree"
)

func main() {
	host := flag.String("host", "127.0.0.1", "server listen host")
	port := flag.Int("port", 7777, "server listen port")
	flag.Parse()

	addr := fmt.Sprintf("%s:%d", *host, *port)

	store := tree.NewStore()

	// 可选：创世事件（Genesis）
	gen, err := store.Emit(tree.EmitRequest{
		Type:    "genesis",
		Parents: nil,
		Payload: map[string]any{"msg": "CelestialTree begins."},
		Meta:    map[string]any{"version": "v0"},
	})
	if err != nil {
		log.Fatalf("genesis failed: %v", err)
	}
	log.Printf("genesis id=%d hash=%s", gen.ID, gen.Hash)

	mux := http.NewServeMux()
	tree.RegisterRoutes(mux, store)

	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 3 * time.Second,
	}

	log.Printf("CelestialTree listening on http://%s", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
