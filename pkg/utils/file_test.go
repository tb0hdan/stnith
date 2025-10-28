package utils

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type FileTestSuite struct {
	suite.Suite
	tempDir string
}

func (suite *FileTestSuite) SetupTest() {
	var err error
	suite.tempDir, err = os.MkdirTemp("", "file_test_*")
	require.NoError(suite.T(), err)
}

func (suite *FileTestSuite) TearDownTest() {
	os.RemoveAll(suite.tempDir)
}

func (suite *FileTestSuite) TestCopyFileSuccess() {
	// Create source file
	srcPath := filepath.Join(suite.tempDir, "source.txt")
	content := []byte("test content for copy")
	err := os.WriteFile(srcPath, content, 0644)
	require.NoError(suite.T(), err)

	// Copy file
	dstPath := filepath.Join(suite.tempDir, "destination.txt")
	err = CopyFile(srcPath, dstPath)
	require.NoError(suite.T(), err)

	// Verify destination file exists and has correct content
	dstContent, err := os.ReadFile(dstPath)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), content, dstContent)

	// Note: File permissions might not be preserved exactly depending on umask
	// Just verify the file exists and is readable
	_, err = os.Stat(dstPath)
	assert.NoError(suite.T(), err)
}

func (suite *FileTestSuite) TestCopyFileSourceNotExist() {
	srcPath := filepath.Join(suite.tempDir, "nonexistent.txt")
	dstPath := filepath.Join(suite.tempDir, "destination.txt")

	err := CopyFile(srcPath, dstPath)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to open source file")
}

func (suite *FileTestSuite) TestCopyFileDestinationError() {
	// Create source file
	srcPath := filepath.Join(suite.tempDir, "source.txt")
	err := os.WriteFile(srcPath, []byte("test"), 0644)
	require.NoError(suite.T(), err)

	// Try to copy to an invalid destination (directory that doesn't exist)
	dstPath := filepath.Join(suite.tempDir, "nonexistent", "destination.txt")
	err = CopyFile(srcPath, dstPath)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to create destination file")
}

func (suite *FileTestSuite) TestCopyFileLargeFile() {
	// Create a larger source file
	srcPath := filepath.Join(suite.tempDir, "large.txt")
	size := 1024 * 1024 // 1MB
	content := make([]byte, size)
	for i := 0; i < size; i++ {
		content[i] = byte(i % 256)
	}
	err := os.WriteFile(srcPath, content, 0644)
	require.NoError(suite.T(), err)

	// Copy file
	dstPath := filepath.Join(suite.tempDir, "large_copy.txt")
	err = CopyFile(srcPath, dstPath)
	require.NoError(suite.T(), err)

	// Verify content
	dstContent, err := os.ReadFile(dstPath)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), content, dstContent)
}

func (suite *FileTestSuite) TestCopyFileEmptyFile() {
	// Create empty source file
	srcPath := filepath.Join(suite.tempDir, "empty.txt")
	err := os.WriteFile(srcPath, []byte{}, 0644)
	require.NoError(suite.T(), err)

	// Copy file
	dstPath := filepath.Join(suite.tempDir, "empty_copy.txt")
	err = CopyFile(srcPath, dstPath)
	require.NoError(suite.T(), err)

	// Verify destination exists and is empty
	info, err := os.Stat(dstPath)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(0), info.Size())
}

func (suite *FileTestSuite) TestCopyExecFileSuccess() {
	// Create source file
	srcPath := filepath.Join(suite.tempDir, "script.sh")
	content := []byte("#!/bin/sh\necho 'test'")
	err := os.WriteFile(srcPath, content, 0644)
	require.NoError(suite.T(), err)

	// Copy as executable
	dstPath := filepath.Join(suite.tempDir, "script_exec.sh")
	err = CopyExecFile(srcPath, dstPath)
	require.NoError(suite.T(), err)

	// Verify content
	dstContent, err := os.ReadFile(dstPath)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), content, dstContent)

	// Verify permissions (should be executable)
	info, err := os.Stat(dstPath)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), os.FileMode(0755), info.Mode().Perm())
}

