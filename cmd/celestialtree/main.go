package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"celestialtree/internal/httpapi"
	"celestialtree/internal/memory"
	"celestialtree/internal/tree"
	"celestialtree/internal/version"
)

func main() {
	host := flag.String("host", "0.0.0.0", "server listen host")
	port := flag.Int("port", 7777, "server listen port")
	flag.Parse()

	addr := fmt.Sprintf("%s:%d", *host, *port)

	store := memory.NewStore()

	// 创世事件（Genesis）
	_, err := store.Emit(tree.EmitRequest{
		Type:    "genesis",
		Parents: nil,
		Message: "CelestialTree begins.",
	})
	if err != nil {
		log.Fatalf("genesis failed: %v", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	httpapi.RegisterRoutes(mux, store)

	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 3 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf(
		"%s %s(%s) built at %s",
		version.Name,
		version.Version,
		version.GitCommit,
		version.BuildTime,
	)

	log.Printf("CelestialTree listening on http://%s", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
