package openkitgo

import (
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo/configuration"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type OpenKitBuilder interface {
	WithApplicationName(applicationName string) OpenKitBuilder
	WithLogLevel(level log.Level) OpenKitBuilder
	WithLogger(log *log.Logger) OpenKitBuilder
	WithApplicationVersion(version string) OpenKitBuilder
	WithTransport(transport *http.Transport) OpenKitBuilder
	WithOperatingSystem(operatingSystem string) OpenKitBuilder
	WithManufacturer(manufacturer string) OpenKitBuilder
	WithModelID(modelID string) OpenKitBuilder
	WithBeaconCacheMaxRecordAge(maxRecordAge time.Duration) OpenKitBuilder
	WithBeaconCacheLowerMemoryBoundary(m int64) OpenKitBuilder
	WithBeaconCacheUpperMemoryBoundary(m int64) OpenKitBuilder
	WithDataCollectionLevel(l configuration.DataCollectionLevel) OpenKitBuilder
	WithCrashReportingLevel(l configuration.CrashReportingLevel) OpenKitBuilder
	Build() OpenKit
}

type OpenKit interface {
	WaitForInitCompletion() bool
	WaitForInitCompletionTimeout(duration time.Duration) bool
	Shutdown()

	CreateSession(clientIPAddress string) Session
	CreateSessionAt(clientIPAddress string, timestamp time.Time) Session

	CreateSessionWithDeviceID(clientIPAddress string, deviceID int64) Session
	CreateSessionAtWithDeviceID(clientIPAddress string, timestamp time.Time, deviceID int64) Session
}
