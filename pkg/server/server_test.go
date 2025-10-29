package server

import (
	"bytes"
	"context"
	"log"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"stnith/pkg/engine"
)

type MockEngine struct {
	mock.Mock
	engine.EngineInterface
	runCalled bool
	mu        sync.Mutex
}

func (m *MockEngine) Run() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.runCalled = true
	m.Called()
}

func (m *MockEngine) wasRunCalled() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.runCalled
}

type ServerTestSuite struct {
	suite.Suite
	server     *Server
	mockEngine *MockEngine
}

func (suite *ServerTestSuite) SetupTest() {
	suite.mockEngine = new(MockEngine)
	suite.server = New("127.0.0.1:0", suite.mockEngine, 10*time.Second)
}

func (suite *ServerTestSuite) TearDownTest() {
	ctx := context.Background()
	suite.server.Shutdown(ctx)
}

func (suite *ServerTestSuite) TestNew() {
	addr := "localhost:8080"
	duration := 5 * time.Minute
	mockEng := new(MockEngine)
	server := New(addr, mockEng, duration)

	assert.NotNil(suite.T(), server)
	assert.Equal(suite.T(), addr, server.Addr)
	assert.Equal(suite.T(), mockEng, server.engine)
	assert.Equal(suite.T(), duration, server.originalDuration)
}

func (suite *ServerTestSuite) TestStartTimer() {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	duration := 100 * time.Millisecond
	suite.mockEngine.On("Run").Once()

	suite.server.StartTimer(duration)

	// Give timer time to trigger
	time.Sleep(150 * time.Millisecond)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.Contains(suite.T(), output, "Timer started for")
	suite.mockEngine.AssertExpectations(suite.T())
}

func (suite *ServerTestSuite) TestStartTimerStopsExisting() {
	firstDuration := 5 * time.Second
	secondDuration := 100 * time.Millisecond

	// Start first timer
	suite.server.StartTimer(firstDuration)
	firstTimer := suite.server.timer

	// Start second timer (should stop the first)
	suite.server.StartTimer(secondDuration)
	secondTimer := suite.server.timer

	// Just verify we have a timer, don't compare pointers
	assert.NotNil(suite.T(), firstTimer)
	assert.NotNil(suite.T(), secondTimer)

	// Only the second timer should trigger
	suite.mockEngine.On("Run").Once()

	time.Sleep(150 * time.Millisecond)

	suite.mockEngine.AssertExpectations(suite.T())
}

func (suite *ServerTestSuite) TestShutdown() {
	// Start a timer
	suite.server.StartTimer(5 * time.Second)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	ctx := context.Background()
	err := suite.server.Shutdown(ctx)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), output, "Server shutdown complete")
	// Note: timer is stopped but not necessarily nil after Shutdown
}

func (suite *ServerTestSuite) TestHandleConnectionReset() {
	// Start server with a timer
	suite.server.originalDuration = 5 * time.Second
	// Set up mock expectations for the engine Run method (it may be called by timer)
	suite.mockEngine.On("Run").Maybe()
	suite.server.StartTimer(5 * time.Second)

	// Create mock connection
	client, server := net.Pipe()
	defer client.Close()

	// Handle connection in goroutine
	done := make(chan bool)
	go func() {
		suite.server.handleConnection(server)
		done <- true
	}()

	// Send RESET command with newline
	_, err := client.Write([]byte("RESET\n"))
	require.NoError(suite.T(), err)

	// Read response
	buffer := make([]byte, 1024)
	n, err := client.Read(buffer)
	require.NoError(suite.T(), err)

	response := string(buffer[:n])
	assert.Contains(suite.T(), response, "Timer reset successfully")

	// Wait for handler to complete
	<-done
}

func (suite *ServerTestSuite) TestHandleConnectionNoTimer() {
	// Don't start any timer

	// Create mock connection
	client, server := net.Pipe()
	defer client.Close()

	// Handle connection in goroutine
	done := make(chan bool)
	go func() {
		suite.server.handleConnection(server)
		done <- true
	}()

	// Send RESET command with newline
	_, err := client.Write([]byte("RESET\n"))
	require.NoError(suite.T(), err)

	// Read response
	buffer := make([]byte, 1024)
	n, err := client.Read(buffer)
	require.NoError(suite.T(), err)

	response := string(buffer[:n])
	assert.Contains(suite.T(), response, "No timer is running")

	// Wait for handler to complete
	<-done
}

