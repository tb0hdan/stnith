package poweroff

import (
	"fmt"
	"syscall"
)

type Poweroff struct {
	enableIt bool
}

func (d *Poweroff) Destroy() error {
	if !d.enableIt {
		fmt.Println("Poweroff will be simulated. Enable it to actually power off the system.")
		return nil
	}
	return syscall.Reboot(syscall.LINUX_REBOOT_CMD_POWER_OFF)
}

func New(enableIt bool) *Poweroff {
	return &Poweroff{
		enableIt: enableIt,
	}
}
