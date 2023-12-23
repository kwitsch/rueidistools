package helper

import (
	"math"
	"time"
)

// MaxTTL is the maximum TTL value that can be stored in redis
// This is the maximum value of a signed 64 bit integer
const maxTTL int64 = 9223372036854775807

// DurationToTTL converts a time.Duration to TTL
// If the duration is negative, 0 is returned
// If the duration is greater than the maximum TTL, the maximum TTL is returned
// Otherwise the duration is rounded to seconds and returned
func DurationToTTL(d time.Duration) int64 {
	seconds := math.Round(d.Seconds())
	if seconds < 0 {
		return 0
	}

	if seconds > float64(maxTTL) {
		return maxTTL
	}

	return int64(seconds)
}

// TTLToDuration prints the TTL as a time.Duration string
func TTLToString(ttl int64) string {
	return time.Duration(ttl).String()
}
