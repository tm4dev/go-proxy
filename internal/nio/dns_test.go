package nio

import (
	"testing"

	"github.com/vlourme/go-proxy/internal/config"
)

func BenchmarkResolveHostname(b *testing.B) {
	domains := []string{
		"google.com",
		"cloudflare.com",
		"microsoft.com",
	}

	for _, domain := range domains {
		b.Run(domain, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				ip, err := ResolveHostname(domain, config.NetworkTypeIPv6)
				if err != nil {
					b.Fatalf("Failed to resolve %s with %s: %v", domain, config.NetworkTypeIPv6, err)
				}
				b.Logf("Resolved %s with %s: %s", domain, config.NetworkTypeIPv6, ip)
			}
		})
	}
}
