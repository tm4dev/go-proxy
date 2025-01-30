package main

import (
	"crypto/rand"
	"fmt"
	"net"
)

// GetIP returns the IP address for a given session
func GetIP(parameters map[string]string) (string, error) {
	session, err := GetSession(parameters)
	if err != nil {
		return "", err
	}

	if session != nil {
		return session.IP, nil
	}

	return GenerateIPv6FromCIDR(cidr)
}

// GenerateIPv6FromCIDR generates a random IPv6 address from a CIDR
func GenerateIPv6FromCIDR(cidr string) (string, error) {
	ip, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", fmt.Errorf("invalid CIDR: %w", err)
	}
	if ip.To4() != nil {
		return "", fmt.Errorf("not an IPv6 CIDR: %s", cidr)
	}

	prefixLen, _ := ipNet.Mask.Size()
	if prefixLen > 128 {
		return "", fmt.Errorf("invalid prefix length: %d", prefixLen)
	}

	// Get the network prefix
	prefix := make([]byte, net.IPv6len)
	copy(prefix, ipNet.IP)

	// Generate random bytes for the host portion
	hostBytes := make([]byte, net.IPv6len)
	_, err = rand.Read(hostBytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Apply network mask to preserve prefix
	for i := 0; i < net.IPv6len; i++ {
		prefix[i] |= hostBytes[i] & ^ipNet.Mask[i]
	}

	return net.IP(prefix).String(), nil
}
