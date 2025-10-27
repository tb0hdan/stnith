//go:build windows

package disks

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
	"sync"

	"github.com/tb0hdan/stnith/pkg/hardware/diskenum"
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

			fmt.Printf("Destroying data on disk: %s\n", partition.Device)

			diskNumber := strings.TrimPrefix(partition.Device, "\\\\.\\PHYSICALDRIVE")
			diskpartScript := fmt.Sprintf("select disk %s\nclean all", diskNumber)
			cmd := exec.Command("diskpart", "/s")
			cmd.Stdin = strings.NewReader(diskpartScript)

			if output, err := cmd.CombinedOutput(); err != nil {
				log.Printf("Failed to destroy data on %s: %v\nOutput: %s", partition.Device, err, string(output))
			}
		}(p)
	}
	wg.Wait()
	return nil
}

func platformInit() error {
	return nil
}
