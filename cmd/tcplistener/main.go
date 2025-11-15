package main

import (
	"fmt"
	"net"

	"http/internal/request"
)

func main() {
	ln, err := net.Listen("tcp", ":42069")
	if err != nil {
		fmt.Printf("Error listening: %v\n", err)
		return
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Printf("Error accepting connection: %v\n", err)
			continue
		}
		fmt.Printf("Accepted connection from %v\n", conn.RemoteAddr())
		requestLine, err := request.RequestFromReader(conn)
		if err != nil {
			fmt.Printf("Error reading request: %v\n", err)
			continue
		}
		fmt.Printf("Request line:\n")
		fmt.Printf("Method: %s\n", requestLine.RequestLine.Method)
		fmt.Printf("Request target: %s\n", requestLine.RequestLine.RequestTarget)
		fmt.Printf("HTTP version: %s\n", requestLine.RequestLine.HttpVersion)
	}
}
