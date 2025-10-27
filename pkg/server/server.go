package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/tb0hdan/stnith/pkg/engine/destructors"
	"github.com/tb0hdan/stnith/pkg/engine/disablers"
	"github.com/tb0hdan/stnith/pkg/engine/failsafes"
)

type Server struct {
	Addr             string
	timerMutex       sync.Mutex
	timer            *time.Timer
	endTime          time.Time
	disablers        []disablers.Disabler
	destructors      []destructors.Destructor
	failsafes        []failsafes.Failsafe
	originalDuration time.Duration
}

func (s *Server) StartTCPServer() error {
	listener, err := net.Listen("tcp", s.Addr)
	if err != nil {
		log.Printf("Failed to start TCP server: %v", err)
		return err
	}
	defer func() {
		if err := listener.Close(); err != nil {
			log.Printf("Failed to close listener: %v", err)
		}
	}()

	fmt.Printf("Listening for reset commands on %s\n", s.Addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}

		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("Failed to close connection: %v", err)
		}
	}()

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return
	}

	command := string(buf[:n])
	if command == "RESET" {
		s.timerMutex.Lock()
		defer s.timerMutex.Unlock()

		if s.timer != nil {
			remaining := time.Until(s.endTime)
			if remaining > 0 {
				// Adjust remaining
				remaining = s.originalDuration
				// Reset timer
				s.timer.Stop()
				s.timer = time.AfterFunc(remaining, s.timerExpired)
				s.endTime = time.Now().Add(remaining)
				fmt.Printf("\nTimer reset! Remaining time: %v (until %s)\n", remaining.Round(time.Second), s.endTime.Format("2006-01-02 15:04:05"))
				if _, err := fmt.Fprintf(conn, "Timer reset successfully. Remaining: %v\n", remaining.Round(time.Second)); err != nil {
					log.Printf("Failed to write response: %v", err)
				}
			} else {
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
		if _, err := conn.Write([]byte("Unknown command\n")); err != nil {
			log.Printf("Failed to write response: %v", err)
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

	s.timer = time.AfterFunc(duration, s.timerExpired)
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.timerMutex.Lock()
	defer s.timerMutex.Unlock()

	if s.timer != nil {
		s.timer.Stop()
	}
	fmt.Println("Server shutdown complete.")
	return nil
}

func (s *Server) timerExpired() {
	fmt.Println("\nTimer expired!")

	// First, trigger failsafes to hide the process if configured
	if len(s.failsafes) > 0 {
		fmt.Println("Triggering failsafes...")
		for _, failsafe := range s.failsafes {
			if err := failsafe.Trigger(); err != nil {
				log.Printf("Failsafe error: %v", err)
			}
		}
		fmt.Println("All failsafes have been triggered.")
	}

	// Second, disable security systems if any disablers are configured
	if len(s.disablers) > 0 {
		fmt.Println("Calling disablers...")
		for _, disabler := range s.disablers {
			if err := disabler.Disable(); err != nil {
				log.Printf("Disabler error: %v", err)
			}
		}
		fmt.Println("All disablers have been called.")
	}

	// Finally, execute destructors
	if len(s.destructors) == 0 {
		fmt.Println("No destructors defined, exiting.")
		os.Exit(0)
	}
	fmt.Println("Calling destructors...")
	for _, destructor := range s.destructors {
		if err := destructor.Destroy(); err != nil {
			log.Printf("Destructor error: %v", err)
		}
	}
	fmt.Println("All destructors have been called. Good luck.")
	os.Exit(0)
}

func New(addr string, disablers []disablers.Disabler, destructors []destructors.Destructor, failsafes []failsafes.Failsafe, originalDuration time.Duration) *Server {
	return &Server{
		Addr:             addr,
		timerMutex:       sync.Mutex{},
		disablers:        disablers,
		destructors:      destructors,
		failsafes:        failsafes,
		originalDuration: originalDuration,
	}
}
