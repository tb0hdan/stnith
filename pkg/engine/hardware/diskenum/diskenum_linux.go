//go:build linux

package diskenum

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

type linuxEnumerator struct{}

func newEnumerator() DiskEnumerator {
	return &linuxEnumerator{}
}

func (le *linuxEnumerator) GetPartitions() ([]Partition, error) {
	partitions := []Partition{}
	seenDevices := make(map[string]bool)

	file, err := os.Open("/proc/mounts")
	if err != nil {
		return nil, fmt.Errorf("failed to open /proc/mounts: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)

		const minExpectedFields = 4
		if len(fields) < minExpectedFields {
			continue
		}

		device := fields[0]
		mountPoint := fields[1]
		fsType := fields[2]

		if !isPhysicalDevice(device, fsType) {
			continue
		}

		// Skip if we've already seen this device
		if seenDevices[device] {
			continue
		}
		seenDevices[device] = true

		partition := Partition{
			Device:     device,
			MountPoint: mountPoint,
			FileSystem: fsType,
		}

		if err := fillPartitionStats(&partition); err == nil {
			partition.Label = getDeviceLabel(device)
			partitions = append(partitions, partition)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading /proc/mounts: %w", err)
	}

	return partitions, nil
}

func isPhysicalDevice(device, fsType string) bool {
	excludedFS := map[string]bool{
		"tmpfs":         true,
		"devtmpfs":      true,
		"sysfs":         true,
		"proc":          true,
		"devpts":        true,
		"securityfs":    true,
		"cgroup":        true,
		"cgroup2":       true,
		"pstore":        true,
		"bpf":           true,
		"autofs":        true,
		"debugfs":       true,
		"tracefs":       true,
		"fusectl":       true,
		"configfs":      true,
		"ramfs":         true,
		"hugetlbfs":     true,
		"mqueue":        true,
		"overlay":       true,
		"fuse":          true,
		"fuse.snapfuse": true,
	}

	if excludedFS[fsType] {
		return false
	}

	if strings.HasPrefix(device, "/dev/loop") {
		return false
	}

	if !strings.HasPrefix(device, "/dev/") {
		return false
	}

	if strings.Contains(device, "/dev/shm") {
		return false
	}

	return true
}

func fillPartitionStats(partition *Partition) error {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(partition.MountPoint, &stat); err != nil {
		return err
	}

	blockSize := stat.Bsize
	if blockSize < 0 {
		return fmt.Errorf("invalid block size: %d", blockSize)
	}
	blockSizeUint := uint64(blockSize)
	partition.Size = blockSizeUint * stat.Blocks
	partition.Available = blockSizeUint * stat.Bavail
	partition.Used = partition.Size - (blockSizeUint * stat.Bfree)

	return nil
}

func getDeviceLabel(device string) string {
	baseName := filepath.Base(device)

	labelPath := filepath.Join("/dev/disk/by-label")
	entries, err := os.ReadDir(labelPath)
	if err != nil {
		return ""
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		linkPath := filepath.Join(labelPath, entry.Name())
		target, err := os.Readlink(linkPath)
		if err != nil {
			continue
		}

		if filepath.Base(target) == baseName || strings.Contains(target, baseName) {
			return entry.Name()
		}
	}

	labelFile := fmt.Sprintf("/sys/class/block/%s/label", baseName)
	if data, err := os.ReadFile(labelFile); err == nil { //nolint:gosec
		return strings.TrimSpace(string(data))
	}

	return ""
}
