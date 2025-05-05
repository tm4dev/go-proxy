package nio

import (
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
)

// CopyTimeout copies data between two io.ReadCloser and io.WriteCloser
// with a timeout
func CopyTimeout(destination, source net.Conn, timeout time.Duration) int64 {
	dst := destination.(*net.TCPConn)
	src := source.(*net.TCPConn)

	src.SetDeadline(time.Now().Add(timeout))
	dst.SetDeadline(time.Now().Add(timeout))

	written, err := src.WriteTo(dst)
	if err != nil {
		log.Error().Err(err).Msg("Error transferring")
		return -1
	}

	return written
}

// CopyBidirectional copies data between two io.ReadCloser and io.WriteCloser
// in both directions
func CopyBidirectional(destination, source net.Conn, timeout time.Duration) int64 {
	defer destination.Close()
	defer source.Close()

	var written int64
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		atomic.AddInt64(&written, CopyTimeout(destination, source, timeout))
	}()

	go func() {
		defer wg.Done()
		atomic.AddInt64(&written, CopyTimeout(source, destination, timeout))
	}()

	wg.Wait()

	return written
}