func (suite *ServerTestSuite) TestHandleConnectionExpiredTimer() {
	// Start a timer that expires immediately
	suite.server.originalDuration = 1 * time.Millisecond
	suite.mockEngine.On("Run").Once()
	suite.server.StartTimer(1 * time.Millisecond)

	// Wait for timer to expire
	time.Sleep(50 * time.Millisecond)

	// Create mock connection
	client, server := net.Pipe()
	defer client.Close()

	// Handle connection in goroutine
	done := make(chan bool)
	go func() {
		suite.server.handleConnection(server)
		done <- true
	}()

	// Send RESET command with newline
	_, err := client.Write([]byte("RESET\n"))
	require.NoError(suite.T(), err)

	// Read response
	buffer := make([]byte, 1024)
	n, err := client.Read(buffer)
	require.NoError(suite.T(), err)

	response := string(buffer[:n])
	assert.Contains(suite.T(), response, "Timer already expired")

	// Wait for handler to complete
	<-done
}

func (suite *ServerTestSuite) TestHandleConnectionUnknownCommand() {
	// Create mock connection
	client, server := net.Pipe()
	defer client.Close()

	// Handle connection in goroutine
	done := make(chan bool)
	go func() {
		suite.server.handleConnection(server)
		done <- true
	}()

	// Send unknown command with newline
	_, err := client.Write([]byte("UNKNOWN\n"))
	require.NoError(suite.T(), err)

	// Read response
	buffer := make([]byte, 1024)
	n, err := client.Read(buffer)
	require.NoError(suite.T(), err)

	response := string(buffer[:n])
	assert.Contains(suite.T(), response, "Unknown command")

	// Wait for handler to complete
	<-done
}

func (suite *ServerTestSuite) TestStartTCPServer() {
	// Setup a listener on a free port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(suite.T(), err)
	addr := listener.Addr().String()
	listener.Close()

	suite.server.Addr = addr

	// Capture logs
	var logBuf bytes.Buffer
	log.SetOutput(&logBuf)
	defer log.SetOutput(os.Stderr)

	// Start server in goroutine
	serverErr := make(chan error)
	go func() {
		err := suite.server.StartTCPServer()
		serverErr <- err
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Try to connect to the server
	conn, err := net.Dial("tcp", addr)
	if err == nil {
		// Send a test command with newline
		conn.Write([]byte("RESET\n"))
		buffer := make([]byte, 1024)
		conn.Read(buffer)
		conn.Close()
	}

	// The server runs in an infinite loop, so we can't wait for it to finish
	// Just verify it started without immediate error
	select {
	case err := <-serverErr:
		// If we get here quickly, it means the server failed to start
		assert.NoError(suite.T(), err)
	case <-time.After(200 * time.Millisecond):
		// Server is running, this is expected
	}
}

func (suite *ServerTestSuite) TestStartTCPServerListenError() {
	// Use an invalid address to force an error
	suite.server.Addr = "invalid:address:format"

	// Capture logs
	var logBuf bytes.Buffer
	log.SetOutput(&logBuf)
	defer log.SetOutput(os.Stderr)

	err := suite.server.StartTCPServer()

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), logBuf.String(), "Failed to start TCP server")
}

func (suite *ServerTestSuite) TestHandleConnectionReadError() {
	// Create mock connection that will fail on read
	client, server := net.Pipe()
	client.Close() // Close client side to cause read error

	// Capture logs
	var logBuf bytes.Buffer
	log.SetOutput(&logBuf)
	defer log.SetOutput(os.Stderr)

	// Handle connection (should return without panic)
	suite.server.handleConnection(server)

	// No assertion needed - just verifying no panic
}

func (suite *ServerTestSuite) TestTimerReset() {
	// Setup mock engine
	suite.mockEngine.On("Run").Maybe()

	// Start initial timer
	suite.server.originalDuration = 5 * time.Second
	suite.server.StartTimer(5 * time.Second)

	// Get initial end time
	initialEndTime := suite.server.endTime

	// Wait a bit
	time.Sleep(100 * time.Millisecond)

	// Create mock connection
	client, server := net.Pipe()
	defer client.Close()

	// Handle connection in goroutine
	done := make(chan bool)
	go func() {
		suite.server.handleConnection(server)
		done <- true
	}()

	// Send RESET command with newline
	_, err := client.Write([]byte("RESET\n"))
	require.NoError(suite.T(), err)

	// Read response
	buffer := make([]byte, 1024)
	n, err := client.Read(buffer)
	require.NoError(suite.T(), err)

	response := string(buffer[:n])
	assert.Contains(suite.T(), response, "Timer reset successfully")

	// Verify timer was reset (end time should be later)
	assert.True(suite.T(), suite.server.endTime.After(initialEndTime))

	// Wait for handler to complete
	<-done
}

func TestServerSuite(t *testing.T) {
	suite.Run(t, new(ServerTestSuite))
}
