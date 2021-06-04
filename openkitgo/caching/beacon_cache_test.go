package caching

import (
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

var logger *log.Logger

func TestMain(m *testing.M) {
	logger = log.New()
	logger.SetLevel(log.DebugLevel)
	os.Exit(m.Run())
}

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
