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
	"strconv"
	"syscall"
	"time"

	"github.com/Mr-xiaotian/CelestialTree/internal/grpcapi"
	"github.com/Mr-xiaotian/CelestialTree/internal/httpapi"
	"github.com/Mr-xiaotian/CelestialTree/internal/memory"
	"github.com/Mr-xiaotian/CelestialTree/internal/tree"
	"github.com/Mr-xiaotian/CelestialTree/internal/version"
	pb "github.com/Mr-xiaotian/CelestialTree/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	httpAddrFlag := flag.String("http_addr", "", "http listen addr host:port (preferred)")
	grpcAddrFlag := flag.String("grpc_addr", "", "grpc listen addr host:port (preferred)")

	host := flag.String("host", "0.0.0.0", "server listen host (http/grpc)")
	httpPort := flag.Int("http_port", 7777, "http listen port")
	grpcPort := flag.Int("grpc_port", 7778, "grpc listen port")

	flag.Parse()

	httpAddr := *httpAddrFlag
	if httpAddr == "" {
		httpAddr = net.JoinHostPort(*host, strconv.Itoa(*httpPort))
	}

	grpcAddr := *grpcAddrFlag
	if grpcAddr == "" {
		grpcAddr = net.JoinHostPort(*host, strconv.Itoa(*grpcPort))
	}

	store := memory.NewStore()

	// 创世事件（Genesis）
	_, err := store.Emit(tree.EmitRequest{
		Type:    "genesis",
		Parents: nil,
		Message: "CelestialTree begins.",
	})
	if err != nil {
		log.Fatalf("genesis failed: %v", err)
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
