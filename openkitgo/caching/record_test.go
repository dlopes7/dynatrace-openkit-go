package caching

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewBeaconCacheRecord(t *testing.T) {
	timestamp := time.Unix(1622830000, 0)
	contents := "contents"
	r := NewBeaconCacheRecord(timestamp, contents)

	assert.Equal(t, r.getDataSizeInBytes(), int64(16))
	assert.Equal(t, r.GetTimestamp(), timestamp)

}
