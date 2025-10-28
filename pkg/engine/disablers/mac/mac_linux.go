//go:build linux

package mac

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// platformDetect detects active MAC systems on Linux.
func (d *Disabler) platformDetect() ([]string, error) {
	var activeSystems []string

	// Check for SELinux
	if detectSELinux() {
		activeSystems = append(activeSystems, "SELinux")
	}

	// Check for AppArmor
	if detectAppArmor() {
		activeSystems = append(activeSystems, "AppArmor")
	}

	return activeSystems, nil
}

// platformDisable disables detected MAC systems on Linux.
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
		case "SELinux":
			if err := disableSELinux(); err != nil {
				return fmt.Errorf("failed to disable SELinux: %w", err)
			}
			fmt.Println("SELinux disabled")
		case "AppArmor":
			if err := disableAppArmor(); err != nil {
				return fmt.Errorf("failed to disable AppArmor: %w", err)
			}
			fmt.Println("AppArmor disabled")
		}
	}

	return nil
}

// detectSELinux checks if SELinux is active.
func detectSELinux() bool {
	// Check if SELinux is enabled via /sys/fs/selinux
	if _, err := os.Stat("/sys/fs/selinux"); err == nil {
		// Check SELinux status via getenforce
		cmd := exec.Command("getenforce")
		output, err := cmd.Output()
		if err != nil {
			// getenforce not available, check config file
			configPath := "/etc/selinux/config"
			if content, err := os.ReadFile(configPath); err == nil {
				if strings.Contains(string(content), "SELINUX=enforcing") ||
					strings.Contains(string(content), "SELINUX=permissive") {
					return true
				}
			}
			return false
		}
		status := strings.TrimSpace(string(output))
		return status == "Enforcing" || status == "Permissive"
	}
	return false
}

// disableSELinux disables SELinux.
func disableSELinux() error {
	// Set SELinux to permissive mode temporarily
	if err := exec.Command("setenforce", "0").Run(); err != nil {
		// Try alternative method if setenforce fails
		fmt.Printf("Warning: setenforce failed: %v\n", err)
	}

	// Disable SELinux permanently by modifying config
	configPath := "/etc/selinux/config"
	if content, err := os.ReadFile(configPath); err == nil {
		newContent := strings.ReplaceAll(string(content), "SELINUX=enforcing", "SELINUX=disabled")
		newContent = strings.ReplaceAll(newContent, "SELINUX=permissive", "SELINUX=disabled")
		const configFilePermissions = 0o644
		if err := os.WriteFile(configPath, []byte(newContent), configFilePermissions); err != nil {
			return fmt.Errorf("failed to update SELinux config: %w", err)
		}
	}

	return nil
}

// detectAppArmor checks if AppArmor is active.
func detectAppArmor() bool {
	// Check if AppArmor is loaded via /sys/kernel/security/apparmor
	if _, err := os.Stat("/sys/kernel/security/apparmor"); err == nil {
		// Check if any profiles are loaded
		profilesPath := "/sys/kernel/security/apparmor/profiles"
		if content, err := os.ReadFile(profilesPath); err == nil {
			return len(content) > 0
		}
		// Alternative: check via aa-status
		cmd := exec.Command("aa-status")
		if output, err := cmd.Output(); err == nil {
			return strings.Contains(string(output), "profiles are in enforce mode") ||
				strings.Contains(string(output), "profiles are in complain mode")
		}
	}
	return false
}

// disableAppArmor disables AppArmor.
func disableAppArmor() error {
	// Stop AppArmor service
	serviceCommands := [][]string{
		{"systemctl", "stop", "apparmor"},
		{"systemctl", "disable", "apparmor"},
		{"service", "apparmor", "stop"},
	}

	var lastErr error
	for _, cmdArgs := range serviceCommands {
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...) //nolint:gosec
		if err := cmd.Run(); err != nil {
			lastErr = err
			continue
		}
		// If any command succeeds, we're done
		lastErr = nil
		break
	}

	// Try to unload all AppArmor profiles
	profilesDir := "/etc/apparmor.d"
	if entries, err := os.ReadDir(profilesDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				profilePath := filepath.Join(profilesDir, entry.Name())
				// Use aa-complain to set profile to complain mode (less disruptive than removing)
				_ = exec.Command("aa-complain", profilePath).Run() //nolint:gosec
			}
		}
	}

	// Remove AppArmor from kernel command line (for permanent disable)
	grubPath := "/etc/default/grub"
	if content, err := os.ReadFile(grubPath); err == nil {
		if strings.Contains(string(content), "apparmor=1") {
			newContent := strings.ReplaceAll(string(content), "apparmor=1", "apparmor=0")
			const grubFilePermissions = 0o644
			if err := os.WriteFile(grubPath, []byte(newContent), grubFilePermissions); err == nil {
				// Update grub configuration
				_ = exec.Command("update-grub").Run()
				_ = exec.Command("update-grub2").Run()
			}
		}
	}

	return lastErr
}

func platformInit() error {
	// No special initialization needed for Linux
	return nil
}
