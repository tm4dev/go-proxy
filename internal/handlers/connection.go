package handlers

import (
	"bufio"
	"net"
	"net/http"

	"github.com/rs/zerolog/log"
)

func HandleConnection(workerId int, conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	for {
		req, err := http.ReadRequest(reader)
		if err != nil {
			break
		}

		var written int64
		if req.Method == http.MethodConnect {
			written = HandleTunneling(conn, req)
		} else {
			written = HandleHTTP(conn, req)
		}

		if written == -1 {
			log.Error().
				Int("worker_id", workerId).
				Str("method", req.Method).
				Str("url", req.RequestURI).
				Msg("Request failed")

			break
		} else {
			log.Trace().
				Int("worker_id", workerId).
				Str("method", req.Method).
				Str("url", req.RequestURI).
				Int64("written", written).
				Msg("Request handled")
		}
	}
}
