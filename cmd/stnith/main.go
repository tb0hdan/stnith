package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tb0hdan/stnith/pkg/client"
	"github.com/tb0hdan/stnith/pkg/destructors"
	"github.com/tb0hdan/stnith/pkg/destructors/disks"
	"github.com/tb0hdan/stnith/pkg/destructors/poweroff"
	"github.com/tb0hdan/stnith/pkg/disablers"
	"github.com/tb0hdan/stnith/pkg/disablers/mac"
	"github.com/tb0hdan/stnith/pkg/server"
	"github.com/tb0hdan/stnith/pkg/utils"
	"github.com/tb0hdan/stnith/pkg/utils/permissions"
)

const (
	ShutdownTimeout = 5 * time.Second
)

var (
	durationFlag string
	addrFlag     string
	resetFlag    bool
	wipeDiskFlag bool
	enableIt     bool
)

func main() {
	flag.BoolVar(&enableIt, "enable-it", false, "I know what I'm doing. Enable it.")
	flag.StringVar(&durationFlag, "dms", "", "Timer duration (e.g., 1yr, 1mo, 1w, 2d, 3h, 30m, 45s)")
	flag.BoolVar(&wipeDiskFlag, "disks", false, "Wipe disks")
	flag.StringVar(&addrFlag, "addr", "localhost:11234", "TCP address to listen on or connect to")
	flag.BoolVar(&resetFlag, "reset", false, "Reset the timer by connecting to the server")
	flag.Parse()

	// Prepare client
	c := client.New(addrFlag)

	// Can be run with user permissions as it only connects to the server.
	if resetFlag {
		c.ResetTimer()
		return
	}

	// Check if running as admin/root
	if !permissions.IsAdmin() {
		fmt.Fprintf(os.Stderr, "Error: This program must be run as admin/root\n")
		os.Exit(1)
	}

	if durationFlag == "" {
		fmt.Fprintf(os.Stderr, "Error: -dms flag is required to specify timer duration\n")
		fmt.Fprintf(os.Stderr, "Usage: %s -dms <duration> [-addr <address>]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "       %s -reset [-addr <address>]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nDuration examples: 1yr (1 year), 1mo (1 month), 1w (1 week), 2d (2 days), 3h (3 hours), 30m (30 minutes), 45s (45 seconds)\n")
		os.Exit(1)
	}

	duration, err := utils.ParseDuration(durationFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing duration: %v\n", err)
		os.Exit(1)
	}

	// Initialize disablers - they will be called when timer expires
	dis := make([]disablers.Disabler, 0)
	macDisabler := mac.New(enableIt)
	dis = append(dis, macDisabler)

	dds := make([]destructors.Destructor, 0)
	if wipeDiskFlag {
		if enableIt {
			fmt.Println("Disk wiping is enabled.")
		} else {
			fmt.Println("Disk wiping will be simulated.")
		}

		dds = append(dds, disks.New(enableIt), poweroff.New(enableIt))
	}
	// Prepare and start server
	srv := server.New(addrFlag, dis, dds, duration)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	fmt.Println("Starting timer...")
	srv.StartTimer(duration)

	go func() {
		if err := srv.StartTCPServer(); err != nil {
			log.Fatalf("Failed to start TCP server: %v", err)
		}
	}()

	<-ctx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}
}
