package diskenum

import "fmt"

type Partition struct {
	Device     string
	MountPoint string
	FileSystem string
	Size       uint64
	Used       uint64
	Available  uint64
	Label      string
}

func (p Partition) String() string {
	return fmt.Sprintf("Device: %s, Mount: %s, FS: %s, Size: %d, Label: %s",
		p.Device, p.MountPoint, p.FileSystem, p.Size, p.Label)
}

type DiskEnumerator interface {
	GetPartitions() ([]Partition, error)
}

func GetPartitions() ([]Partition, error) {
	enumerator := newEnumerator()
	return enumerator.GetPartitions()
}