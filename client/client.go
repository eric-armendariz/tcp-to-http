package main

import (
	"fmt"
	"net"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:42069")
	if err != nil {
		fmt.Printf("Error dialing: %v\n", err)
		return
	}
	defer conn.Close()

	messages := []string{
		"Hello, World!\n",
		"This is a test message.\n",
		"Goodbye!\n",
		"Test",
		" without newline",
		" does this work?\n",
	}

	for _, msg := range messages {
		_, err := conn.Write([]byte(msg))
		if err != nil {
			fmt.Printf("Error writing to connection: %v\n", err)
			return
		}
	}
}
