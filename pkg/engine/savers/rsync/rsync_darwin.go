//go:build darwin

package rsync

import (
	"fmt"
	"os/exec"
)

func (r *Rsync) platformRsync() error {
	if !r.enableIt {
		fmt.Printf("Rsync will be simulated. Would sync from %s to %s\n", r.src, r.dst)
		return nil
	}

	if r.src == "" || r.dst == "" {
		return fmt.Errorf("rsync source and destination must be specified")
	}

	cmd := exec.Command("rsync", "-av", "--progress", r.src, r.dst)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("rsync failed: %w, output: %s", err, output)
	}

	fmt.Printf("Rsync completed: %s\n", output)
	return nil
}

func platformInit() error {
	// Check if rsync is available
	if _, err := exec.LookPath("rsync"); err != nil {
		return fmt.Errorf("rsync not found in PATH: %w", err)
	}
	return nil
}
