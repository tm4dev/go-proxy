package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"runtime"

	"github.com/libp2p/go-reuseport"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/vlourme/go-proxy/internal/config"
	"github.com/vlourme/go-proxy/internal/handlers"
	"github.com/vlourme/go-proxy/internal/sys"
)

func main() {
	config := config.Get()

	if config.DebugMode {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	}

	sys.TuneSysctl()
	for _, prefix := range config.BindPrefixes {
		if err := sys.AddRoute(prefix); err != nil {
			if err.Error() == "file exists" {
				log.Info().Str("prefix", prefix).Msg("Route already exists")
			} else {
				log.Error().Err(err).Str("prefix", prefix).Msg("Failed to add route")
			}
		}
	}

	addr := net.TCPAddr{
		IP:   net.ParseIP(config.ListenAddress),
		Port: int(config.ListenPort),
	}

	if config.TestPort > 0 {
		log.Info().Uint16("port", config.TestPort).Msg("Starting test server")
		go func() {
			server := &http.Server{
				Addr: fmt.Sprintf("[::]:%d", config.TestPort),
				Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Write([]byte("Hello, World!"))
				}),
			}
			if err := server.ListenAndServe(); err != nil {
				log.Error().Err(err).Msg("Failed to start server")
			}
		}()
	}

	log.Info().Int("count", runtime.NumCPU()).Str("address", addr.String()).Msg("Starting listeners")
	for idx := range runtime.NumCPU() {
		go func(idx int) {
			listener, err := reuseport.Listen("tcp", addr.String())
			if err != nil {
				log.Error().Err(err).Msg("Failed to create listener")
				return
			}

			for {
				conn, err := listener.Accept()
				if err != nil {
					log.Error().Err(err).Msg("Failed to accept connection")
				}

				go handlers.HandleConnection(idx, conn)
			}
		}(idx)
	}

	select {}
}
