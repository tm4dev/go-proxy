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

	// Buffered channel to prevent goroutine leaks
	done := make(chan int64, 2)

	go func() {
		n, _ := io.Copy(dst, src)
		// Close the write side to signal the other direction to stop
		if conn, ok := dst.(*net.TCPConn); ok {
			conn.CloseWrite()
		}
		done <- n
	}()

	go func() {
		n, _ := io.Copy(src, dst)
		// Close the write side to signal the other direction to stop
		if conn, ok := src.(*net.TCPConn); ok {
			conn.CloseWrite()
		}
		done <- n
	}()

	// Wait for both directions to finish
	n1 := <-done
	n2 := <-done

	return n1 + n2
}
