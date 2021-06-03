package utils

import (
	"time"
)

func TimeToMillis(timestamp time.Time) int64 {
	return timestamp.UnixNano() / int64(time.Millisecond)
}

func DurationToMillis(duration time.Duration) int64 {
	return duration.Nanoseconds() / int64(time.Millisecond)
}
