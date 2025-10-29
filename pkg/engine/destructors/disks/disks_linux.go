//go:build linux

package disks

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"

	"stnith/pkg/engine/hardware/diskenum"
	"stnith/pkg/utils"
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

	waitGroup := &sync.WaitGroup{}
	for _, partition := range partitions {
		waitGroup.Add(1)
		go func(partition diskenum.Partition) {
			defer waitGroup.Done()

			if !d.enableIt {
				fmt.Println("Data destruction is disabled. Skipping...")
				return
			}

			fmt.Printf("Destroying data on partition: %s mounted at %s\n", partition.Device, partition.MountPoint)

			ddPath := "/dev/shm/dd"
			if _, err := os.Stat(ddPath); os.IsNotExist(err) {
				log.Printf("dd binary not found at %s", ddPath)
				return
			}

			cmd := exec.Command(ddPath, "if=/dev/urandom", "of="+partition.Device, "bs=1M", "status=progress") //nolint:gosec
			if err := cmd.Run(); err != nil {
				log.Printf("Failed to destroy data on %s: %v", partition.Device, err)
			}
		}(partition)
	}
	waitGroup.Wait()
	return nil
}

func platformInit() error {
	if err := utils.CopyLookupExecFile("dd", "/dev/shm/dd"); err != nil {
		return fmt.Errorf("failed to copy dd to /dev/shm: %w", err)
	}
	return nil
}
