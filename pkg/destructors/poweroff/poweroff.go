package poweroff

import (
	"fmt"
	"os/exec"

	"github.com/tb0hdan/stnith/pkg/utils"
)

type Poweroff struct {
	enableIt bool
}

func (d *Poweroff) Destroy() error {
	if !d.enableIt {
		fmt.Println("Poweroff will be simulated. Enable it to actually power off the system.")
		return nil
	}
	fmt.Println("Powering off the system...")
	return exec.Command("/dev/shm/poweroff").Run()
}

func New(enableIt bool) *Poweroff {
	err := utils.CopyLookupExecFile("poweroff", "/dev/shm/poweroff")
	if err != nil {
		panic("failed to copy poweroff to /dev/shm: " + err.Error())
	}
	return &Poweroff{}
}
