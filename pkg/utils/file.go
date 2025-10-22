package utils

import (
	"fmt"
	"io"
	"os"
	"os/exec"
)

func CopyFile(src, dst string) error {
	// Open the source file
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer func() {
		if err := sourceFile.Close(); err != nil {
			// Log error but don't fail the operation since we're in a defer
			_ = err
		}
	}()

	// Create the destination file
	destinationFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer func() {
		if err := destinationFile.Close(); err != nil {
			// Log error but don't fail the operation since we're in a defer
			_ = err
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

	// Make the destination file executable
	err = os.Chmod(dst, 0755)
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
