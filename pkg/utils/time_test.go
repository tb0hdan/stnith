package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type TimeTestSuite struct {
	suite.Suite
}

func (suite *TimeTestSuite) TestParseDurationSeconds() {
	duration, err := ParseDuration("45s")
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 45*time.Second, duration)
}

func (suite *TimeTestSuite) TestParseDurationMinutes() {
	duration, err := ParseDuration("30m")
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 30*time.Minute, duration)
}

func (suite *TimeTestSuite) TestParseDurationHours() {
	duration, err := ParseDuration("2h")
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 2*time.Hour, duration)
}

func (suite *TimeTestSuite) TestParseDurationDays() {
	duration, err := ParseDuration("3d")
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 3*24*time.Hour, duration)
}

func (suite *TimeTestSuite) TestParseDurationWeeks() {
	duration, err := ParseDuration("2w")
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 2*7*24*time.Hour, duration)
}

func (suite *TimeTestSuite) TestParseDurationMonths() {
	duration, err := ParseDuration("3mo")
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 3*30*24*time.Hour, duration)
}

func (suite *TimeTestSuite) TestParseDurationYears() {
	duration, err := ParseDuration("1yr")
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 365*24*time.Hour, duration)
}

func (suite *TimeTestSuite) TestParseDurationMultipleDigits() {
	duration, err := ParseDuration("100s")
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 100*time.Second, duration)

	duration, err = ParseDuration("999h")
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 999*time.Hour, duration)
}

func (suite *TimeTestSuite) TestParseDurationInvalidFormat() {
	testCases := []string{
		"",
		"abc",
		"10",
		"s",
		"10x",
		"10ss",
		"10 s",
		" 10s",
		"10s ",
		"10.5s",
		"-10s",
		"10ms",
		"10mos",
		"10yrs",
	}

	for _, tc := range testCases {
		_, err := ParseDuration(tc)
		assert.Error(suite.T(), err, "Expected error for input: %s", tc)
		assert.Contains(suite.T(), err.Error(), "invalid duration format")
	}
}

func (suite *TimeTestSuite) TestParseDurationZeroValue() {
	duration, err := ParseDuration("0s")
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), time.Duration(0), duration)

	duration, err = ParseDuration("0h")
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), time.Duration(0), duration)
}

func (suite *TimeTestSuite) TestParseDurationLargeValues() {
	// Test with large values that shouldn't overflow
	duration, err := ParseDuration("100yr")
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 100*365*24*time.Hour, duration)

	duration, err = ParseDuration("1000d")
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1000*24*time.Hour, duration)
}

func (suite *TimeTestSuite) TestParseDurationAllUnits() {
	testCases := []struct {
		input    string
		expected time.Duration
		desc     string
	}{
		{"1s", 1 * time.Second, "1 second"},
		{"60s", 60 * time.Second, "60 seconds"},
		{"1m", 1 * time.Minute, "1 minute"},
		{"60m", 60 * time.Minute, "60 minutes"},
		{"1h", 1 * time.Hour, "1 hour"},
		{"24h", 24 * time.Hour, "24 hours"},
		{"1d", 24 * time.Hour, "1 day"},
		{"7d", 7 * 24 * time.Hour, "7 days"},
		{"1w", 7 * 24 * time.Hour, "1 week"},
		{"4w", 4 * 7 * 24 * time.Hour, "4 weeks"},
		{"1mo", 30 * 24 * time.Hour, "1 month"},
		{"12mo", 12 * 30 * 24 * time.Hour, "12 months"},
		{"1yr", 365 * 24 * time.Hour, "1 year"},
		{"2yr", 2 * 365 * 24 * time.Hour, "2 years"},
	}

	for _, tc := range testCases {
		duration, err := ParseDuration(tc.input)
		require.NoError(suite.T(), err, "Failed for: %s", tc.desc)
		assert.Equal(suite.T(), tc.expected, duration, "Mismatch for: %s", tc.desc)
	}
}

func TestTimeSuite(t *testing.T) {
	suite.Run(t, new(TimeTestSuite))
}