package main

import (
	"fmt"
	"log"
	"os"
	"text/tabwriter"

	"github.com/tb0hdan/stnith/pkg/hardware/diskenum"
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

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "DEVICE\tMOUNT\tFILESYSTEM\tSIZE (GB)\tUSED (GB)\tAVAIL (GB)\tLABEL")
	_, _ = fmt.Fprintln(w, "------\t-----\t----------\t---------\t---------\t----------\t-----")

	for _, p := range partitions {
		sizeGB := float64(p.Size) / (1024 * 1024 * 1024)
		usedGB := float64(p.Used) / (1024 * 1024 * 1024)
		availGB := float64(p.Available) / (1024 * 1024 * 1024)

		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%.2f\t%.2f\t%.2f\t%s\n",
			p.Device,
			p.MountPoint,
			p.FileSystem,
			sizeGB,
			usedGB,
			availGB,
			p.Label)
	}

	_ = w.Flush()

	fmt.Printf("\nTotal partitions found: %d\n", len(partitions))
}
