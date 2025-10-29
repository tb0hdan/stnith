package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"stnith/pkg/client"
	"stnith/pkg/engine"
	"stnith/pkg/engine/destructors"
	"stnith/pkg/engine/destructors/disks"
	"stnith/pkg/engine/destructors/poweroff"
	"stnith/pkg/engine/disablers"
	"stnith/pkg/engine/disablers/mac"
	"stnith/pkg/engine/failsafes"
	"stnith/pkg/engine/failsafes/process"
	"stnith/pkg/engine/savers"
	"stnith/pkg/engine/savers/rsync"
	"stnith/pkg/engine/savers/scriptdir"
	"stnith/pkg/server"
	"stnith/pkg/utils"
	"stnith/pkg/utils/permissions"
)

const (
	ShutdownTimeout = 5 * time.Second
)

var (
	durationFlag      string
	addrFlag          string
	resetFlag         bool
	wipeDiskFlag      bool
	enableIt          bool
	rsyncFlag         bool
	rsyncSrcFlag      string
	rsyncDstFlag      string
	scriptDirPathFlag string
)

func main() {
	flag.BoolVar(&enableIt, "enable-it", false, "I know what I'm doing. Enable it.")
	flag.StringVar(&durationFlag, "dms", "", "Timer duration (e.g., 1yr, 1mo, 1w, 2d, 3h, 30m, 45s)")
	flag.BoolVar(&wipeDiskFlag, "disks", false, "Wipe disks")
	flag.StringVar(&addrFlag, "addr", "localhost:11234", "TCP address to listen on or connect to")
	flag.BoolVar(&resetFlag, "reset", false, "Reset the timer by connecting to the server")
	flag.BoolVar(&rsyncFlag, "rsync", false, "Enable rsync functionality")
	flag.StringVar(&rsyncSrcFlag, "rsync-src", "", "Rsync source directory (has no effect if -rsync is disabled)")
	flag.StringVar(&rsyncDstFlag, "rsync-dst", "", "Rsync destination directory (has no effect if -rsync is disabled)")
	flag.StringVar(&scriptDirPathFlag, "saver-script-dir", "", "Directory containing executable scripts to run during save operation")
	flag.Parse()

	// Prepare client
	c := client.New(addrFlag)

	// Can be run with user permissions as it only connects to the server.
	if resetFlag {
		response, err := c.ResetTimer()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error resetting timer: %v\n", err)
			os.Exit(1)
		}
		fmt.Print(response)
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

	// Initialize failsafes - they will be triggered when timer expires to hide the process
	fss := make([]failsafes.FailsafeInterface, 0)
	processHider := process.New(enableIt)
	fss = append(fss, processHider)

	// Initialize disablers - they will be called when timer expires
	dis := make([]disablers.DisablerInterface, 0)
	macDisabler := mac.New(enableIt)
	dis = append(dis, macDisabler)

	// Initialize savers - they will be called when timer expires before destructors
	svs := make([]savers.SaverInterface, 0)
	if rsyncFlag {
		if enableIt {
			fmt.Println("Rsync is enabled.")
		} else {
			fmt.Println("Rsync will be simulated.")
		}

		if rsyncSrcFlag == "" || rsyncDstFlag == "" {
			fmt.Fprintf(os.Stderr, "Error: both -rsync-src and -rsync-dst must be specified when -rsync is enabled\n")
			os.Exit(1)
		}

		svs = append(svs, rsync.New(enableIt, rsyncSrcFlag, rsyncDstFlag))
	}

	if scriptDirPathFlag != "" {
		if enableIt {
			fmt.Println("Script directory execution is enabled.")
		} else {
			fmt.Println("Script directory execution will be simulated.")
		}

		svs = append(svs, scriptdir.New(enableIt, scriptDirPathFlag))
	}

	dds := make([]destructors.DestructorInterface, 0)
	if wipeDiskFlag {
		if enableIt {
			fmt.Println("Disk wiping is enabled.")
		} else {
			fmt.Println("Disk wiping will be simulated.")
		}

		dds = append(dds, disks.New(enableIt), poweroff.New(enableIt))
	}

	// Create the engine with all the components
	eng := engine.New(dis, dds, fss, svs)

	// Prepare and start server with the engine
	srv := server.New(addrFlag, eng, duration)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	fmt.Println("Starting timer...")
	srv.StartTimer(duration)

	go func() {
		if err := srv.StartTCPServerWithContext(ctx); err != nil && !errors.Is(err, context.Canceled) {
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
