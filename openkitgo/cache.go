package openkitgo

import (
	"github.com/op/go-logging"
	"sync"
)

type BeaconCache interface {
	addObserver(Observer)
	addEventData(int, int, string)
	addActionData(int, int, string)

	deleteCacheEntry(int)
	getNextBeaconChunk(int, string, int, rune)
	removeChunkedData(int)
	resetChunkedData(int)

	getCachedEntryOrInsert(int)

	getEvents(int) []string
	getEventsBeingSent(int) []*beaconCacheRecord

	getActions(int) []string
	getActionsBeingSent(int) []*beaconCacheRecord

	extractData([]*beaconCacheRecord) []string

	getCachedEntry(int) BeaconCacheEntry

	getBeaconIDs() map[int]bool

	evictRecordsByAge(int, int) int
	evictRecordsByNumber(int, int) int
	getNumBytesInCache() int

	isEmpty(int) bool
}

type beaconCache struct {
	logger          *logging.Logger
	globalCacheLock sync.Mutex

	cacheSizeInBytes int
	beacons          map[int]*beaconCacheEntry

	observers []Observer
	changed   bool
}

func NewBeaconCache(logger *logging.Logger) *beaconCache {
	b := new(beaconCache)

	b.logger = logger
	b.beacons = make(map[int]*beaconCacheEntry)

	return b
}

func (b *beaconCache) addObserver(observer Observer) {
	b.observers = append(b.observers, observer)

}

func (b *beaconCache) addEventData(beaconID int, timestamp int, data string) {
	b.logger.Debugf("addEventData(sn=%d, timestamp=%d, data=%s)", beaconID, timestamp, data)

	entry := b.getCachedEntryOrInsert(beaconID)
	record := &beaconCacheRecord{
		timestamp: timestamp,
		data:      data,
	}

	entry.lock.Lock()
	entry.addEventData(record)
	entry.lock.Unlock()

	b.cacheSizeInBytes += record.getDataSizeInBytes()

	b.onDataAdded()

}

func (b *beaconCache) getCachedEntryOrInsert(beaconID int) *beaconCacheEntry {

	entry := b.getCachedEntry(beaconID)

	if entry == nil {
		b.globalCacheLock.Lock()

		// If there is a cache entry for this beacon already
		if val, ok := b.beacons[beaconID]; ok {
			entry = val

		} else {
			entry = new(beaconCacheEntry)
			b.beacons[beaconID] = entry
		}

		b.globalCacheLock.Unlock()
	}

	return entry

}

func (b *beaconCache) getCachedEntry(beaconID int) *beaconCacheEntry {
	b.globalCacheLock.Lock()
	entry := b.beacons[beaconID]
	b.globalCacheLock.Unlock()

	return entry
}

func (b *beaconCache) resetChunkedData(beaconID int) {

	entry := b.getCachedEntry(beaconID)
	if entry == nil {
		return
	}

	numBytes := 0

	entry.lock.Lock()

	oldSize := entry.totalNumBytes
	entry.resetDataMarkedForSending()
	newSize := entry.totalNumBytes
	numBytes = newSize - oldSize

	entry.lock.Unlock()

	b.cacheSizeInBytes += numBytes

	b.onDataAdded()

}

func (b *beaconCache) onDataAdded() {
	b.setChanged()
	b.notifyObservers()

}

func (b *beaconCache) notifyObservers() {

}

func (b *beaconCache) setChanged() {
	b.changed = true

}

func (b *beaconCache) addActionData(beaconID int, timestamp int, data string) {
	b.logger.Debugf("addActionData(sn=%d, timestamp=%d, data=%s)\n")

	entry := b.getCachedEntryOrInsert(beaconID)
	record := &beaconCacheRecord{
		timestamp: timestamp,
		data:      data,
	}

	entry.lock.Lock()
	entry.addActionData(record)
	entry.lock.Unlock()

	b.cacheSizeInBytes += record.getDataSizeInBytes()

	b.onDataAdded()

}

type BeaconCacheEntry interface {
}

type beaconCacheEntry struct {
	lock sync.Mutex

	totalNumBytes int

	eventData  []*beaconCacheRecord
	actionData []*beaconCacheRecord

	eventDataBeingSent  []*beaconCacheRecord
	actionDataBeingSent []*beaconCacheRecord
}

func (b *beaconCacheEntry) addEventData(record *beaconCacheRecord) {
	b.eventData = append(b.eventData, record)
	b.totalNumBytes += record.getDataSizeInBytes()

}

func (b *beaconCacheEntry) addActionData(record *beaconCacheRecord) {
	b.actionData = append(b.actionData, record)
	b.totalNumBytes += record.getDataSizeInBytes()

}

func (b *beaconCacheEntry) resetDataMarkedForSending() {

	numBytes := 0
	for _, record := range b.eventDataBeingSent {
		record.markedForSending = false
		numBytes += record.getDataSizeInBytes()
	}

	for _, record := range b.actionDataBeingSent {
		record.markedForSending = false
		numBytes += record.getDataSizeInBytes()
	}

	b.eventDataBeingSent = append(b.eventDataBeingSent, b.eventData...)
	b.actionDataBeingSent = append(b.actionDataBeingSent, b.actionData...)

	b.eventData = b.eventDataBeingSent
	b.actionData = b.actionDataBeingSent

	b.eventDataBeingSent = nil
	b.actionDataBeingSent = nil

	b.totalNumBytes += numBytes

}

type beaconCacheRecord struct {
	timestamp        int
	data             string
	markedForSending bool
}

func (b *beaconCacheRecord) getDataSizeInBytes() int {
	return len(b.data) * 2
}

type Observable interface {
	setChanged()
}

type Observer interface {
}
