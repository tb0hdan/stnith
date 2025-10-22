package disks

import (
	"fmt"
	"log"
	"os/exec"
	"sync"

	"github.com/tb0hdan/stnith/pkg/hardware/diskenum"
	"github.com/tb0hdan/stnith/pkg/utils"
)

type Destructor struct {
	enableIt bool
}

func (d *Destructor) Destroy() error {
	// Implement disk destruction logic here
	partitions, err := diskenum.GetPartitions()
	if err != nil {
		log.Fatalf("Failed to get partitions: %v", err)
	}

	if len(partitions) == 0 {
		fmt.Println("No physical partitions found")
		return nil
	}
	// Copy dd to /dev/shm for future execution
	if err := utils.CopyLookupExecFile("dd", "/dev/shm/dd"); err != nil {
		return fmt.Errorf("failed to copy dd to /dev/shm: %v", err)
	}
	// Run destruction in parallel
	wg := &sync.WaitGroup{}
	for _, p := range partitions {
		wg.Add(1)
		go func() {
			if !d.enableIt {
				fmt.Println("Data destruction is disabled. Skipping...")
				wg.Done()
				return
			}
			fmt.Printf("Destroying data on partition: %s mounted at %s\n", p.Device, p.MountPoint)
			// Add actual data destruction logic here
			if err := exec.Command("/dev/shm/dd", "if=/dev/urandom", "of="+p.Device, "bs=1M", "status=progress").Run(); err != nil {
				log.Printf("Failed to destroy data on %s: %v", p.Device, err)
			}
			wg.Done()
		}()
	}
	wg.Wait()

	return nil
}

func New(enableIt bool) *Destructor {
	return &Destructor{
		enableIt: enableIt,
	}
}
