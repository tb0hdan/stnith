//go:build darwin

package mac

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// platformDetect detects active MAC systems on macOS (SIP, Gatekeeper, TCC)
func (d *Disabler) platformDetect() ([]string, error) {
	var activeSystems []string

	// Check for System Integrity Protection (SIP)
	if isActive, err := detectSIP(); err == nil && isActive {
		activeSystems = append(activeSystems, "SIP")
	}

	// Check for Gatekeeper
	if isActive, err := detectGatekeeper(); err == nil && isActive {
		activeSystems = append(activeSystems, "Gatekeeper")
	}

	// Check for TCC (always present but we can detect if databases exist)
	if isActive, err := detectTCC(); err == nil && isActive {
		activeSystems = append(activeSystems, "TCC")
	}

	return activeSystems, nil
}

// platformDisable disables detected MAC systems on macOS
func (d *Disabler) platformDisable() error {
	if !d.enableIt {
		fmt.Println("MAC disabling is simulated. Enable it to actually disable MAC systems.")
		return nil
	}

	systems, err := d.platformDetect()
	if err != nil {
		return fmt.Errorf("failed to detect MAC systems: %w", err)
	}

	for _, system := range systems {
		switch system {
		case "SIP":
			fmt.Println("Warning: SIP cannot be disabled from userspace. Requires recovery mode and 'csrutil disable' command.")
			fmt.Println("Run: 'csrutil disable' from macOS Recovery Mode to disable SIP")
		case "Gatekeeper":
			if err := disableGatekeeper(); err != nil {
				return fmt.Errorf("failed to disable Gatekeeper: %w", err)
			}
			fmt.Println("Gatekeeper disabled")
		case "TCC":
			fmt.Println("Warning: TCC cannot be safely disabled without SIP disabled first.")
			fmt.Println("TCC databases are protected by SIP and require system-level access to modify.")
		}
	}

	return nil
}

// detectSIP checks if System Integrity Protection is enabled
// Note: csrutil is excluded as per requirements, using alternative detection
func detectSIP() (bool, error) {
	// Check if SIP-protected directories exist and are protected
	sipProtectedPaths := []string{
		"/System",
		"/usr/bin",
		"/usr/sbin",
		"/usr/libexec",
	}

	for _, path := range sipProtectedPaths {
		if _, err := os.Stat(path); err == nil {
			// Try to write a test file to see if SIP is protecting the directory
			testFile := path + "/.sip_test"
			if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
				// If we can't write, SIP is likely enabled
				return true, nil
			}
			// Clean up test file if we could write it
			os.Remove(testFile)
		}
	}

	// Alternative: check if SIP status file exists
	if _, err := os.Stat("/System/Library/Sandbox/rootless.conf"); err == nil {
		return true, nil
	}

	return false, nil
}

// detectGatekeeper checks if Gatekeeper is enabled
func detectGatekeeper() (bool, error) {
	cmd := exec.Command("spctl", "--status")
	output, err := cmd.Output()
	if err != nil {
		// If spctl command fails, assume Gatekeeper is not available
		return false, nil
	}

	status := strings.TrimSpace(string(output))
	// Gatekeeper is enabled if status is "assessments enabled"
	return strings.Contains(status, "assessments enabled"), nil
}

// detectTCC checks if TCC (Transparency, Consent, and Control) databases exist
func detectTCC() (bool, error) {
	tccPaths := []string{
		"/Library/Application Support/com.apple.TCC/TCC.db",
		os.Getenv("HOME") + "/Library/Application Support/com.apple.TCC/TCC.db",
	}

	for _, path := range tccPaths {
		if _, err := os.Stat(path); err == nil {
			return true, nil
		}
	}

	return false, nil
}

// disableGatekeeper disables Gatekeeper using spctl command
func disableGatekeeper() error {
	// Disable Gatekeeper globally
	cmd := exec.Command("spctl", "--master-disable")
	if err := cmd.Run(); err != nil {
		// Try alternative command
		cmd = exec.Command("spctl", "--global-disable")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to disable Gatekeeper: %w", err)
		}
	}

	return nil
}

func platformInit() error {
	// Check if we have necessary permissions
	if os.Geteuid() != 0 {
		fmt.Println("Warning: Running without root privileges. Some MAC operations may fail.")
	}
	return nil
}
