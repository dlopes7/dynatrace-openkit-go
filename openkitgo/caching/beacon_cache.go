package caching

import (
	log "github.com/sirupsen/logrus"
	"sync"
	"sync/atomic"
	"time"
)

type BeaconCache struct {
	log              *log.Logger
	mutex            sync.RWMutex
	beacons          map[BeaconKey]*BeaconCacheEntry
	cacheSizeInBytes int64 // Atomic
	observers        []*chan bool
}

func NewBeaconCache(log *log.Logger) *BeaconCache {
	return &BeaconCache{
		log:     log,
		beacons: map[BeaconKey]*BeaconCacheEntry{},
	}
}

func (c *BeaconCache) AddEventData(key BeaconKey, timestamp time.Time, data string) {
	c.log.WithFields(log.Fields{"key": key.String(), "data": data, "time": timestamp}).Debug("BeaconCache.AddEventData()")

	entry := c.getCachedEntryOrInsert(key)
	record := NewBeaconCacheRecord(timestamp, data)

	entry.mutex.Lock()
	entry.addEventData(record)
	entry.mutex.Unlock()

	atomic.AddInt64(&c.cacheSizeInBytes, record.getDataSizeInBytes())

	c.onDataAdded()
}

func (c *BeaconCache) AddActionData(key BeaconKey, timestamp time.Time, data string) {
	c.log.WithFields(log.Fields{"key": key.String(), "data": data, "time": timestamp}).Debug("BeaconCache.AddActionData()")

	entry := c.getCachedEntryOrInsert(key)
	record := NewBeaconCacheRecord(timestamp, data)

	entry.mutex.Lock()
	entry.addActionData(record)
	entry.mutex.Unlock()

	atomic.AddInt64(&c.cacheSizeInBytes, record.getDataSizeInBytes())

	c.onDataAdded()

}

func (c *BeaconCache) getCachedEntryOrInsert(key BeaconKey) *BeaconCacheEntry {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	entry := c.getCachedEntry(key)

	if entry == nil {
		entry = &BeaconCacheEntry{}
		c.beacons[key] = entry
	} else {
		entry = c.beacons[key]
	}
	return entry
}

func (c *BeaconCache) getCachedEntry(key BeaconKey) *BeaconCacheEntry {
	return c.beacons[key]

}

func (c *BeaconCache) onDataAdded() {
	c.notifyObservers()
}

func (c *BeaconCache) notifyObservers() {
	for _, o := range c.observers {
		*o <- true
	}
}

func (c *BeaconCache) DeleteCacheEntry(key BeaconKey) {
	c.log.WithFields(log.Fields{"key": key.String()}).Debug("BeaconCache.DeleteCacheEntry()")

	var entry *BeaconCacheEntry

	c.mutex.Lock()
	entry = c.beacons[key]
	delete(c.beacons, key)
	c.mutex.Unlock()

	if entry != nil {
		atomic.AddInt64(&c.cacheSizeInBytes, -1*entry.totalNumBytes)
	}

}

func (c *BeaconCache) PrepareDataForSending(key BeaconKey) {
	c.mutex.Lock()
	entry := c.getCachedEntry(key)
	c.mutex.Unlock()
	if entry == nil {
		return
	}

	if entry.needsDataCopyBeforeSending() {
		entry.mutex.Lock()
		numBytes := entry.totalNumBytes
		entry.copyDataForSending()
		entry.mutex.Unlock()

		atomic.AddInt64(&c.cacheSizeInBytes, -1*numBytes)

	}
}

func (c *BeaconCache) HasDataForSending(key BeaconKey) bool {
	c.mutex.Lock()
	entry := c.getCachedEntry(key)
	c.mutex.Unlock()
	if entry == nil {
		return false
	}

	return entry.hasDataToSend()

}

func (c *BeaconCache) GetNextBeaconChunk(key BeaconKey, chunkPrefix string, maxSize int, delimiter rune) string {
	c.mutex.Lock()
	entry := c.getCachedEntry(key)
	c.mutex.Unlock()
	if entry == nil {
		return ""
	}
	return entry.getChunk(chunkPrefix, maxSize, delimiter)
}

func (c *BeaconCache) RemoveChunkedData(key BeaconKey) {
	c.mutex.Lock()
	entry := c.getCachedEntry(key)
	c.mutex.Unlock()
	if entry == nil {
		return
	}
	entry.mutex.Lock()
	entry.removeDataMarkedForSending()
	entry.mutex.Unlock()
}
func (c *BeaconCache) ResetChunkedData(key BeaconKey) {
	c.mutex.Lock()
	entry := c.getCachedEntry(key)
	c.mutex.Unlock()
	if entry == nil {
		return
	}

	entry.mutex.Lock()
	oldSize := entry.totalNumBytes
	entry.resetDataMarkedForSending()
	newSize := entry.totalNumBytes
	numBytes := newSize - oldSize
	entry.mutex.Unlock()

	atomic.AddInt64(&c.cacheSizeInBytes, numBytes)

	c.onDataAdded()

}

func (c *BeaconCache) GetBeaconKeys() []BeaconKey {
	var result []BeaconKey

	c.mutex.Lock()
	defer c.mutex.Unlock()
	for beaconKey := range c.beacons {
		result = append(result, beaconKey)
	}

	return result

}

func (c *BeaconCache) evictRecordsByAge(key BeaconKey, timestamp time.Time) int {
	c.mutex.Lock()
	entry := c.getCachedEntry(key)
	c.mutex.Unlock()
	if entry == nil {
		return 0
	}

	numRecordsRemoved := 0

	entry.mutex.Lock()
	numRecordsRemoved = entry.removeRecordsOlderThan(timestamp)
	entry.mutex.Unlock()

	log.WithFields(log.Fields{"key": key.String(), "timestamp": timestamp, "evicted": numRecordsRemoved}).Debug("BeaconCache.evictRecordsByAge()")

	return numRecordsRemoved
}

func (c *BeaconCache) evictRecordsByNumber(key BeaconKey, numRecords int) int {
	c.mutex.Lock()
	entry := c.getCachedEntry(key)
	c.mutex.Unlock()
	if entry == nil {
		return 0
	}

	numRecordsRemoved := 0

	entry.mutex.Lock()
	numRecordsRemoved = entry.removeOldestRecords(numRecords)
	entry.mutex.Unlock()

	log.WithFields(log.Fields{"key": key.String(), "numRecords": numRecords, "evicted": numRecordsRemoved}).Debug("BeaconCache.evictRecordsByNumber()")

	return numRecordsRemoved
}

func (c *BeaconCache) getNumBytesInCache() int64 {
	return atomic.LoadInt64(&c.cacheSizeInBytes)
}

func (c *BeaconCache) IsEmpty(key BeaconKey) bool {
	c.mutex.Lock()
	entry := c.getCachedEntry(key)
	c.mutex.Unlock()
	if entry == nil {
		return true
	}

	entry.mutex.Lock()
	defer entry.mutex.Unlock()
	return entry.totalNumBytes == 0

}

func (c *BeaconCache) AddObservable(channel *chan bool) {
	c.observers = append(c.observers, channel)
}
