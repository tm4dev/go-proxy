package handlers

import (
	"bufio"
	"net"
	"net/http"

	"github.com/rs/zerolog/log"
	httpParse "github.com/vlourme/go-proxy/internal/http"
)

func HandleConnection(workerId int, conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	if IsSocks(reader) {
		written := HandleSocks(conn, reader)
		if written == -1 {
			log.Error().
				Int("worker_id", workerId).
				Msg("Request failed")
		} else {
			log.Trace().
				Int("worker_id", workerId).
				Int64("written", written).
				Msg("Request handled")
		}

		return
	}

	for {
		req, err := httpParse.ParseRequest(reader)
		if err != nil {
			break
		}

		var written int64
		if string(req.Method) == http.MethodConnect {
			written = HandleTunneling(conn, req)
		} else {
			written = HandleHTTP(conn, reader, req)
		}

		req.Release()

		if written == -1 {
			log.Error().
				Int("worker_id", workerId).
				Str("url", string(req.URL)).
				Msg("Request failed")

			break
		} else {
			log.Trace().
				Int("worker_id", workerId).
				Str("url", string(req.URL)).
				Int64("written", written).
				Msg("Request handled")
		}
	}
}
