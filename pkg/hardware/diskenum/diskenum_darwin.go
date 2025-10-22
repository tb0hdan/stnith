//go:build darwin

package diskenum

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

type darwinEnumerator struct{}

func newEnumerator() DiskEnumerator {
	return &darwinEnumerator{}
}

func (de *darwinEnumerator) GetPartitions() ([]Partition, error) {
	partitions := []Partition{}

	cmd := exec.Command("mount")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute mount command: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		device := fields[0]
		mountPoint := fields[2]

		fsInfo := strings.TrimPrefix(fields[3], "(")
		fsInfo = strings.TrimSuffix(fsInfo, ")")
		fsParts := strings.Split(fsInfo, ",")
		fsType := ""
		if len(fsParts) > 0 {
			fsType = strings.TrimSpace(fsParts[0])
		}

		if !isPhysicalDeviceDarwin(device, fsType) {
			continue
		}

		partition := Partition{
			Device:     device,
			MountPoint: mountPoint,
			FileSystem: fsType,
		}

		if err := fillPartitionStatsDarwin(&partition); err == nil {
			partition.Label = getDeviceLabelDarwin(device)
			partitions = append(partitions, partition)
		}
	}

	return partitions, nil
}

func isPhysicalDeviceDarwin(device, fsType string) bool {
	excludedFS := map[string]bool{
		"devfs":   true,
		"autofs":  true,
		"fdesc":   true,
		"nullfs":  true,
		"vmhgfs":  true,
		"mtmfs":   true,
		"nfs":     true,
		"smbfs":   true,
		"afpfs":   true,
		"ftp":     true,
		"webdav":  true,
	}

	if excludedFS[fsType] {
		return false
	}

	if !strings.HasPrefix(device, "/dev/disk") {
		return false
	}

	if strings.Contains(device, "disk0s1") {
		return false
	}

	return true
}

func fillPartitionStatsDarwin(partition *Partition) error {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(partition.MountPoint, &stat); err != nil {
		return err
	}

	blockSize := uint64(stat.Bsize)
	partition.Size = blockSize * stat.Blocks
	partition.Available = blockSize * stat.Bavail
	partition.Used = partition.Size - (blockSize * stat.Bfree)

	return nil
}

func getDeviceLabelDarwin(device string) string {
	cmd := exec.Command("diskutil", "info", device)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "Volume Name:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				label := strings.TrimSpace(parts[1])
				if label != "" && label != "(null)" {
					return label
				}
			}
		}
	}

	return ""
}

func getDiskInfoDarwin() ([]Partition, error) {
	cmd := exec.Command("diskutil", "list", "-plist")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute diskutil list: %w", err)
	}

	cmd = exec.Command("plutil", "-convert", "json", "-", "-o", "-")
	cmd.Stdin = bytes.NewReader(output)
	jsonOutput, err := cmd.Output()
	if err != nil {
		return parseDiskutilText()
	}

	_ = jsonOutput

	return parseDiskutilText()
}

func parseDiskutilText() ([]Partition, error) {
	cmd := exec.Command("diskutil", "list")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute diskutil list: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	var partitions []Partition

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "/dev/disk") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}

		if fields[0] == "0:" || strings.HasPrefix(fields[0], "1:") ||
			strings.HasPrefix(fields[0], "2:") || strings.HasPrefix(fields[0], "3:") {

			sizeStr := fields[len(fields)-2] + fields[len(fields)-1]
			diskID := fields[len(fields)-3]

			if strings.Contains(diskID, "disk") {
				size := parseSizeString(sizeStr)

				var name string
				if len(fields) > 5 {
					name = strings.Join(fields[2:len(fields)-3], " ")
				} else {
					name = fields[2]
				}

				partition := Partition{
					Device:     "/dev/" + diskID,
					FileSystem: fields[1],
					Label:      name,
					Size:       size,
				}

				if !strings.Contains(fields[1], "Free") &&
					!strings.Contains(fields[1], "EFI") &&
					!strings.Contains(fields[1], "Recovery") {
					partitions = append(partitions, partition)
				}
			}
		}
	}

	return partitions, nil
}

func parseSizeString(sizeStr string) uint64 {
	sizeStr = strings.ReplaceAll(sizeStr, " ", "")

	multipliers := map[string]uint64{
		"B":  1,
		"KB": 1024,
		"MB": 1024 * 1024,
		"GB": 1024 * 1024 * 1024,
		"TB": 1024 * 1024 * 1024 * 1024,
	}

	for suffix, multiplier := range multipliers {
		if strings.HasSuffix(sizeStr, suffix) {
			numStr := strings.TrimSuffix(sizeStr, suffix)
			if val, err := strconv.ParseFloat(numStr, 64); err == nil {
				return uint64(val * float64(multiplier))
			}
		}
	}

	return 0
}