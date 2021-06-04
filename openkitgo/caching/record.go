package caching

import "time"

const (
	CHAR_SIZE_BYTES = 2
)

type BeaconCacheRecord struct {
	timestamp        time.Time
	data             string
	markedForSending bool
}

func NewBeaconCacheRecord(timestamp time.Time, data string) *BeaconCacheRecord {
	return &BeaconCacheRecord{
		timestamp: timestamp,
		data:      data,
	}
}

func (r *BeaconCacheRecord) getDataSizeInBytes() int64 {
	return int64(len(r.data)) * CHAR_SIZE_BYTES
}

func (r *BeaconCacheRecord) GetTimestamp() time.Time {
	return r.timestamp
}
