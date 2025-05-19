package nio

import (
	"io"
	"net"
	"time"
)

// CopyOnce copies data between two io.ReadCloser and io.WriteCloser
// in one direction
func CopyOnce(dst, src net.Conn, timeout time.Duration) int64 {
	src.SetDeadline(time.Now().Add(timeout))
	dst.SetDeadline(time.Now().Add(timeout))

	done := make(chan struct{})
	var n1, n2 int64

	go func() {
		n1, _ = io.Copy(dst, src)
		done <- struct{}{}
	}()

	go func() {
		n2, _ = io.Copy(src, dst)
		done <- struct{}{}
	}()

	// Wait for one direction to finish
	<-done

	return n1 + n2
}
