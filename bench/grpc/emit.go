package main

import (
	"context"
	"flag"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	pb "celestialtree/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/structpb"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:7778", "grpc address host:port")
	n := flag.Int("n", 10000, "total requests")
	c := flag.Int("c", 20, "concurrency")
	timeout := flag.Duration("timeout", 10*time.Second, "per-request timeout")
	flag.Parse()

	// 建议：一个 grpc 连接共享，多 goroutine 并发复用（gRPC client 是并发安全的）
	conn, err := grpc.NewClient(
		*addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(5*time.Second),
	)
	if err != nil {
		panic(fmt.Errorf("grpc dial failed: %w", err))
	}
	defer conn.Close()

	client := pb.NewCelestialTreeServiceClient(conn)

	// payload 32B：如果你 proto 用的是 google.protobuf.Struct，
	// 那这里模拟一个 object payload；如果你未来改成 Value/bytes，这里再改。
	payloadMap := map[string]any{
		"data": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx", // 32 chars
	}
	st, _ := structpb.NewStruct(payloadMap)

	req := &pb.EmitRequest{
		Type:    "bench",
		Message: "bench payload 32B",
		Parents: []uint64{},
		Payload: st,
	}

	var ok uint64
	var fail uint64

	lat := make([]time.Duration, 0, *n)
	var latMu sync.Mutex

	start := time.Now()

	jobs := make(chan struct{}, *n)
	for i := 0; i < *n; i++ {
		jobs <- struct{}{}
	}
	close(jobs)

	var wg sync.WaitGroup
	wg.Add(*c)

	for w := 0; w < *c; w++ {
		go func() {
			defer wg.Done()

			for range jobs {
				t0 := time.Now()

				ctx, cancel := context.WithTimeout(context.Background(), *timeout)
				_, err := client.Emit(ctx, req)
				cancel()

				dt := time.Since(t0)

				latMu.Lock()
				lat = append(lat, dt)
				latMu.Unlock()

				if err != nil {
					atomic.AddUint64(&fail, 1)
					continue
				}
				atomic.AddUint64(&ok, 1)
			}
		}()
	}

	wg.Wait()
	elapsed := time.Since(start)

	total := atomic.LoadUint64(&ok) + atomic.LoadUint64(&fail)
	rps := float64(total) / elapsed.Seconds()

	sort.Slice(lat, func(i, j int) bool { return lat[i] < lat[j] })

	p := func(x float64) time.Duration {
		if len(lat) == 0 {
			return 0
		}
		i := int(float64(len(lat)-1) * x)
		return lat[i]
	}

	fmt.Printf(
		"[go-bench-grpc] total=%d ok=%d fail=%d rps=%.1f "+
			"lat_ms(p50=%.2f p90=%.2f p99=%.2f max=%.2f)\n",
		total,
		ok,
		fail,
		rps,
		float64(p(0.50).Milliseconds()),
		float64(p(0.90).Milliseconds()),
		float64(p(0.99).Milliseconds()),
		float64(p(1.00).Milliseconds()),
	)
}
