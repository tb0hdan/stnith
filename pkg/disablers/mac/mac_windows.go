//go:build windows

package mac

import (
	"fmt"
	"golang.org/x/sys/windows/registry"
	"os"
	"os/exec"
	"strings"
)

// platformDetect detects active MAC systems on Windows
func (d *Disabler) platformDetect() ([]string, error) {
	var activeSystems []string

	// Check for User Account Control (UAC)
	if isActive, err := detectUAC(); err == nil && isActive {
		activeSystems = append(activeSystems, "UAC")
	}

	// Check for Windows Defender Application Control (WDAC)
	if isActive, err := detectWDAC(); err == nil && isActive {
		activeSystems = append(activeSystems, "WDAC")
	}

	// Check for AppLocker
	if isActive, err := detectAppLocker(); err == nil && isActive {
		activeSystems = append(activeSystems, "AppLocker")
	}

	return activeSystems, nil
}

// platformDisable disables detected MAC systems on Windows
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
		case "UAC":
			if err := disableUAC(); err != nil {
				return fmt.Errorf("failed to disable UAC: %w", err)
			}
			fmt.Println("UAC disabled. A reboot is required for changes to take effect.")
		case "WDAC":
			fmt.Println("Warning: Disabling WDAC is a complex process and requires specific boot commands.")
			fmt.Println("Refer to Microsoft documentation for disabling Windows Defender Application Control.")
		case "AppLocker":
			if err := disableAppLocker(); err != nil {
				return fmt.Errorf("failed to disable AppLocker: %w", err)
			}
			fmt.Println("AppLocker disabled.")
		}
	}

	return nil
}

// detectUAC checks if User Account Control is enabled
func detectUAC() (bool, error) {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\System`, registry.QUERY_VALUE)
	if err != nil {
		return false, err
	}
	defer key.Close()

	val, _, err := key.GetIntegerValue("EnableLUA")
	if err != nil {
		if err == registry.ErrNotExist {
			return true, nil // Assuming default enabled
		}
		return false, err
	}

	return val != 0, nil
}

// disableUAC disables User Account Control by setting registry key
func disableUAC() error {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\System`, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer key.Close()

	return key.SetDWordValue("EnableLUA", 0)
}

// detectWDAC checks for signs of Windows Defender Application Control
func detectWDAC() (bool, error) {
	// Check for WDAC policy files
	policyDir := os.ExpandEnv("${SystemRoot}\\System32\\CodeIntegrity\\CiPolicies\\Active")
	if _, err := os.Stat(policyDir); err == nil {
		files, err := os.ReadDir(policyDir)
		if err != nil {
			return false, err
		}
		if len(files) > 0 {
			return true, nil
		}
	}
	return false, nil
}

// detectAppLocker checks if the AppLocker service is running
func detectAppLocker() (bool, error) {
	cmd := exec.Command("sc", "query", "AppIDSvc")
	output, err := cmd.Output()
	if err != nil {
		// If the service doesn't exist, it's not active
		return false, nil
	}

	return strings.Contains(string(output), "RUNNING"), nil
}

// disableAppLocker stops and disables the AppLocker service
func disableAppLocker() error {
	if err := exec.Command("sc", "stop", "AppIDSvc").Run(); err != nil {
		// Ignore error if service is not running
	}
	if err := exec.Command("sc", "config", "AppIDSvc", "start=", "disabled").Run(); err != nil {
		return fmt.Errorf("failed to disable AppLocker service: %w", err)
	}
	return nil
}

func platformInit() error {
	// Check for administrator privileges
	_, err := os.Open("\\\\.\\PHYSICALDRIVE0")
	if err != nil {
		fmt.Println("Warning: Running without administrator privileges. Some MAC operations may fail.")
	}
	return nil
}
