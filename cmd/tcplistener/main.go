package main

import (
	"fmt"
	"io"
	"net"
)

func getLinesChannel(conn net.Conn) <-chan string {
	lines := make(chan string)

	go func() {
		defer close(lines)
		defer conn.Close()
		line := ""

		for {

			buff := make([]byte, 8)
			bytesRead, err := conn.Read(buff)
			if err != nil {
				if line != "" {
					lines <- line
				}
				if err != io.EOF {
					fmt.Printf("Error reading conn: %v\n", err)
				}
				return
			}
			for i := 0; i < bytesRead; i++ {
				if buff[i] == '\n' {
					lines <- line
					line = ""
					continue
				}
				line += string(buff[i])
			}
		}
	}()
	return lines
}

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
		linesChan := getLinesChannel(conn)
		go func() {
			for line := range linesChan {
				fmt.Printf("Received line from %v: %s\n", conn.RemoteAddr(), line)
			}
			fmt.Printf("Connection from %v closed\n", conn.RemoteAddr())
		}()
	}
}
