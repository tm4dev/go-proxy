package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type benchmarkResult struct {
	fastest    time.Duration
	slowest    time.Duration
	average    time.Duration
	total      time.Duration
	throughput float64
}

func runBenchmark(concurrency, totalRequests int, mode string, test_port int) benchmarkResult {
	var durations []time.Duration
	var durationsMu sync.Mutex
	var completed int64

	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        concurrency,
			MaxIdleConnsPerHost: concurrency,
			Proxy: func(req *http.Request) (*url.URL, error) {
				return url.Parse(mode + "://username:password@localhost:8080")
			},
		},
		Timeout: 10 * time.Second,
	}

	startTime := time.Now()
	var wg sync.WaitGroup

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				index := atomic.AddInt64(&completed, 1)
				if index > int64(totalRequests) {
					return
				}

				start := time.Now()
				req, err := http.NewRequest("GET", "http://[::1]:"+strconv.Itoa(test_port), nil)
				if err != nil {
					continue
				}
				resp, err := client.Do(req)
				if err == nil {
					resp.Body.Close()
				}
				duration := time.Since(start)

				durationsMu.Lock()
				durations = append(durations, duration)
				durationsMu.Unlock()
			}
		}()
	}

	wg.Wait()
	totalDuration := time.Since(startTime)

	var total time.Duration
	fastest := time.Duration(1<<63 - 1)
	slowest := time.Duration(0)

	for _, d := range durations {
		if d < fastest {
			fastest = d
		}
		if d > slowest {
			slowest = d
		}
		total += d
	}

	average := total / time.Duration(len(durations))
	throughput := float64(len(durations)) / totalDuration.Seconds()

	return benchmarkResult{
		fastest:    fastest,
		slowest:    slowest,
		average:    average,
		total:      totalDuration,
		throughput: throughput,
	}
}

func main() {
	mode := flag.String("mode", "http", "http, socks5, socks5-connect")
	test_port := flag.Int("test_port", 8081, "test port")
	flag.Parse()

	concurrencyLevels := [][]int{
		{100, 600},
		{250, 1500},
		{500, 4000},
		{1000, 6000},
		{2500, 10000},
		{5000, 20000},
		{10000, 40000},
	}

	for _, c := range concurrencyLevels {
		fmt.Printf("Running benchmark: %d concurrency, %d total requests\n", c[0], c[1])
		result := runBenchmark(c[0], c[1], *mode, *test_port)
		fmt.Printf("Fastest:  %v\n", result.fastest)
		fmt.Printf("Slowest:  %v\n", result.slowest)
		fmt.Printf("Average:  %v\n", result.average)
		fmt.Printf("Total:    %v\n", result.total)
		fmt.Printf("Throughput: %.2f req/s\n\n", result.throughput)
	}
}
