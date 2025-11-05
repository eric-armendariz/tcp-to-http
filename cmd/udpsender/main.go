package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	addr, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		fmt.Printf("Error resolving UDP address: %v\n", err)
		return
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		fmt.Printf("Error dialing UDP: %v\n", err)
		return
	}
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(">")
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading from stdin: %v\n", err)
			return
		}
		_, err = conn.Write([]byte(line))
		if err != nil {
			fmt.Printf("Error writing to UDP: %v\n", err)
			return
		}
	}
}
