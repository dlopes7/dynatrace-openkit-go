package configuration

import (
	"time"
)

type BeaconCacheConfiguration struct {
	MaxRecordAge        time.Duration
	CacheSizeLowerBound int64
	CacheSizeUpperBound int64
}

func NewBeaconCacheConfiguration(maxRecordAge time.Duration, cacheSizeLowerBound int64, cacheSizeUpperBound int64) *BeaconCacheConfiguration {
	return &BeaconCacheConfiguration{
		MaxRecordAge:        maxRecordAge,
		CacheSizeLowerBound: cacheSizeLowerBound,
		CacheSizeUpperBound: cacheSizeUpperBound,
	}
}
