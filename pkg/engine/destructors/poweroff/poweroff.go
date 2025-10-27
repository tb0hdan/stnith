package poweroff

import (
	"log"
)

type PowerOff struct {
	enableIt bool
}

func (p *PowerOff) Destroy() error {
	return p.platformPowerOff()
}

func New(enableIt bool) *PowerOff {
	if err := platformInit(); err != nil {
		log.Fatalf("failed to initialize platform-specific resources: %v", err)
	}
	return &PowerOff{
		enableIt: enableIt,
	}
}
