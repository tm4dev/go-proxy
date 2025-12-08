package handlers

import (
	"net"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/vlourme/go-proxy/internal/auth"
	"github.com/vlourme/go-proxy/internal/config"
	"github.com/vlourme/go-proxy/internal/http"
	"github.com/vlourme/go-proxy/internal/nio"
)

// HandleTunneling handles the HTTPS tunneling request
func HandleTunneling(w net.Conn, r *http.Request) int64 {
	username, password, encodedParams := auth.GetCredentials(r)
	if username == "" {
		log.Error().Msg("No username provided")
		w.Write([]byte("HTTP/1.1 407 Proxy Authentication Required\r\nProxy-Authenticate: Basic\r\n\r\n"))
		return -1
	}

	if !auth.Verify(username, password) {
		log.Error().Msg("Invalid credentials")
		w.Write([]byte("HTTP/1.1 407 Proxy Authentication Required\r\nProxy-Authenticate: Basic\r\n\r\n"))
		return -1
	}

	params := auth.GetParams(encodedParams)

	ip, err := nio.ResolveHostname(string(r.Host))
	if err != nil {
		log.Error().Err(err).Msg("Error resolving hostname")
		w.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))
		return -1
	}

	dialer, err := nio.GetDialer(
		ip,
		params[auth.ParamSession],
		params[auth.ParamTimeout],
		params[auth.ParamLocation],
		params[auth.ParamFallback],
	)
	if err != nil {
		log.Error().Err(err).Msg("Error getting dialer")
		w.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))
		return -1
	}

	destConn, err := dialer.Dial("tcp", ip+":"+string(r.Port))
	if err != nil {
		log.Error().Err(err).Msg("Error dialing")
		w.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))
		return -1
	}
	defer destConn.Close()

	w.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
	return nio.CopyOnce(w, destConn, time.Duration(config.Get().MaxTimeout)*time.Second)
}