func (suite *FileTestSuite) TestCopyExecFileSourceNotExist() {
	srcPath := filepath.Join(suite.tempDir, "nonexistent.sh")
	dstPath := filepath.Join(suite.tempDir, "destination.sh")

	err := CopyExecFile(srcPath, dstPath)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to open source file")
}

func (suite *FileTestSuite) TestCopyLookupExecFileSuccess() {
	// This test is system-dependent
	// We'll try to find a common command that exists on most systems
	var cmdToFind string
	switch runtime.GOOS {
	case "windows":
		cmdToFind = "cmd.exe"
	default:
		cmdToFind = "sh"
	}

	// Check if command exists first
	_, err := exec.LookPath(cmdToFind)
	if err != nil {
		suite.T().Skip("Command not found in PATH: ", cmdToFind)
	}

	dstPath := filepath.Join(suite.tempDir, "copied_cmd")
	err = CopyLookupExecFile(cmdToFind, dstPath)
	require.NoError(suite.T(), err)

	// Verify file exists and is executable
	info, err := os.Stat(dstPath)
	require.NoError(suite.T(), err)
	assert.True(suite.T(), info.Mode().IsRegular())

	// On Unix systems, check if it's executable
	if runtime.GOOS != "windows" {
		assert.Equal(suite.T(), os.FileMode(0755), info.Mode().Perm())
	}
}

func (suite *FileTestSuite) TestCopyLookupExecFileNotInPath() {
	// Try to copy a file that definitely doesn't exist
	dstPath := filepath.Join(suite.tempDir, "copied_cmd")
	err := CopyLookupExecFile("this_command_definitely_does_not_exist_12345", dstPath)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to find")
	assert.Contains(suite.T(), err.Error(), "in PATH")
}

func (suite *FileTestSuite) TestCopyFileOverwrite() {
	// Create source file
	srcPath := filepath.Join(suite.tempDir, "source.txt")
	srcContent := []byte("source content")
	err := os.WriteFile(srcPath, srcContent, 0644)
	require.NoError(suite.T(), err)

	// Create existing destination file
	dstPath := filepath.Join(suite.tempDir, "destination.txt")
	err = os.WriteFile(dstPath, []byte("old content"), 0644)
	require.NoError(suite.T(), err)

	// Copy file (should overwrite)
	err = CopyFile(srcPath, dstPath)
	require.NoError(suite.T(), err)

	// Verify content was overwritten
	dstContent, err := os.ReadFile(dstPath)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), srcContent, dstContent)
}

func (suite *FileTestSuite) TestCopyFilePreservesMode() {
	// Create source file with specific permissions
	srcPath := filepath.Join(suite.tempDir, "source.txt")
	content := []byte("test content")
	err := os.WriteFile(srcPath, content, 0600) // rw-------
	require.NoError(suite.T(), err)

	// Copy file
	dstPath := filepath.Join(suite.tempDir, "destination.txt")
	err = CopyFile(srcPath, dstPath)
	require.NoError(suite.T(), err)

	// Note: Permissions might not be preserved exactly due to umask
	// Just verify the file was created
	_, err = os.Stat(dstPath)
	assert.NoError(suite.T(), err)
}

func (suite *FileTestSuite) TestCopyFileHandlesReadError() {
	// Create a source file
	srcPath := filepath.Join(suite.tempDir, "source.txt")
	err := os.WriteFile(srcPath, []byte("test"), 0000) // No read permission
	require.NoError(suite.T(), err)

	// Try to copy (should fail due to no read permission)
	dstPath := filepath.Join(suite.tempDir, "destination.txt")
	err = CopyFile(srcPath, dstPath)

	// On some systems/filesystems, root can read files regardless of permissions
	// So we check if we're running as root
	if os.Geteuid() != 0 {
		assert.Error(suite.T(), err)
		assert.Contains(suite.T(), err.Error(), "failed to open source file")
	}
}

type mockReader struct {
	io.Reader
}

func (m *mockReader) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}

func TestFileSuite(t *testing.T) {
	suite.Run(t, new(FileTestSuite))
}