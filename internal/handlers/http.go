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

// HandleHTTP handles the HTTP request
func HandleHTTP(w net.Conn, r *http.Request) int64 {
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

	if config.Get().HTTPClose {
		r.Header["Connection"] = []byte("close")
	}

	dialer, err := nio.GetDialer(
		params[auth.ParamSession],
		params[auth.ParamTimeout],
		params[auth.ParamLocation],
	)
	if err != nil {
		log.Error().Err(err).Msg("Error getting dialer")
		w.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))
		return -1
	}

	ip, err := nio.ResolveHostname(string(r.Host), config.Get().NetworkType)
	if err != nil {
		log.Error().Err(err).Msg("Error resolving hostname")
		w.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))
		return -1
	}

	destConn, err := dialer.Dial(
		string(config.Get().NetworkType),
		ip+":"+string(r.Port),
	)
	if err != nil {
		log.Error().Err(err).Msg("Error dialing")
		w.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))
		return -1
	}

	defer destConn.Close()
	_, err = r.WriteTo(destConn)
	if err != nil {
		log.Error().Err(err).Msg("Error writing request")
		w.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))
		return -1
	}

	return nio.CopyTimeout(w, destConn, time.Duration(config.Get().MaxTimeout)*time.Second)
}
