//go:build windows

package rsync

import (
	"fmt"
	"os/exec"
	"path/filepath"
)

func (r *Rsync) platformRsync() error {
	if !r.enableIt {
		fmt.Printf("Rsync will be simulated. Would sync from %s to %s\n", r.src, r.dst)
		return nil
	}

	if r.src == "" || r.dst == "" {
		return fmt.Errorf("rsync source and destination must be specified")
	}

	// Convert Unix-style paths to Windows-style if needed
	src := filepath.FromSlash(r.src)
	dst := filepath.FromSlash(r.dst)

	// Use robocopy as rsync alternative on Windows
	// /MIR = mirror directory tree (equivalent to rsync --delete)
	// /R:3 = retry 3 times on failed copies
	// /W:10 = wait 10 seconds between retries
	// /V = verbose output
	cmd := exec.Command("robocopy", src, dst, "/MIR", "/R:3", "/W:10", "/V")
	output, err := cmd.CombinedOutput()

	// Robocopy returns exit codes that are not necessarily errors
	// 0 = No files copied, no failures
	// 1 = Files copied successfully
	// 2 = Extra files or directories detected and removed
	// 3 = Files copied and extra files removed
	// Values >= 8 indicate errors
	if cmd.ProcessState.ExitCode() >= 8 {
		return fmt.Errorf("robocopy failed with exit code %d: %s", cmd.ProcessState.ExitCode(), output)
	}

	fmt.Printf("Sync completed (robocopy): %s\n", output)
	return nil
}

func platformInit() error {
	// Check if robocopy is available (should be built into Windows)
	if _, err := exec.LookPath("robocopy"); err != nil {
		// If robocopy is not available, check for rsync (might be installed via WSL or third-party)
		if _, err := exec.LookPath("rsync"); err != nil {
			return fmt.Errorf("neither robocopy nor rsync found in PATH")
		}
	}
	return nil
}
