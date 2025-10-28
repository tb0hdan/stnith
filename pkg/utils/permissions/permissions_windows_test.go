//go:build windows

package permissions

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"golang.org/x/sys/windows"
)

type PermissionsWindowsTestSuite struct {
	suite.Suite
}

func (suite *PermissionsWindowsTestSuite) TestIsAdmin() {
	// Test that IsAdmin returns a boolean value
	result := IsAdmin()

	// This is a basic smoke test - we can't easily control whether
	// the test is running with elevated privileges or not
	assert.IsType(suite.T(), bool(false), result)
}

func (suite *PermissionsWindowsTestSuite) TestIsAdminConsistency() {
	// Call IsAdmin multiple times to ensure consistency
	result1 := IsAdmin()
	result2 := IsAdmin()
	result3 := IsAdmin()

	assert.Equal(suite.T(), result1, result2, "IsAdmin should return consistent results")
	assert.Equal(suite.T(), result2, result3, "IsAdmin should return consistent results")
}

func (suite *PermissionsWindowsTestSuite) TestIsAdminMatchesWindowsAPI() {
	// Directly check using Windows API for comparison
	token, err := windows.GetCurrentProcessToken()
	if err != nil {
		// If we can't get the token, IsAdmin should return false
		assert.False(suite.T(), IsAdmin(), "IsAdmin should return false when token cannot be obtained")
		return
	}
	defer token.Close()

	isElevated, err := token.IsElevated()
	expectedResult := false
	if err == nil {
		expectedResult = isElevated
	}

	// Compare with our IsAdmin function
	actualResult := IsAdmin()
	assert.Equal(suite.T(), expectedResult, actualResult,
		"IsAdmin result should match Windows API IsElevated result")
}

func (suite *PermissionsWindowsTestSuite) TestIsAdminHandlesErrors() {
	// This test verifies that IsAdmin handles errors gracefully
	// Since we can't easily simulate token errors, we just verify
	// that the function doesn't panic and returns a boolean
	defer func() {
		if r := recover(); r != nil {
			suite.T().Errorf("IsAdmin should not panic, but panicked with: %v", r)
		}
	}()

	result := IsAdmin()
	assert.IsType(suite.T(), bool(false), result, "IsAdmin should always return a boolean")
}

func TestPermissionsWindowsSuite(t *testing.T) {
	suite.Run(t, new(PermissionsWindowsTestSuite))
}