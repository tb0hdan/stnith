//go:build darwin

package disks

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/tb0hdan/stnith/pkg/hardware/diskenum"
	"github.com/tb0hdan/stnith/pkg/utils"
)

func (d *Destructor) platformDestroy() error {
	partitions, err := diskenum.GetPartitions()
	if err != nil {
		return fmt.Errorf("failed to get partitions: %w", err)
	}

	if len(partitions) == 0 {
		fmt.Println("No physical partitions found")
		return nil
	}

	wg := &sync.WaitGroup{}
	for _, p := range partitions {
		wg.Add(1)
		go func(partition diskenum.Partition) {
			defer wg.Done()

			if !d.enableIt {
				fmt.Println("Data destruction is disabled. Skipping...")
				return
			}

			fmt.Printf("Destroying data on partition: %s mounted at %s\n", partition.Device, partition.MountPoint)

			tmpDir := os.TempDir()
			ddPath := filepath.Join(tmpDir, "dd")
			if _, err := os.Stat(ddPath); os.IsNotExist(err) {
				log.Printf("dd binary not found at %s", ddPath)
				return
			}

			device := convertToRawDevice(partition.Device)

			cmd := exec.Command(ddPath, "if=/dev/urandom", "of="+device, "bs=1048576", "status=progress")
			if err := cmd.Run(); err != nil {
				log.Printf("Failed to destroy data on %s: %v", device, err)
			}
		}(p)
	}
	wg.Wait()
	return nil
}

func platformInit() error {
	tmpDir := os.TempDir()
	ddPath := filepath.Join(tmpDir, "dd")

	if err := utils.CopyLookupExecFile("dd", ddPath); err != nil {
		return fmt.Errorf("failed to copy dd to %s: %w", ddPath, err)
	}
	return nil
}

func convertToRawDevice(device string) string {
	if strings.HasPrefix(device, "/dev/disk") {
		return strings.Replace(device, "/dev/disk", "/dev/rdisk", 1)
	}
	return device
}
