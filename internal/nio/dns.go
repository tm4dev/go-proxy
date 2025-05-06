package nio

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/phuslu/lru"
	"github.com/vlourme/go-proxy/internal/config"
)

var resolver = &net.Resolver{
	PreferGo: true,
	Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
		d := net.Dialer{
			Timeout: 5 * time.Second,
		}
		return d.DialContext(ctx, "udp", "1.1.1.1:53")
	},
}

const CACHE_TTL = 3 * time.Minute

var dnsCache = lru.NewTTLCache[string, string](4096)

// ResolveHostname resolves the hostname to an IP address
// based on the network type.
func ResolveHostname(hostname string, networkType config.NetworkType) (string, error) {
	if isLocalhost(hostname) {
		if !config.Get().DebugMode {
			return "", fmt.Errorf("localhost is not allowed in non-debug mode")
		}
		return hostname, nil
	}

	ip, ok := dnsCache.Get(hostname)
	if ok {
		return ip, nil
	}

	addrs, err := resolver.LookupHost(context.Background(), hostname)
	if err != nil {
		return "", err
	}

addrLoop:
	for _, addr := range addrs {
		switch networkType {
		case config.NetworkTypeIPv4:
			if !strings.Contains(addr, ":") {
				ip = addr
				break addrLoop
			}
		case config.NetworkTypeIPv6:
			if strings.Contains(addr, ":") {
				ip = addr
				break addrLoop
			}
		}
	}

	if ip == "" {
		return "", fmt.Errorf("no %s address found", networkType)
	}

	if networkType == config.NetworkTypeIPv6 {
		ip = "[" + ip + "]"
	}

	dnsCache.Set(hostname, ip, CACHE_TTL)
	return ip, nil
}

func isLocalhost(hostname string) bool {
	return hostname == "localhost" || hostname == "127.0.0.1" || hostname == "::1" || hostname == "[::1]"
}
