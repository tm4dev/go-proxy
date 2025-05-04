package main

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"testing"
	"time"
)

type benchmarkResult struct {
	fastest    time.Duration
	slowest    time.Duration
	average    time.Duration
	throughput float64
}

func runBenchmark(concurrent int) benchmarkResult {
	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        concurrent,
			MaxIdleConnsPerHost: concurrent,
			Proxy: func(req *http.Request) (*url.URL, error) {
				return url.Parse("http://username:password@localhost:8080")
			},
		},
	}

	startTime := time.Now()
	var wg sync.WaitGroup
	durations := make([]time.Duration, concurrent)

	for i := 0; i < concurrent; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			reqStart := time.Now()

			req, err := http.NewRequest("GET", "http://[::1]:8081", nil)
			if err != nil {
				return
			}

			resp, err := client.Do(req)
			if err != nil {
				return
			}
			resp.Body.Close()

			durations[index] = time.Since(reqStart)
		}(i)
	}

	wg.Wait()
	totalDuration := time.Since(startTime)

	// Calculate statistics from the array
	var totalTime time.Duration
	fastest := time.Duration(1<<63 - 1) // Initialize with max duration
	slowest := time.Duration(0)

	for i := 0; i < concurrent; i++ {
		duration := durations[i]
		if duration == 0 { // Skip failed requests
			continue
		}
		if duration < fastest {
			fastest = duration
		}
		if duration > slowest {
			slowest = duration
		}
		totalTime += duration
	}

	return benchmarkResult{
		fastest:    fastest,
		slowest:    slowest,
		average:    totalTime / time.Duration(concurrent),
		throughput: float64(concurrent) / totalDuration.Seconds(),
	}
}

func BenchmarkProxyServer(b *testing.B) {
	concurrentLevels := []int{100, 250, 500, 1000, 2500, 5000}

	for _, concurrent := range concurrentLevels {
		b.Run(fmt.Sprintf("Concurrent=%d", concurrent), func(b *testing.B) {
			result := runBenchmark(concurrent)
			b.ReportMetric(result.throughput, "req/s")
			b.ReportMetric(float64(result.fastest.Milliseconds()), "ms/fast")
			b.ReportMetric(float64(result.slowest.Milliseconds()), "ms/slow")
			b.ReportMetric(float64(result.average.Milliseconds()), "ms/avg")
		})
	}
}
