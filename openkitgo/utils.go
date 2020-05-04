package openkitgo

import (
	"time"
)

func TimeToMillis(timestamp time.Time) int {
	return int(timestamp.UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond)))
}

func DurationToMillis(duration time.Duration) int {
	return int(duration.Nanoseconds() / 1000000)
}
