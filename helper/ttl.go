package helper

import (
	"math"
	"time"
)

// MaxTTL is the maximum TTL value that can be stored in redis
// This is the maximum value of a signed 32 bit integer
const maxTTL uint32 = 2147483647

// TTL is used for conversion of redis key experations.
type TTL interface {
	int32 | uint32 | int64 | uint64
}

// DurationToTTL converts a time.Duration to TTL
// If the duration is negative, 0 is returned
// If the duration is greater than the maximum TTL, the maximum TTL is returned
// Otherwise the duration is rounded to seconds and returned
func DurationToTTL[T TTL](d time.Duration) T {
	seconds := math.Round(d.Seconds())
	if seconds < 0 {
		return T(0)
	}

	if seconds > float64(maxTTL) {
		return T(maxTTL)
	}

	return T(seconds)
}
