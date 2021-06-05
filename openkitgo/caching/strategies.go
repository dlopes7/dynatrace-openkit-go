package caching

import (
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo/configuration"
	log "github.com/sirupsen/logrus"
	"time"
)

type SpaceEvictionStrategy struct {
	log           *log.Logger
	cache         *BeaconCache
	configuration *configuration.BeaconCacheConfiguration
}

func NewSpaceEvictionStrategy(log *log.Logger, cache *BeaconCache, configuration *configuration.BeaconCacheConfiguration) *SpaceEvictionStrategy {
	return &SpaceEvictionStrategy{log: log, cache: cache, configuration: configuration}
}

func (s *SpaceEvictionStrategy) execute() {
	if s.cache.getNumBytesInCache() > s.configuration.CacheSizeUpperBound {
		numRecordsRemoved := 0
		for _, key := range s.cache.GetBeaconKeys() {
			numRecordsRemoved += s.cache.evictRecordsByNumber(key, 1)
		}
		s.log.WithFields(log.Fields{"numRecordsRemoved": numRecordsRemoved}).Debug("SpaceEvictionStrategy removed records")
	}
}

type TimeEvictionStrategy struct {
	log              *log.Logger
	cache            *BeaconCache
	configuration    *configuration.BeaconCacheConfiguration
	lastRunTimestamp time.Time
}

func NewTimeEvictionStrategy(log *log.Logger, cache *BeaconCache, configuration *configuration.BeaconCacheConfiguration) *TimeEvictionStrategy {
	return &TimeEvictionStrategy{log: log, cache: cache, configuration: configuration}
}

func (s *TimeEvictionStrategy) execute() {

	if s.lastRunTimestamp.IsZero() {
		s.lastRunTimestamp = time.Now()
	}

	if time.Now().Sub(s.lastRunTimestamp) > s.configuration.MaxRecordAge {
		minAllowedAge := s.lastRunTimestamp.Add(-1 * s.configuration.MaxRecordAge)
		numRecordsRemoved := 0
		for _, key := range s.cache.GetBeaconKeys() {
			numRecordsRemoved += s.cache.evictRecordsByAge(key, minAllowedAge)
		}
		s.log.WithFields(log.Fields{"numRecordsRemoved": numRecordsRemoved}).Debug("TimeEvictionStrategy removed records")
	}
	s.lastRunTimestamp = time.Now()
}
