package http

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"sync"

	"github.com/rs/zerolog/log"
)

type Request struct {
	Host    []byte
	Port    []byte
	Method  []byte
	URL     []byte
	Version []byte
	Header  map[string][]byte
}

var requestPool = sync.Pool{
	New: func() any {
		return &Request{
			Header: make(map[string][]byte, 16),
		}
	},
}

func getRequest() *Request {
	req := requestPool.Get().(*Request)
	for k := range req.Header {
		delete(req.Header, k)
	}
	return req
}

func (req *Request) WriteTo(w io.Writer, src *bufio.Reader) (total int64, err error) {
	buf := bytes.NewBuffer(nil)
	buf.Write(req.Method)
	buf.Write([]byte(" "))
	buf.Write(req.URL)
	buf.Write([]byte(" "))
	buf.Write(req.Version)
	buf.Write([]byte("\r\n"))

	for k, v := range req.Header {
		buf.Write([]byte(k))
		buf.Write([]byte(": "))
		buf.Write(v)
		buf.Write([]byte("\r\n"))
	}

	buf.Write([]byte("\r\n"))

	total, err = buf.WriteTo(w)
	if err != nil {
		return total, err
	}

	if clStr, ok := req.Header["Content-Length"]; ok {
		log.Info().Str("content-length", string(clStr)).Msg("Writing body")
		cl, err := strconv.Atoi(string(clStr))
		if err != nil {
			return total, fmt.Errorf("invalid Content-Length: %v", err)
		}

		log.Info().Msg("more to write")
		n, err := io.CopyN(w, src, int64(cl))
		if err != nil {
			return total, err
		}
		total += n
	}

	return total, nil
}

func (req *Request) Release() {
	requestPool.Put(req)
}

func ParseRequest(r *bufio.Reader) (*Request, error) {
	line, err := r.ReadSlice('\n')
	if err != nil {
		return nil, err
	}
	line = bytes.TrimSpace(line)

	// METHOD
	method, line, found := bytes.Cut(line, []byte(" "))
	if !found {
		return nil, fmt.Errorf("invalid request line")
	}

	// URL
	url, version, found := bytes.Cut(line, []byte(" "))
	if !found {
		return nil, fmt.Errorf("invalid request line")
	}

	req := getRequest()
	req.Method = method
	// If buffer > 4096, the slice will move and the URL will be invalid
	// This is kept only for logging purposes, it works without but
	// the URL in the log will be invalid due to the buffer re-use.
	req.URL = bytes.Clone(url)
	req.Version = version

	req.Host, req.Port, err = extractHostPort(req.Method, req.URL)
	if err != nil {
		return nil, err
	}

	for {
		line, err := r.ReadSlice('\n')
		if err != nil {
			return nil, err
		}

		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			break
		}

		key, value, found := bytes.Cut(line, []byte(":"))
		if !found {
			continue
		}

		req.Header[string(key)] = bytes.TrimSpace(value)
	}

	return req, nil
}

func extractHostPort(method, rawURL []byte) ([]byte, []byte, error) {
	if bytes.Equal(method, []byte("CONNECT")) {
		host, port, found := bytes.Cut(rawURL, []byte(":"))
		if !found {
			return host, []byte("443"), nil
		}

		return host, port, nil
	}

	// Strip scheme manually for GET/POST/...
	const httpPrefix = "http://"
	const httpsPrefix = "https://"

	var host []byte
	var defaultPort []byte

	switch {
	case bytes.HasPrefix(rawURL, []byte(httpPrefix)):
		defaultPort = []byte("80")
		raw := rawURL[len(httpPrefix):]
		slash := bytes.IndexByte(raw, '/')
		if slash == -1 {
			host = raw
		} else {
			host = raw[:slash]
		}

	case bytes.HasPrefix(rawURL, []byte(httpsPrefix)):
		defaultPort = []byte("443")
		raw := rawURL[len(httpsPrefix):]
		slash := bytes.IndexByte(raw, '/')
		if slash == -1 {
			host = raw
		} else {
			host = raw[:slash]
		}

	default:
		return nil, nil, fmt.Errorf("invalid absolute URL in request line: %s", rawURL)
	}

	portIdx := bytes.LastIndex(host, []byte(":"))
	if portIdx == -1 {
		return host, defaultPort, nil
	}

	port := host[portIdx+1:]
	host = host[:portIdx]

	return host, port, nil
}
