package nio

import (
	"context"
	"errors"
	"net"
	"time"

	"github.com/phuslu/lru"
	"github.com/rs/zerolog/log"
	"github.com/vlourme/go-proxy/internal/config"
)

var (
	ErrNoLocalhost = errors.New("localhost is not allowed in non-debug mode")
	ErrNoIPFound   = errors.New("no IP address found")
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
func ResolveHostname(hostname string) (string, error) {
	cfg := config.Get()
	if isLocalhost(hostname) {
		if !cfg.DebugMode {
			return "", ErrNoLocalhost
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

	for _, addr := range addrs {
		if IsIPv6(addr) { // IPv6
			ip = addr
			break
		} else { // IPv4
			ip = addr
		}
	}

	if ip == "" {
		return "", ErrNoIPFound
	}

	if len(config.GetReplaceIPs()) > 0 {
		parsedIP := net.ParseIP(ip)
		for cidr, replacement := range config.GetReplaceIPs() {
			if cidr.Contains(parsedIP) {
				ip = replacement
				log.Info().Str("ip", ip).Str("hostname", hostname).Msg("DNS override found")
				break
			}
		}
	}

	if IsIPv6(ip) {
		ip = "[" + ip + "]"
	}

	dnsCache.Set(hostname, ip, CACHE_TTL)
	return ip, nil
}

func isLocalhost(hostname string) bool {
	return hostname == "localhost" || hostname == "127.0.0.1" || hostname == "::1" || hostname == "[::1]"
}
