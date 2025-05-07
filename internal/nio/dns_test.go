package nio

import (
	"testing"
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
				ip, err := ResolveHostname(domain)
				if err != nil {
					b.Fatalf("Failed to resolve %s: %v", domain, err)
				}
				b.Logf("Resolved %s: %s", domain, ip)
			}
		})
	}
}
