package server

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"stnith/pkg/engine"
)

type Server struct {
	Addr             string
	timerMutex       sync.Mutex
	timer            *time.Timer
	endTime          time.Time
	engine           engine.EngineInterface
	originalDuration time.Duration
	listener         net.Listener
	listenerMutex    sync.Mutex
	shutdownCh       chan struct{}
}

func (s *Server) StartTCPServer() error {
	return s.StartTCPServerWithContext(context.Background())
}

func (s *Server) StartTCPServerWithContext(ctx context.Context) error {
	listener, err := net.Listen("tcp", s.Addr)
	if err != nil {
		log.Printf("Failed to start TCP server: %v", err)
		return err
	}

	s.listenerMutex.Lock()
	s.listener = listener
	s.listenerMutex.Unlock()

	defer func() {
		if err := listener.Close(); err != nil {
			log.Printf("Failed to close listener: %v", err)
		}
	}()

	fmt.Printf("Listening for reset commands on %s\n", s.Addr)

	// Create a channel to receive accept results
	acceptCh := make(chan struct {
		conn net.Conn
		err  error
	})

	// Start goroutine to handle accepts
	go func() {
		for {
			conn, err := listener.Accept()
			select {
			case acceptCh <- struct {
				conn net.Conn
				err  error
			}{conn, err}:
			case <-ctx.Done():
				return
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-s.shutdownCh:
			return nil
		case result := <-acceptCh:
			if result.err != nil {
				log.Printf("Error accepting connection: %v", result.err)
				continue
			}
			go s.handleConnection(result.conn)
		}
	}
}

func (s *Server) StartTimer(duration time.Duration) {
	s.timerMutex.Lock()
	defer s.timerMutex.Unlock()

	s.endTime = time.Now().Add(duration)
	fmt.Printf("Timer started for %v (until %s)\n", duration, s.endTime.Format("2006-01-02 15:04:05"))

	if s.timer != nil {
		s.timer.Stop()
	}

	s.timer = time.AfterFunc(duration, s.engine.Run)
}

func (s *Server) Shutdown(ctx context.Context) error {
	// Signal shutdown (only close if not already closed)
	if s.shutdownCh != nil {
		select {
		case <-s.shutdownCh:
			// Channel already closed
		default:
			close(s.shutdownCh)
		}
	}

	// Close the listener to interrupt Accept()
	s.listenerMutex.Lock()
	if s.listener != nil {
		if err := s.listener.Close(); err != nil {
			log.Printf("Error closing listener: %v", err)
		}
	}
	s.listenerMutex.Unlock()

	// Stop the timer
	s.timerMutex.Lock()
	if s.timer != nil {
		s.timer.Stop()
	}
	s.timerMutex.Unlock()

	fmt.Println("Server shutdown complete.")
	return nil
}

func (s *Server) handleConnection(conn net.Conn) {
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("Failed to close connection: %v", err)
		}
	}()

	// Set a reasonable timeout for reading
	const readTimeout = 5 * time.Second
	if err := conn.SetReadDeadline(time.Now().Add(readTimeout)); err != nil {
		log.Printf("Failed to set read deadline: %v", err)
		return
	}

	// Use a limited reader to prevent excessive memory consumption
	const maxCommandLength = 256 // Maximum command length allowed
	limitedReader := io.LimitReader(conn, maxCommandLength+1)
	reader := bufio.NewReader(limitedReader)

	// Read until newline or max length
	commandBytes, err := reader.ReadBytes('\n')
	if err != nil && err != io.EOF {
		log.Printf("Error reading command: %v", err)
		if _, writeErr := conn.Write([]byte("Error reading command\n")); writeErr != nil {
			log.Printf("Failed to write error response: %v", writeErr)
		}
		return
	}

	// Check if command was too long (no newline found within limit)
	if len(commandBytes) > maxCommandLength {
		if _, err := conn.Write([]byte("Command too long\n")); err != nil {
			log.Printf("Failed to write response: %v", err)
		}
		return
	}

	// Trim whitespace and newlines
	command := strings.TrimSpace(string(commandBytes))

	// Validate command is not empty
	if command == "" {
		if _, err := conn.Write([]byte("Empty command\n")); err != nil {
			log.Printf("Failed to write response: %v", err)
		}
		return
	}

	// Process the command
	if command == "RESET" {
		s.timerMutex.Lock()
		defer s.timerMutex.Unlock()

		if s.timer != nil {
			// Stop timer first and check if it was successfully stopped
			if s.timer.Stop() {
				// Timer was successfully stopped, it hasn't fired yet
				remaining := s.originalDuration
				s.timer = time.AfterFunc(remaining, s.engine.Run)
				s.endTime = time.Now().Add(remaining)
				fmt.Printf("\nTimer reset! Remaining time: %v (until %s)\n", remaining.Round(time.Second), s.endTime.Format("2006-01-02 15:04:05"))
				if _, err := fmt.Fprintf(conn, "Timer reset successfully. Remaining: %v\n", remaining.Round(time.Second)); err != nil {
					log.Printf("Failed to write response: %v", err)
				}
			} else {
				// Timer already expired or is in the process of firing
				if _, err := conn.Write([]byte("Timer already expired\n")); err != nil {
					log.Printf("Failed to write response: %v", err)
				}
			}
		} else {
			if _, err := conn.Write([]byte("No timer is running\n")); err != nil {
				log.Printf("Failed to write response: %v", err)
			}
		}
	} else {
		if _, err := fmt.Fprintf(conn, "Unknown command: %s\n", command); err != nil {
			log.Printf("Failed to write response: %v", err)
		}
	}
}

func New(addr string, eng engine.EngineInterface, originalDuration time.Duration) *Server {
	return &Server{
		Addr:             addr,
		timerMutex:       sync.Mutex{},
		engine:           eng,
		originalDuration: originalDuration,
		shutdownCh:       make(chan struct{}),
	}
}
