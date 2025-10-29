package utils

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
)

func CopyFile(src, dst string) error {
	// Open the source file
	sourceFile, err := os.Open(src) //nolint:gosec
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer func() {
		if closeErr := sourceFile.Close(); closeErr != nil {
			log.Printf("Warning: failed to close source file %s: %v", src, closeErr)
		}
	}()

	// Create the destination file
	destinationFile, err := os.Create(dst) //nolint:gosec
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer func() {
		if closeErr := destinationFile.Close(); closeErr != nil {
			log.Printf("Warning: failed to close destination file %s: %v", dst, closeErr)
		}
	}()

	// Copy the content
	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	// Flush file metadata to disk
	err = destinationFile.Sync()
	if err != nil {
		return fmt.Errorf("failed to sync destination file: %w", err)
	}

	return nil
}

func CopyExecFile(src, dst string) error {
	err := CopyFile(src, dst)
	if err != nil {
		return err
	}

	// Make the destination file executable (owner: read/write/execute, group/others: read/execute)
	const execFilePermissions = 0o755
	err = os.Chmod(dst, execFilePermissions)
	if err != nil {
		return fmt.Errorf("failed to make file executable: %w", err)
	}

	return nil
}

func CopyLookupExecFile(fileName, dst string) error {
	src, err := exec.LookPath(fileName)
	if err != nil {
		return fmt.Errorf("failed to find %s in PATH: %w", fileName, err)
	}
	return CopyExecFile(src, dst)
}
