package core

import (
	"fmt"
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo"
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo/caching"
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo/configuration"
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo/utils"
	log "github.com/sirupsen/logrus"
	"net/http"
	"sync"
	"time"
)

type OpenKit struct {
	log                  *log.Logger
	privacyConfiguration *configuration.PrivacyConfiguration
	openKitConfiguration *configuration.OpenKitConfiguration
	beaconCache          *caching.BeaconCache
	beaconCacheEvictor   *caching.BeaconCacheEvictor
	beaconSender         *BeaconSender
	isShutDown           bool
	mutex                sync.Mutex
	// TODO sessionWatchdog

	children []OpenKitObject
}

func (o *OpenKit) CreateSessionWithDeviceID(clientIPAddress string, deviceID int64) openkitgo.Session {
	return o.CreateSessionAtWithDeviceID(clientIPAddress, time.Now(), deviceID)

}

func (o *OpenKit) CreateSessionAtWithDeviceID(clientIPAddress string, timestamp time.Time, deviceID int64) openkitgo.Session {
	o.log.WithFields(log.Fields{"clientIPAddress": clientIPAddress, "timestamp": timestamp}).Debug("OpenKit.CreateSessionAt()")

	o.mutex.Lock()
	defer o.mutex.Unlock()
	if !o.isShutDown {

		o.openKitConfiguration.DeviceID = deviceID
		sessionProxy := NewSessionProxy(
			o.log,
			o,
			o.beaconSender,
			o,
			clientIPAddress,
			// TODO o.sessionWatchdog
			timestamp,
		)

		o.storeChildInList(sessionProxy)
		return sessionProxy
	}

	return NewNullSession()
}

func (o *OpenKit) CreateSession(clientIPAddress string) openkitgo.Session {
	return o.CreateSessionAtWithDeviceID(clientIPAddress, time.Now(), o.openKitConfiguration.DeviceID)
}

func (o *OpenKit) CreateSessionAt(clientIPAddress string, timestamp time.Time) openkitgo.Session {
	return o.CreateSessionAtWithDeviceID(clientIPAddress, timestamp, o.openKitConfiguration.DeviceID)
}

func NewOpenKit(builder *OpenKitBuilder) openkitgo.OpenKit {

	privacyConfig := &configuration.PrivacyConfiguration{
		DataCollectionLevel: builder.dataCollectionLevel,
		CrashReportingLevel: builder.crashReportLevel,
	}

	openKitConfig := &configuration.OpenKitConfiguration{
		EndpointURL:                 builder.endpointURL,
		DeviceID:                    builder.deviceID,
		OrigDeviceID:                builder.origDeviceID,
		OpenKitType:                 OPENKIT_TYPE,
		ApplicationID:               builder.applicationID,
		PercentEncodedApplicationID: utils.PercentEncode(builder.applicationID),
		ApplicationName:             builder.applicationName,
		ApplicationVersion:          builder.applicationVersion,
		OperatingSystem:             builder.operatingSystem,
		Manufacturer:                builder.manufacturer,
		ModelID:                     builder.modelID,
		DefaultServerID:             DEFAULT_SERVER_ID,
		Transport:                   http.Transport{},
	}

	beaconCache := caching.NewBeaconCache(builder.log)
	beaconCacheConfig := configuration.NewBeaconCacheConfiguration(
		builder.beaconCacheMaxRecordAge,
		builder.beaconCacheLowerMemoryBoundary,
		builder.beaconCacheUpperMemoryBoundary)
	beaconCacheEvictor := caching.NewBeaconCacheEvictor(builder.log, beaconCache, beaconCacheConfig)

	httpClientConfig := &configuration.HttpClientConfiguration{
		BaseURL:       builder.endpointURL,
		ServerID:      DEFAULT_SERVER_ID,
		ApplicationID: builder.applicationID,
		Transport:     builder.transport,
	}

	beaconSender := NewBeaconSender(builder.log, httpClientConfig)
	// TODO sessionWatchdog

	ok := &OpenKit{
		log:                  builder.log,
		privacyConfiguration: privacyConfig,
		openKitConfiguration: openKitConfig,
		beaconCache:          beaconCache,
		beaconCacheEvictor:   beaconCacheEvictor,
		beaconSender:         beaconSender,
		// TODO sessionWatchdog
	}

	return ok
}

func (o *OpenKit) initialize() {

	o.beaconCacheEvictor.Start()
	// TODO sessionWatchdog.Start()
	o.beaconSender.Initialize()

}

func (o *OpenKit) String() string {
	return fmt.Sprintf("OpenKit(%s, %s)", o.openKitConfiguration.OpenKitType, DEFAULT_APPLICATION_VERSION)
}
func (o *OpenKit) DetailedString() string {
	return fmt.Sprintf("OpenKit(Type=%s, Version=%s, ApplicationName=%s, ApplicationID=%s, DeviceID=%d, OrigDeviceID=%s, EndpointURL=%s)",
		o.openKitConfiguration.OpenKitType,
		DEFAULT_APPLICATION_VERSION,
		o.openKitConfiguration.ApplicationName,
		o.openKitConfiguration.ApplicationID,
		o.openKitConfiguration.DeviceID,
		o.openKitConfiguration.OrigDeviceID,
		o.openKitConfiguration.EndpointURL,
	)
}

func (o *OpenKit) close() {
	o.closeAt(time.Now())
}

func (o *OpenKit) closeAt(timestamp time.Time) {
	o.Shutdown()
}

func (o *OpenKit) WaitForInitCompletion() bool {
	return o.beaconSender.WaitForInit()
}

func (o *OpenKit) WaitForInitCompletionTimeout(duration time.Duration) bool {
	return o.beaconSender.WaitForInitTimeout(duration)
}

func (o *OpenKit) Shutdown() {
	o.log.Debug("OpenKit.shutdown()")
	o.mutex.Lock()
	if o.isShutDown {
		return
	}
	o.isShutDown = true

	for _, child := range o.getCopyOfChildObjects() {
		child.close()
	}

	o.beaconCacheEvictor.Stop()
	// TODO  o.sessionWatchdog.Shutdown();
	o.beaconSender.Shutdown()

}

func (o *OpenKit) getCopyOfChildObjects() []OpenKitObject {
	return o.children[:]
}

func (o *OpenKit) onChildClosed(child OpenKitObject) {
	o.removeChildFromList(child)
}

func (o *OpenKit) storeChildInList(child OpenKitObject) {
	o.children = append(o.children, child)

}

func (o *OpenKit) removeChildFromList(child OpenKitObject) bool {
	removed := false

	var keep []OpenKitObject
	for _, c := range o.children {
		if c != child {
			keep = append(keep, c)
		} else {
			removed = true
		}
	}
	o.children = keep
	return removed
}

func (o *OpenKit) getChildCount() int {
	return len(o.children)
}

func (o *OpenKit) getActionID() int {
	return DEFAULT_ACTION_ID
}
