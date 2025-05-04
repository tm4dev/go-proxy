package handlers

import (
	"net"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/vlourme/go-proxy/internal/auth"
	"github.com/vlourme/go-proxy/internal/config"
	"github.com/vlourme/go-proxy/internal/nio"
	"github.com/vlourme/go-proxy/internal/utils"
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
		r.Header.Del(header)
	}

	dialer, err := nio.GetDialer(params[auth.ParamSession], params[auth.ParamTimeout])
	if err != nil {
		log.Error().Err(err).Msg("Error getting dialer")
		w.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))
		return -1
	}

	host, port := utils.GetHostAndPort(r.URL)
	ip, err := nio.ResolveHostname(host, config.Get().NetworkType)
	if err != nil {
		log.Error().Err(err).Msg("Error resolving hostname")
		w.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))
		return -1
	}

	destConn, err := dialer.Dial(
		string(config.Get().NetworkType),
		ip+":"+port,
	)
	if err != nil {
		log.Error().Err(err).Msg("Error dialing")
		w.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))
		return -1
	}

	defer destConn.Close()
	r.Write(destConn)
	return nio.CopyTimeout(w, destConn, 15*time.Second)
}
