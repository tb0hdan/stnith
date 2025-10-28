//go:build darwin

package scriptdir

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
)

func (s *ScriptDir) platformExec() error {
	if !s.enableIt {
		fmt.Printf("ScriptDir execution will be simulated. Would execute scripts from %s\n", s.dir)
		return nil
	}

	if s.dir == "" {
		return fmt.Errorf("script directory must be specified")
	}

	// Check if directory exists
	if _, err := os.Stat(s.dir); os.IsNotExist(err) {
		return fmt.Errorf("script directory does not exist: %s", s.dir)
	}

	// List executable files in directory
	var scripts []string
	err := filepath.WalkDir(s.dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Check if file is executable
		info, err := d.Info()
		if err != nil {
			return err
		}

		if info.Mode()&0111 != 0 { // Check if any execute bit is set
			scripts = append(scripts, path)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to list scripts in directory %s: %w", s.dir, err)
	}

	if len(scripts) == 0 {
		fmt.Printf("No executable scripts found in directory %s\n", s.dir)
		return nil
	}

	// Sort scripts to ensure consistent execution order
	sort.Strings(scripts)

	fmt.Printf("Found %d executable script(s) in %s\n", len(scripts), s.dir)

	// Execute each script
	for _, script := range scripts {
		fmt.Printf("Executing script: %s\n", script)

		cmd := exec.Command(script)
		output, err := cmd.CombinedOutput()

		if err != nil {
			fmt.Printf("Script %s failed: %v, output: %s\n", script, err, output)
			// Continue with other scripts even if one fails
		} else {
			fmt.Printf("Script %s completed successfully: %s\n", script, output)
		}
	}

	return nil
}

func platformInit() error {
	// No special initialization required for Darwin
	return nil
}