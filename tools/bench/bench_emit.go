package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type EmitReq struct {
	Type    string   `json:"type"`
	Parents []uint64 `json:"parents"`
	Message string   `json:"message"`
	Payload []byte   `json:"payload"`
}

func main() {
	base := flag.String("base", "http://127.0.0.1:7777", "base url")
	n := flag.Int("n", 10000, "total requests")
	c := flag.Int("c", 20, "concurrency")
	flag.Parse()

	tr := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		MaxConnsPerHost:     100,
		IdleConnTimeout:     60 * time.Second,
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   10 * time.Second,
	}

	reqBody := EmitReq{
		Type:    "bench",
		Parents: []uint64{}, // chain parent 可自行改
		Message: "bench payload 32B",
		Payload: []byte("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"), // 32 bytes
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
				buf, _ := json.Marshal(reqBody)

				t0 := time.Now()
				resp, err := client.Post(
					*base+"/emit",
					"application/json",
					bytes.NewReader(buf),
				)
				dt := time.Since(t0)

				latMu.Lock()
				lat = append(lat, dt)
				latMu.Unlock()

				if err != nil {
					atomic.AddUint64(&fail, 1)
					continue
				}
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()

				if resp.StatusCode == 200 {
					atomic.AddUint64(&ok, 1)
				} else {
					atomic.AddUint64(&fail, 1)
				}
			}
		}()
	}

	wg.Wait()
	elapsed := time.Since(start)

	total := atomic.LoadUint64(&ok) + atomic.LoadUint64(&fail)
	rps := float64(total) / elapsed.Seconds()

	// 简单算 p50 / p90 / p99
	p := func(x float64) time.Duration {
		if len(lat) == 0 {
			return 0
		}
		i := int(float64(len(lat)-1) * x)
		return lat[i]
	}

	sort.Slice(lat, func(i, j int) bool {
		return lat[i] < lat[j]
	})

	fmt.Printf(
		"[go-bench] total=%d ok=%d fail=%d rps=%.1f "+
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
