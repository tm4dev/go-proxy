package utils

import "net/url"

// GetHostAndPort returns the hostname and port of the URL.
func GetHostAndPort(url *url.URL) (string, string) {
	host := url.Hostname()
	port := url.Port()
	if port == "" {
		if url.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}

	return host, port
}
