package openkitgo

import "github.com/op/go-logging"

type openKitType int

const (
	OpenKitTypeAPPMON    openKitType = 1
	OpenKitTypeDYNATRACE openKitType = 1
)

type OpenKit interface {
	CreateSession(string) *Session
	waitForInitCompletion(int) bool
	isInitialized() bool
}

type openkit struct {
	beaconSender  *BeaconSender
	configuration *Configuration
	logger        logging.Logger
}

func (*openkit) CreateSession(clientIPAddress string) *Session {

	// TODO Create Beacon
	return &Session{}
}

type OpenKitBuilder interface {
	WithLogLevel(int) OpenKitBuilder
	WithLogger(logging.Logger) OpenKitBuilder
	WithApplicationName(string) OpenKitBuilder
	WithApplicationVersion(string) OpenKitBuilder
	WithOperatingSystem(string) OpenKitBuilder
	WithManufacturer(string) OpenKitBuilder
	WithModelID(string) OpenKitBuilder

	Build() OpenKit
}

type openKitBuilder struct {
	logLevel int
	logger   logging.Logger

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

func (ob *openKitBuilder) WithLogLevel(logLevel int) OpenKitBuilder {
	ob.logLevel = logLevel
	return ob
}

func (ob *openKitBuilder) WithLogger(logger logging.Logger) OpenKitBuilder {
	ob.logger = logger
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

	return &openkit{
		// TODO - Implement BeaconSender
		beaconSender:  nil,
		configuration: c,
		logger:        ob.logger,
	}
}
