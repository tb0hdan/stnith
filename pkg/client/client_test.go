package client

import (
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
	_, w, _ := os.Pipe()
	os.Stdout = w

	// Run client in goroutine to avoid blocking
	clientDone := make(chan bool)
	clientErr := make(chan error, 1)
	var clientResponse string
	go func() {
		response, err := suite.client.ResetTimer()
		if err != nil {
			clientErr <- err
		} else {
			clientResponse = response
		}
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

		// Check if there was an error
		select {
		case err := <-clientErr:
			suite.T().Fatalf("Client error: %v", err)
		default:
		}

		// Verify the response
		assert.Equal(suite.T(), serverResponse, clientResponse)
	case <-time.After(2 * time.Second):
		w.Close()
		os.Stdout = oldStdout
		suite.T().Fatal("Client timeout")
	}
}

func (suite *ClientTestSuite) TestResetTimerIntegration() {
	// This integration test verifies basic connectivity and proper error handling
	serverDone := make(chan bool)
	clientDone := make(chan bool)
	clientErr := make(chan error, 1)
	var clientResponse string

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

	// Run client in goroutine
	go func() {
		response, err := suite.client.ResetTimer()
		if err != nil {
			clientErr <- err
		} else {
			clientResponse = response
		}
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
			suite.T().Fatal("Test timeout")
			return
		}
	}

	// Check for client errors
	select {
	case err := <-clientErr:
		suite.T().Fatalf("Client error: %v", err)
	default:
	}

	// Verify the response
	assert.NotEmpty(suite.T(), clientResponse, "Expected response from client but got none")
	assert.Contains(suite.T(), clientResponse, "Timer reset")
}

func TestClientSuite(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}
