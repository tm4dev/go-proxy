package main

import (
	"io"
	"log"
	"net/http"
)

// handleTunneling handles the HTTPS tunneling request
func handleTunneling(w http.ResponseWriter, r *http.Request, auth *ProxyAuth) (int, error) {
	ip, err := GetIP(auth.Parameters)
	if err != nil {
		log.Printf("Error getting IP: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return 500, err
	}

	dialer := GetDialer(ip)
	destConn, err := dialer.Dial("tcp6", r.Host)
	if err != nil {
		log.Printf("Error dialing: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return 500, err
	}

	w.WriteHeader(http.StatusOK)
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		log.Printf("Hijacking not supported")
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return 500, err
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		log.Printf("Error hijacking: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return 500, err
	}

	go transfer(destConn, clientConn)
	go transfer(clientConn, destConn)

	return 200, nil
}

// transfer copies data between two io.ReadCloser and io.WriteCloser
func transfer(destination io.WriteCloser, source io.ReadCloser) {
	defer destination.Close()
	defer source.Close()

	io.Copy(destination, source)
}
