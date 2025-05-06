package http

import (
	"bufio"
	"strings"
	"testing"
)

func TestParseRequest(t *testing.T) {
	request := "GET http://api.ipquery.io/?format=json HTTP/1.1\r\nHost: api.ipquery.io\r\nUser-Agent: curl/8.5.0\r\nAccept: */*\r\n\r\n"
	reader := bufio.NewReader(strings.NewReader(request))
	req, err := ParseRequest(reader)
	defer req.Release()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if string(req.Method) != "GET" {
		t.Fatalf("expected method GET, got %s", req.Method)
	}

	if string(req.Host) != "api.ipquery.io" {
		t.Fatalf("expected Host api.ipquery.io, got %s", req.Host)
	}

	if string(req.Port) != "80" {
		t.Fatalf("expected Port 80, got %s", req.Port)
	}

	if string(req.URL) != "http://api.ipquery.io/?format=json" {
		t.Fatalf("expected URL http://api.ipquery.io/?format=json, got %s", req.URL)
	}

	if string(req.Version) != "HTTP/1.1" {
		t.Fatalf("expected version HTTP/1.1, got %s", req.Version)
	}

	if string(req.Header["Host"]) != "api.ipquery.io" {
		t.Fatalf("expected Host header api.ipquery.io, got %s", req.Header["Host"])
	}

	if string(req.Header["User-Agent"]) != "curl/8.5.0" {
		t.Fatalf("expected User-Agent header curl/8.5.0, got %s", req.Header["User-Agent"])
	}

	if string(req.Header["Accept"]) != "*/*" {
		t.Fatalf("expected Accept header */*, got %s", req.Header["Accept"])
	}
}

func TestParseRequestConnect(t *testing.T) {
	request := "CONNECT example.com:443 HTTP/1.1\r\nHost: example.com\r\nConnection: close\r\n\r\n"
	reader := bufio.NewReader(strings.NewReader(request))
	req, err := ParseRequest(reader)
	defer req.Release()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if string(req.Host) != "example.com" {
		t.Fatalf("expected Host example.com, got %s", req.Host)
	}

	if string(req.Port) != "443" {
		t.Fatalf("expected Port 443, got %s", req.Port)
	}

	if string(req.Method) != "CONNECT" {
		t.Fatalf("expected method CONNECT, got %s", req.Method)
	}

	if string(req.URL) != "example.com:443" {
		t.Fatalf("expected URL example.com:443, got %s", req.URL)
	}

	if string(req.Version) != "HTTP/1.1" {
		t.Fatalf("expected version HTTP/1.1, got %s", req.Version)
	}

	if string(req.Header["Connection"]) != "close" {
		t.Fatalf("expected Connection header close, got %s", req.Header["Connection"])
	}
}

func BenchmarkParseRequest(b *testing.B) {
	request := "GET http://api.ipquery.io/?format=json HTTP/1.1\r\nHost: api.ipquery.io\r\nUser-Agent: curl/8.5.0\r\nAccept: */*\r\n\r\n"
	for i := 0; i < b.N; i++ {
		reader := bufio.NewReader(strings.NewReader(request))
		req, err := ParseRequest(reader)
		if err != nil {
			b.Fatalf("expected no error, got %v", err)
		}
		req.Release()
	}
}
