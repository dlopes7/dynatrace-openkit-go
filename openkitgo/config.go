package openkitgo

import (
	"math/rand"
	"strconv"
	"time"
)

const DEFAULT_SEND_INTERVAL = 2 * 60 * 1000 // default: wait 2m (in ms) to send beacon
const DEFAULT_MAX_BEACON_SIZE = 30 * 1024   // default: max 30KB (in B) to send in one beacon
const DEFAULT_CAPTURE = true
const DEFAULT_CAPTURE_ERRORS = true
const DEFAULT_CAPTURE_CRASHES = true

const DEFAULT_DATA_COLLECTION_LEVEL = 2
const DEFAULT_CRASH_REPORTING_LEVEL = 2

type Configuration struct {
	openKitType openKitType

	applicationName string
	applicationID   string
	deviceID        string
	endpointURL     string

	capture        bool
	sendInterval   int
	maxBeaconSize  int
	captureErrors  bool
	captureCrashes bool

	device *Device

	applicationVersion string

	httpClientConfiguration *HTTPClientConfiguration
	beaconConfiguration     *BeaconConfiguration
}

func NewConfiguration(endpointURL string, applicationName string, applicationID string, applicationVersion string, deviceID int, operatingSystem string, manufacturer string, modelID string) *Configuration {

	// TODO - Implement BeaconCacheConfiguration
	// TODO - Implement BeaconConfiguration
	// TODO - Implement DefaultSessionIDProvider
	// TODO - Implement getTrustManager

	c := new(Configuration)
	c.endpointURL = endpointURL
	c.applicationName = applicationName
	c.applicationID = applicationID
	c.applicationVersion = applicationVersion
	c.deviceID = strconv.Itoa(deviceID)

	d := &Device{
		operatingSystem: operatingSystem,
		manufacturer:    manufacturer,
		modelID:         modelID,
	}

	c.httpClientConfiguration = &HTTPClientConfiguration{
		serverID:      1, // TODO - This might change in the future
		applicationID: applicationID,
		baseURL:       endpointURL,
	}

	c.beaconConfiguration = &BeaconConfiguration{
		multiplicity:        1,
		dataCollectionLevel: DEFAULT_DATA_COLLECTION_LEVEL,
		crashReportingLevel: DEFAULT_CRASH_REPORTING_LEVEL,
	}

	c.device = d

	return c
}

func (c *Configuration) createSessionNumber() int {
	return rand.Intn(2147483647)
}

func (c *Configuration) makeTimestamp() int {
	return int(time.Now().UnixNano() / int64(time.Millisecond))
}

type HTTPClientConfiguration struct {
	baseURL       string
	applicationID string
	serverID      int
}

type BeaconConfiguration struct {
	multiplicity        int
	dataCollectionLevel int
	crashReportingLevel int
}
