package main

import (
	"log"
	"net/http"
	"time"

	"celestialtree/internal/tree"
)

func main() {
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
		Addr:              ":7777",
		Handler:           mux,
		ReadHeaderTimeout: 3 * time.Second,
	}

	log.Printf("CelestialTree listening on http://127.0.0.1:7777")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
