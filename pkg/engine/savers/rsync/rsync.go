package rsync

import (
	"log"
)

type Rsync struct {
	enableIt bool
	src      string
	dst      string
}

func (r *Rsync) Save() error {
	return r.platformRsync()
}

func New(enableIt bool, src, dst string) *Rsync {
	if err := platformInit(); err != nil {
		log.Fatalf("failed to initialize platform-specific resources: %v", err)
	}
	return &Rsync{
		enableIt: enableIt,
		src:      src,
		dst:      dst,
	}
}
