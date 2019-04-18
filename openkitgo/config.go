package openkitgo

import "strconv"

const DEFAULT_SEND_INTERVAL = 2 * 60 * 1000 // default: wait 2m (in ms) to send beacon
const DEFAULT_MAX_BEACON_SIZE = 30 * 1024   // default: max 30KB (in B) to send in one beacon
const DEFAULT_CAPTURE = true
const DEFAULT_CAPTURE_ERRORS = true
const DEFAULT_CAPTURE_CRASHES = true

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

	c.device = d

	return c
}
