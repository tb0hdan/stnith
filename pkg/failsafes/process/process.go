package process

import (
	"os"
	"runtime"

	"github.com/tb0hdan/stnith/pkg/failsafes"
)

type ProcessHider struct {
	enabled bool
	pid     int
}

func New(enabled bool) failsafes.Failsafe {
	return &ProcessHider{
		enabled: enabled,
		pid:     os.Getpid(),
	}
}

func (p *ProcessHider) Trigger() error {
	if !p.enabled {
		return nil
	}

	switch runtime.GOOS {
	case "linux":
		return p.hideLinux()
	case "darwin":
		return p.hideDarwin()
	case "windows":
		return p.hideWindows()
	default:
		return nil
	}
}
