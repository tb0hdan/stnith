package client

import (
	"fmt"
	"net"
	"os"
)

type Client struct {
	Addr string
}

func (c *Client) ResetTimer() (string, error) {
	conn, err := net.Dial("tcp", c.Addr)
	if err != nil {
		return "", fmt.Errorf("failed to connect to timer server at %s: %w", c.Addr, err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to close connection: %v\n", err)
		}
	}()

	_, err = conn.Write([]byte("RESET\n"))
	if err != nil {
		return "", fmt.Errorf("failed to send reset command: %w", err)
	}

	const responseBufferSize = 1024
	response := make([]byte, responseBufferSize)
	bytesRead, err := conn.Read(response)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return string(response[:bytesRead]), nil
}

func New(addr string) *Client {
	return &Client{Addr: addr}
}
