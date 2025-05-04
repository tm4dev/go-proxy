# Go Proxy Server

A configurable proxy server in Go, supporting rotating IPv4/IPv6 addresses, session and external authentication.
## Features

- IPv4 or IPv6 back-connect
- Multiple IPv6-IPv4 prefixes supported
- Session and timeout support to re-use generated IP.
- HTTP/HTTPS tunneling
- Up to 16k requests per second.
- DNS resolution and caching.
- Automatic IPv6 routing and sysctl setup.
- One core = one listener via reuseport.

## Limitations

- No SOCKS5 support
- Redis authentication is not supported yet.
- After 1000 concurrent connections, RPS decreases, this is being worked on.

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
BenchmarkProxyServer/Concurrent=100-24         	1000000000	         0.007529 ns/op	         4.000 ms/avg	         2.000 ms/fast	         6.000 ms/slow	     13305 req/s
BenchmarkProxyServer/Concurrent=250-24         	1000000000	         0.01743 ns/op	        11.00 ms/avg	         3.000 ms/fast	        17.00 ms/slow	     14351 req/s
BenchmarkProxyServer/Concurrent=500-24         	1000000000	         0.03109 ns/op	        21.00 ms/avg	         2.000 ms/fast	        29.00 ms/slow	     16088 req/s
BenchmarkProxyServer/Concurrent=1000-24        	1000000000	         0.06422 ns/op	        43.00 ms/avg	         9.000 ms/fast	        62.00 ms/slow	     15575 req/s
BenchmarkProxyServer/Concurrent=2500-24        	   78427	     18964 ns/op	       307.0 ms/avg	         2.000 ms/fast	      1092 ms/slow	      1681 req/s
BenchmarkProxyServer/Concurrent=5000-24        	       1	9154174756 ns/op	         0 ms/avg	9223372036854 ms/fast	         0 ms/slow	       546.2 req/s
PASS
ok  	github.com/vlourme/go-proxy	13.621s
```