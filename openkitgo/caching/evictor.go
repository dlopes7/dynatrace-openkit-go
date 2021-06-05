package caching

import (
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo/configuration"
	log "github.com/sirupsen/logrus"
	"time"
)

const (
	THREAD_NAME                  = "BeaconCacheEvictor"
	EVICTION_THREAD_JOIN_TIMEOUT = 2 * time.Second
)

type BeaconCacheEvictionStrategy interface {
	execute()
}

type BeaconCacheEvictor struct {
	log     *log.Logger
	channel *chan bool
	alive   bool
	cache   *BeaconCache
	config  *configuration.BeaconCacheConfiguration
}

func EvictionRoutine(log *log.Logger, cache *BeaconCache, strategies ...BeaconCacheEvictionStrategy) *chan bool {

	recordAdded := make(chan bool)
	stop := make(chan bool)

	cache.AddObservable(&recordAdded)

	go func() {
		log.Debug("EvictionRoutine.run()")
		for {

			select {
			case <-recordAdded:
				for _, strategy := range strategies {
					strategy.execute()
				}
			case <-stop:
				log.Debug("EvictionRoutine.stop()")
				return
			}
		}

	}()

	return &stop

}

func NewBeaconCacheEvictor(
	log *log.Logger,
	cache *BeaconCache,
	configuration *configuration.BeaconCacheConfiguration,
) *BeaconCacheEvictor {

	return &BeaconCacheEvictor{
		log:    log,
		cache:  cache,
		config: configuration,
	}
}

func (e *BeaconCacheEvictor) Stop() {
	*e.channel <- true
}

func (e *BeaconCacheEvictor) Start() {

	if !e.alive {
		spaceEvictionStrategy := NewSpaceEvictionStrategy(e.log, e.cache, e.config)
		timeEvictionStrategy := NewTimeEvictionStrategy(e.log, e.cache, e.config)
		e.channel = EvictionRoutine(e.log, e.cache, spaceEvictionStrategy, timeEvictionStrategy)
		e.alive = true
	} else {
		log.Debug("Not starting the evictor because it is already running")
	}
}
