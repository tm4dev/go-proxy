package main

import (
	"crypto/tls"
	"io"
	"net/http"
	"time"
)

func handleHTTP(w http.ResponseWriter, r *http.Request, auth *ProxyAuth) (int, error) {
	ip, err := GetIP(auth.Parameters)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return 500, err
	}

	dialer := GetDialer(ip)

	transport := &http.Transport{
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
		Dial:                  dialer.Dial,
		DialContext:           dialer.DialContext,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   100,
		MaxConnsPerHost:       100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		DisableCompression:    false,
		ForceAttemptHTTP2:     false,
	}
	resp, err := transport.RoundTrip(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return 503, err
	}
	defer resp.Body.Close()

	// Copy headers
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)

	return resp.StatusCode, nil
}
