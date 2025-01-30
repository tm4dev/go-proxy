# IPv6 Proxy Server

A simple proxy server implementation in Go, featuring IPv4 to IPv6 tunneling with rotating IP and session support.

## Features

- IPv4 backconnect
- IPv6 rotation via CIDR
- Session support
- HTTP/HTTPS tunneling
- Serving over 1200 requests per second (maximum tested)

## Setup

### Prerequisites

- [Go](https://golang.org/dl/)
- Server with IPv6 support and large-enough subnet ([Hetzner](https://hetzner.cloud/?ref=BV2rohR8OBWQ) offers /64 subnets, *sponsored link*).
- IPv6 forwarding enabled on the server.

### IPv6 forwarding

The configuration might change depending on the network interface and OS.
The following configuration is for Ubuntu 24.04.

1. Add the following to `/etc/sysctl.conf`:
```conf
net.ipv6.conf.eth0.proxy_ndp=1
net.ipv6.conf.all.proxy_ndp=1
net.ipv6.conf.default.forwarding=1
net.ipv6.conf.all.forwarding=1
net.ipv6.ip_nonlocal_bind=1
```

2. Apply the changes:
```sh
sudo sysctl -p
```

3. Link the IPv6 subnet to the server:
```sh
ip -6 addr add local <IPv6_SUBNET>/64 dev eth0
```

### Build

```bash
go run . \
    [-username USERNAME] \
    [-password PASSWORD] \
    [-port PORT] \
    CIDR
```

### Example

```bash
go run . -username john -password doe 2001:0db8:85a3:0001::/64
```

### Usage example

```bash
# Rotating IPv6 address
curl -x http://john:doe@<HOST>:8080 <URL>

# Session
curl -x http://john-session-abcdef1234:doe@<HOST>:8080 <URL>

# Session with lifetime
curl -x http://john-session-abcdef1234-lifetime-10:doe@<HOST>:8080 <URL>
```

> Session ID must be alphanumeric, between 6 and 10 characters.
> The lifetime is in minutes, between 1 and 60 minutes.

## Benchmark

Benchmarks are measured with [hey](https://github.com/rakyll/hey).

**10000 requests, 1000 concurrent connections, rotating IPv6 address:**
```sh
hey -n 10000 -c 1000 -x "http://test:test@<HOST>:8080" "https://api.ipquery.io/?format=json"

Summary:
  Total:        8.3602 secs
  Slowest:      4.3739 secs
  Fastest:      0.2211 secs
  Average:      0.6059 secs
  Requests/sec: 1196.1435
```

**10000 requests, 1000 concurrent connections, static session:**

```sh
hey -n 10000 -c 1000 -x "http://username-session-static:password@<HOST>:8080" "https://api.ipquery.io/?format=json"

Summary:
  Total:        8.0344 secs
  Slowest:      2.9934 secs
  Fastest:      0.2267 secs
  Average:      0.5678 secs
  Requests/sec: 1244.643
```

## Sponsors
<a href="https://sneakersapi.dev"><img src="https://i.ibb.co/ksdfJqh6/A6-1-1.png" alt="Sponsored by SneakersAPI" width="250"></a>
