package server

import (
	"bytes"
	"fmt"
	"http/internal/request"
	"http/internal/response"
	"io"
	"net"
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

type Handler func(w io.Writer, req *request.Request) *HandlerError

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

	request, err := request.RequestFromReader(conn)
	if err != nil {
		hErr := &HandlerError{
			StatusCode: response.StatusCode500,
			Message:    "Internal Server Error",
		}
		hErr.writeHandlerError(conn)
		return
	}
	fmt.Println("read request")

	buf := bytes.NewBuffer([]byte{})
	hErr := s.handler(buf, request)
	if hErr != nil {
		hErr.writeHandlerError(conn)
		return
	}
	b := buf.Bytes()
	response.WriteStatusLine(conn, response.StatusCode200)
	headers := response.GetDefaultHeaders(len(b))
	response.WriteHeaders(conn, headers)
	conn.Write(b)
	return
}

func (h *HandlerError) writeHandlerError(w io.Writer) error {
	err := response.WriteStatusLine(w, h.StatusCode)
	if err != nil {
		return err
	}
	headers := response.GetDefaultHeaders(len(h.Message))
	err = response.WriteHeaders(w, headers)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(h.Message))
	return err
}
