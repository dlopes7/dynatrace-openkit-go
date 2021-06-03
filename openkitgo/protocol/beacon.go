package protocol

import (
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo/caching"
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo/configuration"
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo/core"
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo/providers"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"time"
)

const (
	BEACON_KEY_PROTOCOL_VERSION      = "vv"
	BEACON_KEY_OPENKIT_VERSION       = "va"
	BEACON_KEY_APPLICATION_ID        = "ap"
	BEACON_KEY_APPLICATION_NAME      = "an"
	BEACON_KEY_APPLICATION_VERSION   = "vn"
	BEACON_KEY_PLATFORM_TYPE         = "pt"
	BEACON_KEY_AGENT_TECHNOLOGY_TYPE = "tt"
	BEACON_KEY_VISITOR_ID            = "vi"
	BEACON_KEY_SESSION_NUMBER        = "sn"
	BEACON_KEY_SESSION_SEQUENCE      = "ss"
	BEACON_KEY_CLIENT_IP_ADDRESS     = "ip"
	BEACON_KEY_MULTIPLICITY          = "mp"
	BEACON_KEY_DATA_COLLECTION_LEVEL = "dl"
	BEACON_KEY_CRASH_REPORTING_LEVEL = "cl"
	BEACON_KEY_VISIT_STORE_VERSION   = "vs"

	// device data constants
	BEACON_KEY_DEVICE_OS           = "os"
	BEACON_KEY_DEVICE_MANUFACTURER = "mf"
	BEACON_KEY_DEVICE_MODEL        = "md"

	// timestamp constants
	BEACON_KEY_SESSION_START_TIME = "tv"
	BEACON_KEY_TRANSMISSION_TIME  = "tx"

	// Action related constants
	BEACON_KEY_EVENT_TYPE            = "et"
	BEACON_KEY_NAME                  = "na"
	BEACON_KEY_THREAD_ID             = "it"
	BEACON_KEY_ACTION_ID             = "ca"
	BEACON_KEY_PARENT_ACTION_ID      = "pa"
	BEACON_KEY_START_SEQUENCE_NUMBER = "s0"
	BEACON_KEY_TIME_0                = "t0"
	BEACON_KEY_END_SEQUENCE_NUMBER   = "s1"
	BEACON_KEY_TIME_1                = "t1"

	// data, error & crash capture constants
	BEACON_KEY_VALUE                 = "vl"
	BEACON_KEY_ERROR_VALUE           = "ev" // can be an integer code or string (Exception class name
	BEACON_KEY_ERROR_REASON          = "rs"
	BEACON_KEY_ERROR_STACKTRACE      = "st"
	BEACON_KEY_ERROR_TECHNOLOGY_TYPE = "tt"

	// web request constants
	BEACON_KEY_WEBREQUEST_RESPONSECODE   = "rc"
	BEACON_KEY_WEBREQUEST_BYTES_SENT     = "bs"
	BEACON_KEY_WEBREQUEST_BYTES_RECEIVED = "br"

	// in Java 6 there is no constant for "UTF-8" in the JDK yet, so we define it ourselves
	CHARSET = "UTF-8"

	// max name length
	MAX_NAME_LEN = 250

	// web request tag prefix constant
	TAG_PREFIX = "MT"

	// web request tag reserved characters
	RESERVED_CHARACTERS = '_'

	BEACON_DATA_DELIMITER = '&'
)

type Beacon struct {
	nextID             uint32 // Atomic
	nextSequenceNumber uint32 // Atomic
	beaconKey          caching.BeaconKey
	sessionStartTime   time.Time
	deviceID           int
	clientIPAddress    string

	immutableBasicBeaconData string
	configuration            configuration.Beacon
	trafficControlValue      int
	log                      *log.Logger
	cache                    *caching.BeaconCache
	sessionIDProvider        *providers.SessionIDProvider
}

func NewBeacon(
	log *log.Logger,
	beaconCache *caching.BeaconCache,
	sessionIDProvider *providers.SessionIDProvider,
	sessionProxy *core.SessionProxy,
	beaconConfiguration configuration.Beacon,
	sessionStartTime time.Time,
	deviceID int,
	ipAddress string,

) *Beacon {
	sessionNumber := sessionIDProvider.GetNextSessionID()
	sessionSequenceNumber := sessionProxy.GetSessionSequenceNumber()

	return &Beacon{
		nextID:                   0,
		nextSequenceNumber:       0,
		beaconKey:                caching.NewBeaconKey(sessionNumber, sessionSequenceNumber),
		sessionStartTime:         sessionStartTime,
		deviceID:                 deviceID,
		clientIPAddress:          ipAddress,
		immutableBasicBeaconData: "",
		configuration:            beaconConfiguration,
		trafficControlValue:      rand.Intn(100),
		log:                      log,
		cache:                    beaconCache,
		sessionIDProvider:        sessionIDProvider,
	}

}
