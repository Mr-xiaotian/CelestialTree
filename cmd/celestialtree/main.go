package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"celestialtree/internal/grpcapi"
	"celestialtree/internal/httpapi"
	"celestialtree/internal/memory"
	"celestialtree/internal/tree"
	"celestialtree/internal/version"
	pb "celestialtree/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	host := flag.String("host", "0.0.0.0", "server listen host (http/grpc)")
	port := flag.Int("port", 7777, "http listen port")

	grpcHost := flag.String("grpc_host", "", "grpc listen host (default: same as -host)")
	grpcPort := flag.Int("grpc_port", 7778, "grpc listen port")

	flag.Parse()

	httpAddr := fmt.Sprintf("%s:%d", *host, *port)

	gh := *grpcHost
	if gh == "" {
		gh = *host
	}
	grpcAddr := fmt.Sprintf("%s:%d", gh, *grpcPort)

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

	// ---------------- HTTP ----------------
	mux := http.NewServeMux()
	httpapi.RegisterRoutes(mux, store)

	httpSrv := &http.Server{
		Addr:              httpAddr,
		Handler:           mux,
		ReadHeaderTimeout: 3 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// ---------------- gRPC ----------------
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("grpc listen failed on %s: %v", grpcAddr, err)
	}

	grpcSrv := grpc.NewServer(
	// 以后可以在这里加拦截器：日志、鉴权、trace、限流...
	)

	pb.RegisterCelestialTreeServiceServer(grpcSrv, grpcapi.New(store))
	reflection.Register(grpcSrv) // 方便 grpcurl 调试（生产环境可按需关）

	// ---------------- logs ----------------
	log.Printf(
		"%s %s(%s) built at %s",
		version.Name,
		version.Version,
		version.GitCommit,
		version.BuildTime,
	)

	// ---------------- run ----------------
	errCh := make(chan error, 2)

	go func() {
		log.Printf("CelestialTree listening on http://%s", httpAddr)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("http server error: %w", err)
		}
	}()

	go func() {
		log.Printf("CelestialTree listening on grpc://%s", grpcAddr)
		if err := grpcSrv.Serve(lis); err != nil {
			errCh <- fmt.Errorf("grpc server error: %w", err)
		}
	}()

	// ---------------- graceful shutdown ----------------
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		log.Printf("shutdown signal received: %v", sig)
	case err := <-errCh:
		log.Printf("server error received: %v", err)
	}

	// 先停 gRPC（它没有 context 超时参数，用 GracefulStop）
	go func() {
		grpcSrv.GracefulStop()
	}()

	// 再停 HTTP（带超时）
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpSrv.Shutdown(ctx); err != nil {
		log.Printf("http shutdown error: %v", err)
	}

	log.Printf("bye.")
}
