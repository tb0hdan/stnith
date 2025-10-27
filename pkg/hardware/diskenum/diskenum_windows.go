//go:build windows

package diskenum

import (
	"fmt"
	"log"

	"github.com/StackExchange/wmi"
)

type windowsEnumerator struct{}

func newEnumerator() DiskEnumerator {
	return &windowsEnumerator{}
}

// Win32_DiskDrive represents the WMI class for physical disk drives.
// We define the fields we are interested in.
type Win32_DiskDrive struct {
	Name   string
	Model  string
	Size   uint64 // Size in bytes
	Status string
}

func (we *windowsEnumerator) GetPartitions() ([]Partition, error) {
	var disks []Win32_DiskDrive
	query := "SELECT Name, Model, Size, Status FROM Win32_DiskDrive"

	err := wmi.Query(query, &disks)
	if err != nil {
		log.Fatalf("Failed to query WMI: %v", err)
	}

	if len(disks) == 0 {
		fmt.Println("No physical disk drives found.")
		return nil, nil
	}

	var partitions []Partition
	for _, disk := range disks {
		partitions = append(partitions, Partition{
			Device:     disk.Name,
			MountPoint: "", // Not applicable for physical drives
			Label:      disk.Model,
			Size:       disk.Size,
		})
	}

	return partitions, nil
}