package client

import (
	"bytes"
	"io"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

var stdoutMutex sync.Mutex

type ClientTestSuite struct {
	suite.Suite
	listener   net.Listener
	serverAddr string
	client     *Client
}

func (suite *ClientTestSuite) SetupTest() {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(suite.T(), err)
	suite.listener = listener
	suite.serverAddr = listener.Addr().String()
	suite.client = New(suite.serverAddr)
}

func (suite *ClientTestSuite) TearDownTest() {
	if suite.listener != nil {
		suite.listener.Close()
	}
}

func (suite *ClientTestSuite) TestNew() {
	addr := "localhost:8080"
	client := New(addr)

	assert.NotNil(suite.T(), client)
	assert.Equal(suite.T(), addr, client.Addr)
}

func (suite *ClientTestSuite) TestResetTimerSuccess() {
	// Protect stdout access with mutex
	stdoutMutex.Lock()
	defer stdoutMutex.Unlock()

	// Start mock server
	serverResponse := "Timer reset successfully\n"
	serverDone := make(chan bool)

	go func() {
		conn, err := suite.listener.Accept()
		require.NoError(suite.T(), err)
		defer conn.Close()

		// Read the RESET command
		buffer := make([]byte, 5)
		n, err := conn.Read(buffer)
		require.NoError(suite.T(), err)
		assert.Equal(suite.T(), "RESET", string(buffer[:n]))

		// Send response
		_, err = conn.Write([]byte(serverResponse))
		require.NoError(suite.T(), err)

		serverDone <- true
	}()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run client in goroutine to avoid blocking
	clientDone := make(chan bool)
	go func() {
		suite.client.ResetTimer()
		clientDone <- true
	}()

	// Wait for server to finish
	select {
	case <-serverDone:
	case <-time.After(2 * time.Second):
		suite.T().Fatal("Server timeout")
	}

	// Wait for client to finish or timeout
	select {
	case <-clientDone:
		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		assert.Equal(suite.T(), serverResponse, output)
	case <-time.After(2 * time.Second):
		w.Close()
		os.Stdout = oldStdout
		suite.T().Fatal("Client timeout")
	}
}

// Note: The following tests would normally test error conditions,
// but since the client calls os.Exit() on errors, we cannot test
// these conditions without refactoring the client code to return errors
// instead of exiting. In a production environment, the client should be
// refactored to return errors for better testability.

func (suite *ClientTestSuite) TestResetTimerIntegration() {
	// Protect stdout access with mutex
	stdoutMutex.Lock()
	defer stdoutMutex.Unlock()

	// This is a simple integration test that verifies basic connectivity
	serverDone := make(chan bool)
	clientDone := make(chan bool)

	go func() {
		conn, err := suite.listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		// Read command
		buffer := make([]byte, 1024)
		n, err := conn.Read(buffer)
		if err != nil {
			return
		}

		// Send appropriate response
		command := string(buffer[:n])
		if command == "RESET" {
			conn.Write([]byte("Timer reset\n"))
		} else {
			conn.Write([]byte("Unknown command\n"))
		}

		serverDone <- true
	}()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run client in goroutine
	go func() {
		suite.client.ResetTimer()
		w.Close()
		clientDone <- true
	}()

	// Wait for both server and client to complete
	serverCompleted := false
	clientCompleted := false
	timeout := time.After(2 * time.Second)

	for !serverCompleted || !clientCompleted {
		select {
		case <-serverDone:
			serverCompleted = true
		case <-clientDone:
			clientCompleted = true
		case <-timeout:
			w.Close()
			os.Stdout = oldStdout
			suite.T().Fatal("Test timeout")
			return
		}
	}

	// Restore stdout and read captured output
	os.Stdout = oldStdout
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Verify some output was received
	assert.NotEmpty(suite.T(), output, "Expected output from client but got none")
	assert.Contains(suite.T(), output, "Timer reset")
}

func TestClientSuite(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}