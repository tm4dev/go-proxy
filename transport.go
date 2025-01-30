package main

import (
	"net"
	"time"
)

// GetDialer returns a new net.Dialer with the given IP address
func GetDialer(ip string) *net.Dialer {
	return &net.Dialer{
		LocalAddr: &net.TCPAddr{
			IP: net.ParseIP(ip),
		},
		Timeout:   30 * time.Second,
		DualStack: false,
	}
}
