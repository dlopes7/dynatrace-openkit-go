package core

import (
	"crypto/tls"
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo/configuration"
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo/interfaces"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"time"
)

const (
	DEFAULT_SERVER_ID           = 1
	OPENKIT_TYPE                = "DynatraceOpenKitGo"
	WEBREQUEST_TAG_HEADER       = "X-dynaTrace"
	DEFAULT_APPLICATION_VERSION = "1.1.2"
	DEFAULT_OPERATING_SYSTEM    = "OpenKit " + DEFAULT_APPLICATION_VERSION
	DEFAULT_MANUFACTURER        = "Dynatrace"
	DEFAULT_MODEL_ID            = "OpenKitDevice"
)

type OpenKitBuilder struct {
	endpointURL  string
	deviceID     int64
	origDeviceID string

	log                            *log.Logger
	transport                      *http.Transport
	logLevel                       log.Level
	operatingSystem                string
	manufacturer                   string
	modelID                        string
	applicationVersion             string
	beaconCacheMaxRecordAge        time.Duration
	beaconCacheLowerMemoryBoundary int64
	beaconCacheUpperMemoryBoundary int64
	dataCollectionLevel            configuration.DataCollectionLevel
	crashReportLevel               configuration.CrashReportingLevel

	applicationID   string
	applicationName string
}

func (b *OpenKitBuilder) ApplicationID() string {
	return b.applicationID
}

func (b *OpenKitBuilder) ApplicationName() string {
	return b.applicationName
}

func NewOpenKitBuilder(endpointURL string, applicationID string, deviceID int64) interfaces.OpenKitBuilder {

	return &OpenKitBuilder{
		endpointURL:                    endpointURL,
		applicationID:                  applicationID,
		deviceID:                       deviceID,
		origDeviceID:                   strconv.FormatInt(deviceID, 10),
		log:                            log.New(),
		transport:                      &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
		operatingSystem:                DEFAULT_OPERATING_SYSTEM,
		manufacturer:                   DEFAULT_MANUFACTURER,
		modelID:                        DEFAULT_MODEL_ID,
		applicationVersion:             DEFAULT_APPLICATION_VERSION,
		beaconCacheMaxRecordAge:        configuration.DEFAULT_MAX_RECORD_AGE,
		beaconCacheLowerMemoryBoundary: configuration.DEFAULT_LOWER_MEMORY_BOUNDARY_IN_BYTES,
		beaconCacheUpperMemoryBoundary: configuration.DEFAULT_UPPER_MEMORY_BOUNDARY_IN_BYTES,
		dataCollectionLevel:            configuration.DEFAULT_DATA_COLLECTION_LEVEL,
		crashReportLevel:               configuration.DEFAULT_CRASH_REPORTING_LEVEL,
	}

}

func (b *OpenKitBuilder) WithApplicationName(applicationName string) interfaces.OpenKitBuilder {
	b.applicationName = applicationName
	return b
}

func (b *OpenKitBuilder) WithLogLevel(level log.Level) interfaces.OpenKitBuilder {
	b.logLevel = level
	b.log.SetLevel(level)
	return b
}

func (b *OpenKitBuilder) WithLogger(log *log.Logger) interfaces.OpenKitBuilder {
	b.log = log
	return b
}

func (b *OpenKitBuilder) WithApplicationVersion(version string) interfaces.OpenKitBuilder {
	b.applicationVersion = version
	return b
}

func (b *OpenKitBuilder) WithTransport(transport *http.Transport) interfaces.OpenKitBuilder {
	if transport == nil {
		transport = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	}
	b.transport = transport
	return b
}

func (b *OpenKitBuilder) WithOperatingSystem(operatingSystem string) interfaces.OpenKitBuilder {
	b.operatingSystem = operatingSystem
	return b
}

func (b *OpenKitBuilder) WithManufacturer(manufacturer string) interfaces.OpenKitBuilder {
	b.manufacturer = manufacturer
	return b
}

func (b *OpenKitBuilder) WithModelID(modelID string) interfaces.OpenKitBuilder {
	b.modelID = modelID
	return b
}

func (b *OpenKitBuilder) WithBeaconCacheMaxRecordAge(maxRecordAge time.Duration) interfaces.OpenKitBuilder {
	b.beaconCacheMaxRecordAge = maxRecordAge
	return b
}

func (b *OpenKitBuilder) WithBeaconCacheLowerMemoryBoundary(m int64) interfaces.OpenKitBuilder {
	b.beaconCacheLowerMemoryBoundary = m
	return b
}

func (b *OpenKitBuilder) WithBeaconCacheUpperMemoryBoundary(m int64) interfaces.OpenKitBuilder {
	b.beaconCacheUpperMemoryBoundary = m
	return b
}

func (b *OpenKitBuilder) WithDataCollectionLevel(l configuration.DataCollectionLevel) interfaces.OpenKitBuilder {
	b.dataCollectionLevel = l
	return b
}

func (b *OpenKitBuilder) WithCrashReportingLevel(l configuration.CrashReportingLevel) interfaces.OpenKitBuilder {
	b.crashReportLevel = l
	return b
}

func (b *OpenKitBuilder) Build() interfaces.OpenKit {

	openKit := NewOpenKit(b).(*OpenKit)
	openKit.initialize()

	log.WithFields(log.Fields{"instance": openKit.String()}).Info("OpenKit instantiated")
	log.WithFields(log.Fields{"instance": openKit.DetailedString()}).Debug("OpenKit instantiated")

	return openKit

}
