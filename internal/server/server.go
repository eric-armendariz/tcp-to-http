package server

import (
	"bytes"
	"fmt"
	"http/internal/headers"
	"http/internal/request"
	"http/internal/response"
	"net"
	"net/http"
	"strings"
	"sync/atomic"
)

type Server struct {
	port     int
	closed   atomic.Bool
	listener net.Listener
	handler  Handler
}

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

type Handler func(w *response.Writer, req *request.Request)

const httpbin = "/httpbin/"

func Serve(port int, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	s := &Server{port: port, listener: listener}
	s.closed.Store(false)
	s.handler = handler
	go s.listen()
	return s, nil
}

func (s *Server) Close() error {
	if err := s.listener.Close(); err != nil {
		return err
	}
	s.closed.Store(true)
	return nil
}

func (s *Server) listen() {
	for {
		if s.closed.Load() {
			return
		}
		conn, err := s.listener.Accept()
		if err != nil {
			fmt.Printf("Error accepting connection: %v\n", err)
			return
		}
		fmt.Printf("Accepted connection from %v\n", conn.RemoteAddr())
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	req, err := request.RequestFromReader(conn)
	if err != nil {
		fmt.Printf("Error reading request: %v\n", err)
		return
	}

	w := response.NewWriter(conn)
	// Proxy handler for httpbin requests.
	if strings.HasPrefix(req.RequestLine.RequestTarget, httpbin) {
		s.handleHTTPBin(w, req)
		return
	}
	s.handler(w, req)
	return
}

func (s *Server) handleHTTPBin(w *response.Writer, req *request.Request) {
	targetSuffix := strings.TrimPrefix(req.RequestLine.RequestTarget, httpbin)
	targetURL := fmt.Sprintf("https://httpbin.org/%s", targetSuffix)
	resp, err := http.Get(targetURL)
	if err != nil {
		fmt.Printf("Error making HTTP request: %v\n", err)
		w.WriteStatusLine(response.StatusCode500)
		w.WriteHeaders(response.GetDefaultHeaders(0))
		return
	}
	defer resp.Body.Close()
	w.WriteStatusLine(response.StatusCode200)
	h := headers.Headers{
		"content-type":      "text/plain",
		"connection":        "close",
		"transfer-encoding": "chunked",
	}
	w.WriteHeaders(h)
	buf := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buf)
		fmt.Printf("Read %d bytes from response body\n", n)
		if err != nil {
			break
		}
		var b bytes.Buffer
		fmt.Fprintf(&b, "%x\r\n%s\r\n", n, string(buf[:n]))
		encodedBytes := b.Bytes()
		w.WriteChunkedBody(encodedBytes)
	}
	w.WriteChunkedBodyDone()
}
