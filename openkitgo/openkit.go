package openkitgo

import (
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

type openKitType int

const (
	OpenKitTypeAPPMON    openKitType = 1
	OpenKitTypeDYNATRACE openKitType = 1
)

type OpenKit interface {
	CreateSession(string) Session
	CreateSessionWithTime(string, time.Time) Session
	CreateSessionWithTimeAndDevice(string, time.Time, string) Session
	// waitForInitCompletion(int) bool
	// isInitialized() bool
}

type openkit struct {
	beaconCache   *beaconCache
	beaconSender  *BeaconSender
	configuration *Configuration
	log           *log.Logger

	lock sync.Mutex
}

func (o *openkit) CreateSession(clientIPAddress string) Session {

	o.log.Debugf("Creating session with IP address %s", clientIPAddress)

	o.lock.Lock()
	defer o.lock.Unlock()

	beacon := NewBeacon(o.log, o.beaconCache, o.configuration, clientIPAddress)

	return newSession(o.log, o.beaconSender, beacon)
}

func (o *openkit) CreateSessionWithTime(clientIPAddress string, timestamp time.Time) Session {

	o.log.Debugf("Creating session with IP address %s", clientIPAddress)

	beacon := NewBeaconWithTime(o.log, o.beaconCache, o.configuration, clientIPAddress, timestamp)

	return newSession(o.log, o.beaconSender, beacon)
}

func (o *openkit) CreateSessionWithTimeAndDevice(clientIPAddress string, timestamp time.Time, deviceID string) Session {

	o.log.Debugf("Creating session with IP address %s", clientIPAddress)

	beacon := NewBeaconWithTimeAndDevice(o.log, o.beaconCache, o.configuration, clientIPAddress, timestamp, deviceID)

	return newSession(o.log, o.beaconSender, beacon)
}

func (o *openkit) initialize() {
	o.beaconSender.initialize()
}

type OpenKitBuilder interface {
	WithLogLevel(log.Level) OpenKitBuilder
	WithLogger(*log.Logger) OpenKitBuilder
	WithApplicationName(string) OpenKitBuilder
	WithApplicationVersion(string) OpenKitBuilder
	WithOperatingSystem(string) OpenKitBuilder
	WithManufacturer(string) OpenKitBuilder
	WithModelID(string) OpenKitBuilder
	Build() OpenKit
}

type openKitBuilder struct {
	logLevel log.Level
	log      *log.Logger

	endpointURL   string
	applicationID string
	deviceID      int

	applicationName    string
	applicationVersion string
	operatingSystem    string
	manufacturer       string
	modelID            string
}

func NewOpenKitBuilder(endpointURL string, applicationID string, deviceID int) OpenKitBuilder {

	return &openKitBuilder{
		endpointURL:   endpointURL,
		applicationID: applicationID,
		deviceID:      deviceID,
	}
}

func (ob *openKitBuilder) WithLogLevel(logLevel log.Level) OpenKitBuilder {
	ob.logLevel = logLevel
	return ob
}

func (ob *openKitBuilder) WithLogger(log *log.Logger) OpenKitBuilder {
	ob.log = log
	return ob
}

func (ob *openKitBuilder) WithApplicationVersion(applicationVersion string) OpenKitBuilder {
	ob.applicationVersion = applicationVersion
	return ob
}

func (ob *openKitBuilder) WithApplicationName(applicationName string) OpenKitBuilder {
	ob.applicationName = applicationName
	return ob
}

func (ob *openKitBuilder) WithOperatingSystem(operatingSystem string) OpenKitBuilder {
	ob.operatingSystem = operatingSystem
	return ob
}

func (ob *openKitBuilder) WithManufacturer(manufacturer string) OpenKitBuilder {
	ob.manufacturer = manufacturer
	return ob
}

func (ob *openKitBuilder) WithModelID(modelID string) OpenKitBuilder {
	ob.modelID = modelID
	return ob
}

func (ob *openKitBuilder) Build() OpenKit {
	// TODO - Set Defaults manually here if they were not set?

	c := NewConfiguration(ob.endpointURL, ob.applicationName, ob.applicationID, ob.applicationVersion, ob.deviceID, ob.operatingSystem, ob.manufacturer, ob.modelID)
	client := NewHttpClient(ob.log, *c.httpClientConfiguration)

	b := NewBeaconSender(ob.log, c, client)

	openkit := &openkit{
		beaconCache:   NewBeaconCache(ob.log),
		beaconSender:  b,
		configuration: c,
		log:           ob.log,
	}

	openkit.log.Debug("Initializing OpenKit...")

	openkit.initialize()

	return openkit

}
