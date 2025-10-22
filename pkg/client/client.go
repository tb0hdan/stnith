package client

import (
	"fmt"
	"net"
	"os"
)

type Client struct {
	Addr string
}

func (c *Client) ResetTimer() {
	conn, err := net.Dial("tcp", c.Addr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to timer server at %s: %v\n", c.Addr, err)
		os.Exit(1)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to close connection: %v\n", err)
		}
	}()

	_, err = conn.Write([]byte("RESET"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to send reset command: %v\n", err)
		os.Exit(1)
	}

	response := make([]byte, 1024)
	n, err := conn.Read(response)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read response: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(string(response[:n]))
}

func New(addr string) *Client {
	return &Client{Addr: addr}
}
