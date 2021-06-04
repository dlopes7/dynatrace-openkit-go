package caching

import (
	"sort"
	"strings"
	"sync"
	"time"
)

type BeaconCacheEntry struct {
	eventData  []*BeaconCacheRecord
	actionData []*BeaconCacheRecord
	mutex      sync.Mutex

	eventDataBeingSent  []*BeaconCacheRecord
	actionDataBeingSent []*BeaconCacheRecord

	totalNumBytes int64
}

func (e *BeaconCacheEntry) addEventData(record *BeaconCacheRecord) {
	e.eventData = append(e.eventData, record)
	e.totalNumBytes += record.getDataSizeInBytes()
}

func (e *BeaconCacheEntry) addActionData(record *BeaconCacheRecord) {
	e.actionData = append(e.actionData, record)
	e.totalNumBytes += record.getDataSizeInBytes()
}

func (e *BeaconCacheEntry) needsDataCopyBeforeSending() bool {
	return !e.hasDataToSend()
}

func (e *BeaconCacheEntry) hasDataToSend() bool {
	return len(e.eventDataBeingSent) > 0 || len(e.actionDataBeingSent) > 0
}

func (e *BeaconCacheEntry) copyDataForSending() {
	e.actionDataBeingSent = e.actionData
	e.eventDataBeingSent = e.eventData
	e.actionData = []*BeaconCacheRecord{}
	e.eventData = []*BeaconCacheRecord{}
	e.totalNumBytes = 0
}

func (e *BeaconCacheEntry) getChunk(chunkPrefix string, maxSize int, delimiter rune) string {
	if !e.hasDataToSend() {
		return ""
	}

	return e.getNextChunk(chunkPrefix, maxSize, delimiter)
}

func (e *BeaconCacheEntry) getNextChunk(chunkPrefix string, maxSize int, delimiter rune) string {

	var b strings.Builder
	b.WriteString(chunkPrefix)

	e.chunkifyDataList(&b, e.eventDataBeingSent, maxSize, delimiter)
	e.chunkifyDataList(&b, e.actionDataBeingSent, maxSize, delimiter)

	return b.String()

}

func (e *BeaconCacheEntry) chunkifyDataList(builder *strings.Builder, dataBeingSent []*BeaconCacheRecord, maxSize int, delimiter rune) {

	for _, record := range dataBeingSent {

		if builder.Len() <= maxSize {
			record.markedForSending = true
			builder.WriteRune(delimiter)
			builder.WriteString(record.data)
		}
	}
}

func (e *BeaconCacheEntry) removeDataMarkedForSending() {

	if !e.hasDataToSend() {
		return
	}

	var keepEvents []*BeaconCacheRecord
	for _, eventRecord := range e.eventDataBeingSent {
		if !eventRecord.markedForSending {
			keepEvents = append(keepEvents, eventRecord)
		}
	}
	e.eventDataBeingSent = keepEvents

	var keepActions []*BeaconCacheRecord
	for _, eventRecord := range e.actionDataBeingSent {
		if !eventRecord.markedForSending {
			keepActions = append(keepEvents, eventRecord)
		}
	}
	e.actionDataBeingSent = keepActions

}

func (e *BeaconCacheEntry) resetDataMarkedForSending() {

	if !e.hasDataToSend() {
		return
	}

	numBytes := int64(0)

	for _, record := range e.eventDataBeingSent {
		record.markedForSending = false
		numBytes += record.getDataSizeInBytes()
	}

	for _, record := range e.actionDataBeingSent {
		record.markedForSending = false
		numBytes += record.getDataSizeInBytes()
	}

	e.eventDataBeingSent = append(e.eventDataBeingSent, e.eventData...)
	e.actionDataBeingSent = append(e.actionDataBeingSent, e.actionData...)

	e.eventData = e.eventDataBeingSent
	e.actionData = e.actionDataBeingSent

	e.totalNumBytes += numBytes

}

func (e *BeaconCacheEntry) removeRecordsOlderThan(timestamp time.Time) int {

	numRecordsRemoved := 0

	var keepEvents []*BeaconCacheRecord
	for _, eventRecord := range e.eventDataBeingSent {
		if eventRecord.timestamp.After(timestamp) {
			keepEvents = append(keepEvents, eventRecord)
		} else {
			numRecordsRemoved += 1
		}
	}
	e.eventDataBeingSent = keepEvents

	var keepActions []*BeaconCacheRecord
	for _, actionRecord := range e.actionDataBeingSent {
		if actionRecord.timestamp.After(timestamp) {
			keepActions = append(keepActions, actionRecord)
		} else {
			numRecordsRemoved += 1
		}
	}
	e.actionDataBeingSent = keepActions

	return numRecordsRemoved

}

func (e *BeaconCacheEntry) removeOldestRecords(numRecords int) int {

	numRecordsRemoved := 0

	// First, sort our slices, oldest events and actions first
	sort.Slice(e.eventData, func(i, j int) bool {
		return e.eventData[i].timestamp.Before(e.eventData[j].timestamp)
	})

	sort.Slice(e.actionData, func(i, j int) bool {
		return e.actionData[i].timestamp.Before(e.actionData[j].timestamp)
	})

	for numRecordsRemoved < numRecords {

		// if we have actions and events, remove the oldest one
		if len(e.actionData) > 0 && len(e.eventData) > 0 {
			if e.eventData[0].timestamp.Before(e.actionData[0].timestamp) {
				e.eventData = e.eventData[1:]
			} else {
				e.actionData = e.actionData[1:]
			}
		} else if len(e.actionData) > 0 {
			// We only have actions, remove one
			e.actionData = e.actionData[1:]
		} else if len(e.eventData) > 0 {
			// We only have events, remove one
			e.eventData = e.eventData[1:]
		}
		// This always increases, even if both are empty so we are guaranteed to leave
		numRecordsRemoved += 1

	}

	return numRecordsRemoved

}
