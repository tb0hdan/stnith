//go:build darwin

package poweroff

import (
	"fmt"
	"os/exec"
)

func (p *PowerOff) platformPowerOff() error {
	if !p.enableIt {
		fmt.Println("PowerOff will be simulated. Enable it to actually power off the system.")
		return nil
	}

	// On macOS, use the shutdown command
	// -h: halt after shutdown
	// now: shutdown immediately
	cmd := exec.Command("shutdown", "-h", "now")
	return cmd.Run()
}

func platformInit() error {
	// No special initialization needed for macOS
	return nil
}