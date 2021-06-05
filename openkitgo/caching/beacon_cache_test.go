package caching

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestAddEventData(t *testing.T) {

	c := NewBeaconCache(logger)
	k := NewBeaconKey(1, 1)
	c.AddEventData(k, time.Now(), "contents_1")

	assert.Equal(t, int64(20), c.cacheSizeInBytes)
	assert.Equal(t, 1, len(c.beacons))
}

func TestAddActionData(t *testing.T) {

	c := NewBeaconCache(logger)
	k := NewBeaconKey(1, 1)
	c.AddActionData(k, time.Now(), "contents_2")

	assert.Equal(t, int64(20), c.cacheSizeInBytes)
	assert.Equal(t, 1, len(c.beacons))
}

func TestDeleteCacheEntry(t *testing.T) {

	c := NewBeaconCache(logger)
	k := NewBeaconKey(1, 1)
	c.AddActionData(k, time.Now(), "contents_2")

	assert.Equal(t, int64(20), c.cacheSizeInBytes)
	assert.Equal(t, 1, len(c.beacons))

	c.DeleteCacheEntry(k)
	assert.Equal(t, int64(0), c.cacheSizeInBytes)
	assert.Equal(t, 0, len(c.beacons))

}
