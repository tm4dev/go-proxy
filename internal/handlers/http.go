package handlers

import (
	"bufio"
	"net"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/vlourme/go-proxy/internal/auth"
	"github.com/vlourme/go-proxy/internal/config"
	"github.com/vlourme/go-proxy/internal/http"
	"github.com/vlourme/go-proxy/internal/nio"
)

// HandleHTTP handles the HTTP request
func HandleHTTP(w net.Conn, buf *bufio.Reader, r *http.Request) int64 {
	username, password, encodedParams := auth.GetCredentials(r)
	if !auth.Verify(username, password) {
		log.Error().Msg("Invalid credentials")
		w.Write([]byte("HTTP/1.1 407 Proxy Authentication Required\r\n\r\n"))
		return -1
	}

	params := auth.GetParams(encodedParams)

	for _, header := range config.Get().DeletedHeaders {
		delete(r.Header, header)
	}

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
	_, err = r.WriteTo(destConn, buf)
	if err != nil {
		log.Error().Err(err).Msg("Error writing request")
		w.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))
		return -1
	}

	return nio.CopyOnce(w, destConn, time.Duration(config.Get().MaxTimeout)*time.Second)
}
