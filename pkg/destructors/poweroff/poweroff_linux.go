//go:build linux

package poweroff

import (
	"fmt"
	"syscall"
)

func (p *PowerOff) platformPowerOff() error {
	if !p.enableIt {
		fmt.Println("PowerOff will be simulated. Enable it to actually power off the system.")
		return nil
	}

	// Use Linux-specific syscall for power off
	return syscall.Reboot(syscall.LINUX_REBOOT_CMD_POWER_OFF)
}

func platformInit() error {
	// No special initialization needed for Linux
	return nil
}