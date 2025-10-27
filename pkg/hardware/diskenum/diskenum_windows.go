//go:build windows

package diskenum

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

type windowsEnumerator struct{}

func newEnumerator() DiskEnumerator {
	return &windowsEnumerator{}
}

var (
	kernel32                 = syscall.NewLazyDLL("kernel32.dll")
	procGetLogicalDrives     = kernel32.NewProc("GetLogicalDrives")
	procGetVolumeInformation = kernel32.NewProc("GetVolumeInformationW")
	procGetDiskFreeSpaceEx   = kernel32.NewProc("GetDiskFreeSpaceExW")
	procGetDriveType         = kernel32.NewProc("GetDriveTypeW")
)

const (
	DRIVE_UNKNOWN     = 0
	DRIVE_NO_ROOT_DIR = 1
	DRIVE_REMOVABLE   = 2
	DRIVE_FIXED       = 3
	DRIVE_REMOTE      = 4
	DRIVE_CDROM       = 5
	DRIVE_RAMDISK     = 6
)

func (we *windowsEnumerator) GetPartitions() ([]Partition, error) {
	var partitions []Partition

	// Get logical drives bitmask
	drives, _, err := procGetLogicalDrives.Call()
	if err != nil && err.Error() != "The operation completed successfully." {
		return nil, fmt.Errorf("failed to get logical drives: %w", err)
	}

	// Iterate through each possible drive letter
	for i := 0; i < 26; i++ {
		if drives&(1<<uint(i)) != 0 {
			driveLetter := string(rune('A' + i))
			drivePath := driveLetter + ":\\"

			// Get drive type
			driveType := getDriveType(drivePath)

			// Only include fixed drives (hard disks) and removable drives
			if driveType != DRIVE_FIXED && driveType != DRIVE_REMOVABLE {
				continue
			}

			partition := Partition{
				Device:     driveLetter + ":",
				MountPoint: drivePath,
			}

			// Get volume information
			if err := fillVolumeInfo(&partition, drivePath); err == nil {
				// Get disk space information
				if err := fillDiskSpace(&partition, drivePath); err == nil {
					partitions = append(partitions, partition)
				}
			}
		}
	}

	return partitions, nil
}

func getDriveType(drivePath string) uint32 {
	drivePathPtr, _ := syscall.UTF16PtrFromString(drivePath)
	driveType, _, _ := procGetDriveType.Call(uintptr(unsafe.Pointer(drivePathPtr)))
	return uint32(driveType)
}

func fillVolumeInfo(partition *Partition, drivePath string) error {
	drivePathPtr, _ := syscall.UTF16PtrFromString(drivePath)

	var volumeName [256]uint16
	var fileSystemName [256]uint16
	var serialNumber uint32
	var maxComponentLength uint32
	var fileSystemFlags uint32

	ret, _, err := procGetVolumeInformation.Call(
		uintptr(unsafe.Pointer(drivePathPtr)),
		uintptr(unsafe.Pointer(&volumeName[0])),
		uintptr(len(volumeName)),
		uintptr(unsafe.Pointer(&serialNumber)),
		uintptr(unsafe.Pointer(&maxComponentLength)),
		uintptr(unsafe.Pointer(&fileSystemFlags)),
		uintptr(unsafe.Pointer(&fileSystemName[0])),
		uintptr(len(fileSystemName)),
	)

	if ret == 0 {
		return fmt.Errorf("failed to get volume information: %w", err)
	}

	partition.Label = syscall.UTF16ToString(volumeName[:])
	partition.FileSystem = syscall.UTF16ToString(fileSystemName[:])

	return nil
}

func fillDiskSpace(partition *Partition, drivePath string) error {
	drivePathPtr, _ := syscall.UTF16PtrFromString(drivePath)

	var freeBytesAvailable uint64
	var totalNumberOfBytes uint64
	var totalNumberOfFreeBytes uint64

	ret, _, err := procGetDiskFreeSpaceEx.Call(
		uintptr(unsafe.Pointer(drivePathPtr)),
		uintptr(unsafe.Pointer(&freeBytesAvailable)),
		uintptr(unsafe.Pointer(&totalNumberOfBytes)),
		uintptr(unsafe.Pointer(&totalNumberOfFreeBytes)),
	)

	if ret == 0 {
		return fmt.Errorf("failed to get disk space: %w", err)
	}

	partition.Size = totalNumberOfBytes
	partition.Available = freeBytesAvailable
	partition.Used = totalNumberOfBytes - totalNumberOfFreeBytes

	return nil
}

// getPhysicalDrives returns information about physical drives using WMI-like approach
// This is an alternative implementation that could be used for more detailed physical drive info
func getPhysicalDrives() ([]Partition, error) {
	// This would require additional Windows API calls or WMI queries
	// For now, we'll stick with the logical drives approach above
	// which is simpler and covers the most common use cases
	return nil, fmt.Errorf("physical drive enumeration not implemented")
}

// Additional utility functions for Windows-specific drive information

// isDriveReady checks if a drive is ready for access
func isDriveReady(drivePath string) bool {
	_, err := os.Stat(drivePath)
	return err == nil
}

// getVolumeSerial gets the volume serial number for a drive
func getVolumeSerial(drivePath string) (uint32, error) {
	drivePathPtr, _ := syscall.UTF16PtrFromString(drivePath)

	var serialNumber uint32
	var maxComponentLength uint32
	var fileSystemFlags uint32

	ret, _, err := procGetVolumeInformation.Call(
		uintptr(unsafe.Pointer(drivePathPtr)),
		0, // volumeName
		0, // volumeNameSize
		uintptr(unsafe.Pointer(&serialNumber)),
		uintptr(unsafe.Pointer(&maxComponentLength)),
		uintptr(unsafe.Pointer(&fileSystemFlags)),
		0, // fileSystemName
		0, // fileSystemNameSize
	)

	if ret == 0 {
		return 0, fmt.Errorf("failed to get volume serial: %w", err)
	}

	return serialNumber, nil
}

// getDriveGeometry could be implemented for more detailed physical drive information
// This would require additional Windows API calls to get sector size, cylinder count, etc.
func getDriveGeometry(drivePath string) error {
	// Implementation would go here using DeviceIoControl with IOCTL_DISK_GET_DRIVE_GEOMETRY
	return fmt.Errorf("drive geometry not implemented")
}
