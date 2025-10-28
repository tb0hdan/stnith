//go:build unix

package permissions

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type PermissionsUnixTestSuite struct {
	suite.Suite
}

func (suite *PermissionsUnixTestSuite) TestIsAdmin() {
	// Get the actual effective user ID
	euid := os.Geteuid()

	// Test IsAdmin function
	result := IsAdmin()

	// Verify the result matches the expected value
	if euid == 0 {
		assert.True(suite.T(), result, "IsAdmin should return true when running as root (euid=0)")
	} else {
		assert.False(suite.T(), result, "IsAdmin should return false when not running as root (euid=%d)", euid)
	}
}

func (suite *PermissionsUnixTestSuite) TestIsAdminConsistency() {
	// Call IsAdmin multiple times to ensure consistency
	result1 := IsAdmin()
	result2 := IsAdmin()
	result3 := IsAdmin()

	assert.Equal(suite.T(), result1, result2, "IsAdmin should return consistent results")
	assert.Equal(suite.T(), result2, result3, "IsAdmin should return consistent results")
}

func (suite *PermissionsUnixTestSuite) TestIsAdminMatchesSystemState() {
	// This test verifies that IsAdmin correctly reflects the system state
	isRoot := os.Geteuid() == 0
	isAdminResult := IsAdmin()

	assert.Equal(suite.T(), isRoot, isAdminResult,
		"IsAdmin result should match actual euid check (euid=%d)", os.Geteuid())
}

func TestPermissionsUnixSuite(t *testing.T) {
	suite.Run(t, new(PermissionsUnixTestSuite))
}