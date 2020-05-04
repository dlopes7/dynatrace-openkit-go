package openkitgo

import (
	log "github.com/sirupsen/logrus"
	"strings"
	"sync"
	"time"
)

type BeaconCache interface {
	addObserver(Observer)
	addEventData(int, int, string)
	addActionData(int, int, string)

	deleteCacheEntry(int)
	getNextBeaconChunk(int, string, int, string)
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
	log             *log.Logger
	globalCacheLock sync.Mutex

	cacheSizeInBytes int
	beacons          map[int]*beaconCacheEntry

	observers []Observer
	changed   bool
}

func NewBeaconCache(log *log.Logger) *beaconCache {
	b := new(beaconCache)

	b.log = log
	b.beacons = make(map[int]*beaconCacheEntry)

	return b
}

func (b *beaconCache) addObserver(observer Observer) {
	b.observers = append(b.observers, observer)

}

func (b *beaconCache) addEventData(beaconID int, timestamp time.Time, data string) {
	b.log.Debugf("addEventData(sn: %d, timestamp: %d, data: %s)", beaconID, timestamp, data)

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

func (b *beaconCache) deleteCacheEntry(beaconID int) {
	b.log.Debugf("deleteCacheEntry(sn=%d)", beaconID)

	var entry *beaconCacheEntry

	b.globalCacheLock.Lock()
	entry = b.beacons[beaconID]

	delete(b.beacons, beaconID)
	b.globalCacheLock.Unlock()

	if entry != nil {
		b.cacheSizeInBytes += -1 * entry.totalNumBytes
	}

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

func (b *beaconCache) getNextBeaconChunk(beaconID int, chunkPrefix string, maxSize int, delimiter string) *string {
	entry := b.getCachedEntry(beaconID)

	if entry == nil {
		// a cache entry for the given beaconID does not exist
		return nil
	}

	if entry.needsDataCopyBeforeChunking() {
		var numBytes int

		entry.lock.Lock()
		numBytes = entry.totalNumBytes
		entry.copyDataForChunking()
		entry.lock.Unlock()

		b.cacheSizeInBytes += -1 * numBytes
	}

	return entry.getChunk(chunkPrefix, maxSize, delimiter)

}

func (b *beaconCache) addActionData(beaconID int, timestamp time.Time, data string) {
	b.log.Debugf("addActionData(sn=%d, timestamp=%d, data=%s)", beaconID, timestamp, data)

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

func (b *beaconCache) removeChunkedData(beaconID int) {
	entry := b.getCachedEntry(beaconID)
	if entry == nil {
		// a cache entry for the given beaconID does not exist
		return
	}

	entry.removeDataMarkedForSending()

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

func (b *beaconCacheEntry) needsDataCopyBeforeChunking() bool {
	return b.actionDataBeingSent == nil && b.eventDataBeingSent == nil
}

func (b *beaconCacheEntry) copyDataForChunking() {
	b.actionDataBeingSent = b.actionData
	b.eventDataBeingSent = b.eventData

	b.actionData = make([]*beaconCacheRecord, 0)
	b.eventData = make([]*beaconCacheRecord, 0)

	b.totalNumBytes = 0

}

func (b *beaconCacheEntry) getChunk(chunkPrefix string, maxSize int, delimiter string) *string {
	if !b.hasDataToSend() {

		b.eventDataBeingSent = nil
		b.actionDataBeingSent = nil

		ret := ""
		return &ret
	}
	return b.getNextChunk(chunkPrefix, maxSize, delimiter)
}

func (b *beaconCacheEntry) getNextChunk(chunkPrefix string, maxSize int, delimiter string) *string {
	var beaconBuilder strings.Builder

	beaconBuilder.WriteString(chunkPrefix)

	b.chunkifyDataList(&beaconBuilder, b.eventDataBeingSent, maxSize, delimiter)
	b.chunkifyDataList(&beaconBuilder, b.actionDataBeingSent, maxSize, delimiter)

	res := beaconBuilder.String()
	return &res
}

func (b *beaconCacheEntry) chunkifyDataList(chunkBuilder *strings.Builder, dataBeingSent []*beaconCacheRecord, maxSize int, delimiter string) {

	for _, dataBeingSent := range dataBeingSent {

		if chunkBuilder.Len() <= maxSize {

			dataBeingSent.markedForSending = true
			chunkBuilder.WriteString(delimiter)
			chunkBuilder.WriteString(dataBeingSent.data)
		}
	}
}

func (b *beaconCacheEntry) hasDataToSend() bool {
	return (b.eventDataBeingSent != nil && len(b.eventDataBeingSent) > 0) || (b.actionDataBeingSent != nil && len(b.actionDataBeingSent) > 0)
}

func (b *beaconCacheEntry) removeDataMarkedForSending() {

	if !b.hasDataToSend() {
		// data has not been copied yet - avoid NPE
		return
	}

	i := 0
	for _, e := range b.eventDataBeingSent {
		if !e.markedForSending {
			b.eventDataBeingSent[i] = e
			i++
		}
	}
	b.eventDataBeingSent = b.eventDataBeingSent[:i]

	i = 0
	for _, e := range b.actionDataBeingSent {
		if !e.markedForSending {
			b.actionDataBeingSent[i] = e
			i++
		}
	}
	b.actionDataBeingSent = b.actionDataBeingSent[:i]
}

type beaconCacheRecord struct {
	timestamp        time.Time
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
