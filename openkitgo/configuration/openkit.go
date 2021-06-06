package configuration

import (
	"net/http"
	"time"
)

const (
	ENCODING_CHARSET                       = "UTF-8"
	RESERVED_CHARACTERS                    = "_"
	DEFAULT_MAX_RECORD_AGE                 = 105 * time.Minute
	DEFAULT_UPPER_MEMORY_BOUNDARY_IN_BYTES = int64(100 * 1024 * 1024)
	DEFAULT_LOWER_MEMORY_BOUNDARY_IN_BYTES = int64(80 * 1024 * 1024)
	DEFAULT_DATA_COLLECTION_LEVEL          = DATA_USER_BEHAVIOR
	DEFAULT_CRASH_REPORTING_LEVEL          = CRASH_OPT_IN_CRASHES
)

type OpenKitConfiguration struct {
	EndpointURL                 string
	DeviceID                    int64
	OrigDeviceID                string
	OpenKitType                 string
	ApplicationID               string
	PercentEncodedApplicationID string
	ApplicationName             string
	ApplicationVersion          string
	OperatingSystem             string
	Manufacturer                string
	ModelID                     string
	DefaultServerID             int
	Transport                   *http.Transport
}
