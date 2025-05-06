# Go Proxy Server

A configurable proxy server in Go, supporting rotating IPv4/IPv6 addresses, session and external authentication.
## Features

- IPv4 or IPv6 back-connect
- Multiple IPv6-IPv4 prefixes supported, with per-country prefixes
- Session and timeout support to re-use generated IP.
- Up to 16,000 requests per second.
- DNS resolution and caching.
- Automatic IPv6 routing and sysctl setup.
- One core = one listener via reuseport.

## Limitations

- No SOCKS5 support
- Redis authentication is not supported yet.

## Setup

### Prerequisites

- [Go](https://golang.org/dl/)
- Server with IPv6 support and large-enough subnet ([Hetzner](https://hetzner.cloud/?ref=BV2rohR8OBWQ) offers /64 subnets, *sponsored link*).

### Configuration

Copy the `config.example.yaml` file to `config.yaml` and modify the settings as needed.

Example:
```yaml
listen_address: "::"
listen_port: 8080
debug_mode: false # Enable pretty-print logs, don't enable in production
test_port: -1 # Enable a test server on port 8081 for benchmarking
network_type: "tcp6" # tcp = dual-stack, tcp6 = IPv6 only, tcp4 = IPv4 only
max_timeout: 30 # Maximum timeout for a request in seconds
http_close: true # Force "Connection: close" header in HTTP-only requests (faster)
auth:
  type: "credentials" # none, credentials, redis
  credentials:
    username: "username"
    password: "password"
  redis:
    dsn: "redis://localhost:6379" # Will compare the username as key to the password
bind_prefixes:
  - "2a14:dead:beef::1/48" # List of prefixes to bind to
  - "2a14:dead:feed::1/48"
located_prefixes:
  ch:
    - "2a14:dead:beef::/48"
  uk:
    - "2a14:dead:feed::/48"
deleted_headers: # List of headers to delete (HTTP only), this make proxy anonymous
  - "Proxy-Authorization"
  - "Proxy-Connection"
```

### Build

```bash
go build .
./go-proxy
```

### Usage example

```bash
# Without credentials, if enabled
curl -x http://localhost:8080 http://api.ipquery.io

# Session
curl -x http://john-session-abcdef1234:doe@localhost:8080 http://api.ipquery.io

# Session with timeout
curl -x http://john-session-abcdef1234-timeout-10:doe@localhost:8080 http://api.ipquery.io
```

> Session ID must be alphanumeric, between 6 and 24 characters.
> The timeout is in minutes, between 1 and 30 minutes.

## Benchmark

Benchmark can be run with `go test -bench=.`. Configuration should have `test_port` set to 8081 and credentials
set to `username:password`.

**Benchmark results on 24-cores dedicated server:**
```sh
goos: linux
goarch: amd64
pkg: github.com/vlourme/go-proxy
cpu: Intel(R) Xeon(R) CPU E5-2630L v2 @ 2.40GHz
BenchmarkProxyServer
BenchmarkProxyServer/Concurrent=100
BenchmarkProxyServer/Concurrent=100-24         	1000000000	         0.006638 ns/op	         3.000 ms/avg	         1.000 ms/fast	         6.000 ms/slow	     15103 req/s
BenchmarkProxyServer/Concurrent=250
BenchmarkProxyServer/Concurrent=250-24         	1000000000	         0.01649 ns/op	         9.000 ms/avg	         2.000 ms/fast	        15.00 ms/slow	     15173 req/s
BenchmarkProxyServer/Concurrent=500
BenchmarkProxyServer/Concurrent=500-24         	1000000000	         0.03191 ns/op	        14.00 ms/avg	         1.000 ms/fast	        30.00 ms/slow	     15676 req/s
BenchmarkProxyServer/Concurrent=1000
BenchmarkProxyServer/Concurrent=1000-24        	1000000000	         0.06184 ns/op	        35.00 ms/avg	         6.000 ms/fast	        60.00 ms/slow	     16176 req/s
BenchmarkProxyServer/Concurrent=2500
BenchmarkProxyServer/Concurrent=2500-24        	1000000000	         0.1845 ns/op	       117.0 ms/avg	        23.00 ms/fast	       180.0 ms/slow	     13551 req/s
BenchmarkProxyServer/Concurrent=5000
BenchmarkProxyServer/Concurrent=5000-24        	1000000000	         0.3751 ns/op	       247.0 ms/avg	        25.00 ms/fast	       365.0 ms/slow	     13332 req/s
PASS
ok  	github.com/vlourme/go-proxy	10.439s
```