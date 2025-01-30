package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"
)

var cidr string
var username = flag.String("username", "", "The username to use for authentication")
var password = flag.String("password", "", "The password to use for authentication")
var port = flag.String("port", "8080", "The port to listen on")

func main() {
	flag.Parse()
	cidr = flag.Arg(0)

	if cidr == "" {
		fmt.Println("Usage: proxy-server [flags] <ipv6-cidr>")
		fmt.Println("\nFlags:")
		flag.PrintDefaults()
		fmt.Println("\nExample:")
		fmt.Println("  proxy-server -port 8080 -username admin -password secret fd00::/64")
		return
	}

	// Create a proxy server
	proxy := &http.Server{
		Addr: ":" + *port,
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := NewProxyAuth()
			var err error

			if *username != "" && *password != "" {
				auth, err = parseProxyAuth(r.Header.Get("Proxy-Authorization"))
				if err != nil {
					http.Error(w, "Invalid proxy authentication", http.StatusBadRequest)
					return
				}

				if auth.Username != *username || auth.Password != *password {
					http.Error(w, "Invalid proxy authentication", http.StatusBadRequest)
					return
				}
			}

			now := time.Now()
			var status int

			if r.Method == http.MethodConnect {
				status, err = handleTunneling(w, r, auth) // HTTPS
			} else {
				status, err = handleHTTP(w, r, auth) // HTTP
			}

			if err != nil {
				log.Printf("[%s][%s][%d] %s -> %s: %s", r.RemoteAddr, time.Since(now), status, r.Method, r.Host, err)
			} else {
				log.Printf("[%s][%s][%d] %s -> %s", r.RemoteAddr, time.Since(now), status, r.Method, r.Host)
			}
		}),
	}

	fmt.Printf("Starting proxy server on :e%s\n", *port)
	if err := proxy.ListenAndServe(); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
