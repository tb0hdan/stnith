package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

func ParseDuration(s string) (time.Duration, error) {
	re := regexp.MustCompile(`^(\d+)(yr|mo|[wdhms])$`)
	matches := re.FindStringSubmatch(s)
	if matches == nil {
		return 0, fmt.Errorf("invalid duration format: %s (use format like 1yr, 1mo, 1w, 2d, 3h, 30m, 45s)", s)
	}

	value, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, err
	}

	unit := matches[2]
	var duration time.Duration

	switch unit {
	case "yr":
		duration = time.Duration(value) * 365 * 24 * time.Hour
	case "mo":
		duration = time.Duration(value) * 30 * 24 * time.Hour
	case "w":
		duration = time.Duration(value) * 7 * 24 * time.Hour
	case "d":
		duration = time.Duration(value) * 24 * time.Hour
	case "h":
		duration = time.Duration(value) * time.Hour
	case "m":
		duration = time.Duration(value) * time.Minute
	case "s":
		duration = time.Duration(value) * time.Second
	default:
		return 0, fmt.Errorf("unknown time unit: %s", unit)
	}

	return duration, nil
}
