//go:build !linux && !darwin && !windows

package process

func (p *ProcessHider) hideLinux() error {
	return nil
}

func (p *ProcessHider) hideDarwin() error {
	return nil
}

func (p *ProcessHider) hideWindows() error {
	return nil
}
