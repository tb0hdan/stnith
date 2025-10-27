package mac

import (
	"log"
)

// Disabler provides methods to detect and disable Mandatory Access Control systems
type Disabler struct {
	enableIt bool
}

// Disable attempts to disable MAC systems (SELinux, AppArmor)
func (d *Disabler) Disable() error {
	return d.platformDisable()
}

// Detect checks if any MAC systems are active
func (d *Disabler) Detect() ([]string, error) {
	return d.platformDetect()
}

// New creates a new MAC Disabler
func New(enableIt bool) *Disabler {
	if err := platformInit(); err != nil {
		log.Printf("Warning: failed to initialize MAC disabler: %v", err)
	}
	return &Disabler{
		enableIt: enableIt,
	}
}
