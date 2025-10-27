package disks

import (
	"log"
	"time"
)

const (
	DiskSleep = 2 * time.Second
)

type Destructor struct {
	enableIt bool
}

func (d *Destructor) Destroy() error {
	if err := d.platformDestroy(); err != nil {
		return err
	}
	time.Sleep(DiskSleep)
	return nil
}

func New(enableIt bool) *Destructor {
	if err := platformInit(); err != nil {
		log.Fatalf("failed to initialize platform-specific resources: %v", err)
	}
	return &Destructor{
		enableIt: enableIt,
	}
}
