package utils

import (
	"net"
)

// GenerateIP generates a random IP address based on the given CIDR
func GenerateIP(cidr net.IPNet) (net.IP, error) {
	if cidr.IP.To4() != nil {
		return generateIPv4(cidr.IP, cidr.Mask), nil
	}

	return generateIPv6(cidr.IP, cidr.Mask), nil
}

// generateIPv4 generates a random IPv4 address within the given network
func generateIPv4(ip net.IP, mask net.IPMask) net.IP {
	// Convert to 4-byte representation
	ip = ip.To4()
	if ip == nil {
		return nil
	}

	// Create a new IP with the same network portion
	result := make(net.IP, len(ip))
	copy(result, ip)

	// Calculate the number of host bits
	ones, _ := mask.Size()
	hostBits := 32 - ones

	// Generate random bits only for the host portion
	for i := 0; i < hostBits/8; i++ {
		byteIndex := len(result) - 1 - i
		if byteIndex >= 0 {
			result[byteIndex] = byte(RandomInt(256))
		}
	}

	return result
}

// generateIPv6 generates a random IPv6 address within the given network
func generateIPv6(ip net.IP, mask net.IPMask) net.IP {
	// Create a new IP with the same network portion
	result := make(net.IP, len(ip))
	copy(result, ip)

	// Calculate the number of host bits
	ones, _ := mask.Size()
	hostBits := 128 - ones

	// Generate random bits only for the host portion
	for i := 0; i < hostBits/8; i++ {
		byteIndex := len(result) - 1 - i
		if byteIndex >= 0 {
			result[byteIndex] = byte(RandomInt(256))
		}
	}

	return result
}
