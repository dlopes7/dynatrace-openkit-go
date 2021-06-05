package utils

import (
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo/configuration"
	"net/url"
	"strings"
	"time"
)

func TimeToMillis(timestamp time.Time) int64 {
	return timestamp.UnixNano() / int64(time.Millisecond)
}

func DurationToMillis(duration time.Duration) int64 {
	return duration.Nanoseconds() / int64(time.Millisecond)
}

func PercentEncode(text string) string {
	return url.QueryEscape(
		strings.ReplaceAll(text, configuration.RESERVED_CHARACTERS, ""))
}
