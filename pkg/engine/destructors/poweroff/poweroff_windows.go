//go:build windows

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

	// On Windows, we use the shutdown command.
	// /s: shutdown the computer
	// /t 0: shutdown immediately
	cmd := exec.Command("shutdown", "/s", "/t", "0")
	return cmd.Run()
}

func platformInit() error {
	// No special initialization needed for Windows
	return nil
}
