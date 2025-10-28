package scriptdir

import (
	"log"
)

type ScriptDir struct {
	enableIt bool
	dir      string
}

func (s *ScriptDir) Save() error {
	return s.platformExec()
}

func New(enableIt bool, dir string) *ScriptDir {
	if err := platformInit(); err != nil {
		log.Fatalf("failed to initialize platform-specific resources: %v", err)
	}
	return &ScriptDir{
		enableIt: enableIt,
		dir:      dir,
	}
}