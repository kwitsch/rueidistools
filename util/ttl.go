package util

import (
	"math"
	"time"
)

// MaxTTL is the maximum TTL value that can be stored in redis
// This is the maximum value of a signed 64 bit integer
const maxTTL int64 = 9223372036854775807

// DurationToTTL converts a time.Duration to TTL(int64).
// If the input is negative, 0 is returned.
// If the input is greater than the maximum TTL, the maximum TTL is returned.
// Otherwise the input is rounded to seconds and returned.
func DurationToTTL(d time.Duration) int64 {
	return Float64ToTTL(d.Seconds())
}

// Float64ToTTL converts a float64 with seconds to TTL(int64)
// If the input is negative, 0 is returned.
// If the input is greater than the maximum TTL, the maximum TTL is returned.
// Otherwise the input is rounded to seconds and returned.
func Float64ToTTL(f float64) int64 {
	seconds := math.Round(f)
	if seconds < 0 {
		return 0
	}

	if seconds > float64(maxTTL) {
		return maxTTL
	}

	return int64(seconds)
}
