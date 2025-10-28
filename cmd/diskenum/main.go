package main

import (
	"fmt"
	"log"
	"os"
	"text/tabwriter"

	"github.com/tb0hdan/stnith/pkg/engine/hardware/diskenum"
)

const (
	tabWriterMinWidth = 2
	bytesPerGB        = 1024 * 1024 * 1024
)

func main() {
	partitions, err := diskenum.GetPartitions()
	if err != nil {
		log.Fatalf("Failed to get partitions: %v", err)
	}

	if len(partitions) == 0 {
		fmt.Println("No physical partitions found")
		return
	}

	tabWriter := tabwriter.NewWriter(os.Stdout, 0, 0, tabWriterMinWidth, ' ', 0)
	_, _ = fmt.Fprintln(tabWriter, "DEVICE\tMOUNT\tFILESYSTEM\tSIZE (GB)\tUSED (GB)\tAVAIL (GB)\tLABEL")
	_, _ = fmt.Fprintln(tabWriter, "------\t-----\t----------\t---------\t---------\t----------\t-----")

	for _, partition := range partitions {
		sizeGB := float64(partition.Size) / bytesPerGB
		usedGB := float64(partition.Used) / bytesPerGB
		availGB := float64(partition.Available) / bytesPerGB

		_, _ = fmt.Fprintf(tabWriter, "%s\t%s\t%s\t%.2f\t%.2f\t%.2f\t%s\n",
			partition.Device,
			partition.MountPoint,
			partition.FileSystem,
			sizeGB,
			usedGB,
			availGB,
			partition.Label)
	}

	_ = tabWriter.Flush()

	fmt.Printf("\nTotal partitions found: %d\n", len(partitions))
}
