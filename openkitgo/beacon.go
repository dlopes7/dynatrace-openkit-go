package openkitgo

import (
	"github.com/op/go-logging"
	"strconv"
	"strings"
	"sync/atomic"
)

type BeaconSender struct {
	logger  logging.Logger
	context BeaconSenderContext
}

func NewBeaconSender(logger logging.Logger, config *Configuration, client *HttpClient) *BeaconSender {
	b := new(BeaconSender)
	b.logger = logger

	b.context = BeaconSenderContext{
		logger:       logger,
		config:       config,
		httpClient:   client,
		shutdown:     false,
		currentState: new(beaconSendingInitState),
		sessions:     make(map[int]*session, 0),
	}

	return b
}

func (b *BeaconSender) initialize() {

	go func() {
		b.logger.Info("BeaconSender goroutine initialized")
		for !b.context.isTerminalState {
			b.context.executeCurrentState()
		}
	}()

}

func (b *BeaconSender) startSession(session *session) {
	b.logger.Debug("BeaconSender startSession()")

	b.context.startSession(session)

}

func (b *BeaconSender) finishSession(session Session) {
	b.logger.Debug("BeaconSender finishSession()")

	b.context.finishSession(session)

}

var RESERVED_CHARACTERS = []rune{'_'}

const (
	// basic data constants
	BEACON_KEY_PROTOCOL_VERSION      = "vv"
	BEACON_KEY_OPENKIT_VERSION       = "va"
	BEACON_KEY_APPLICATION_ID        = "ap"
	BEACON_KEY_APPLICATION_NAME      = "an"
	BEACON_KEY_APPLICATION_VERSION   = "vn"
	BEACON_KEY_PLATFORM_TYPE         = "pt"
	BEACON_KEY_AGENT_TECHNOLOGY_TYPE = "tt"
	BEACON_KEY_VISITOR_ID            = "vi"
	BEACON_KEY_SESSION_NUMBER        = "sn"
	BEACON_KEY_CLIENT_IP_ADDRESS     = "ip"
	BEACON_KEY_MULTIPLICITY          = "mp"
	BEACON_KEY_DATA_COLLECTION_LEVEL = "dl"
	BEACON_KEY_CRASH_REPORTING_LEVEL = "cl"

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
	BEACON_KEY_VALUE                     = "vl"
	BEACON_KEY_ERROR_CODE                = "ev"
	BEACON_KEY_ERROR_REASON              = "rs"
	BEACON_KEY_ERROR_STACKTRACE          = "st"
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

	BEACON_DATA_DELIMITER = "&"
)

type Beacon struct {
	logger             logging.Logger
	beaconCache        *beaconCache
	config             *Configuration
	clientIPAddress    string
	nextSequenceNumber uint64

	nextID           uint64
	sessionNumber    int
	sessionStartTime int

	beaconConfiguration BeaconConfiguration

	immutableBasicBeaconData string
}

func NewBeacon(logger logging.Logger, beaconCache *beaconCache, config *Configuration, clientIPAddress string) *Beacon {
	b := new(Beacon)

	b.sessionNumber = b.config.createSessionNumber()
	b.sessionStartTime = b.config.makeTimestamp()
	b.logger = logger
	b.beaconCache = beaconCache
	b.config = config
	b.clientIPAddress = clientIPAddress
	b.beaconConfiguration = *config.beaconConfiguration

	b.immutableBasicBeaconData = b.createImmutableBasicBeaconData()

	return b

}

func (b *Beacon) startSession() {

	var eventBuilder strings.Builder
	b.buildBasicEventData(&eventBuilder, EventTypeSESSION_START, "")

	b.addKeyValuePair(&eventBuilder, BEACON_KEY_PARENT_ACTION_ID, "0")
	b.addKeyValuePair(&eventBuilder, BEACON_KEY_START_SEQUENCE_NUMBER, strconv.Itoa(b.createSequenceNumber()))
	b.addKeyValuePair(&eventBuilder, BEACON_KEY_TIME_0, "0")

	b.addEventData(b.sessionStartTime, &eventBuilder)

}

func (b *Beacon) getCurrentTimestamp() int {
	return b.config.makeTimestamp()

}

func (b *Beacon) endSession(session *session) {

	var eventBuilder strings.Builder

	b.buildBasicEventData(&eventBuilder, EventTypeSESSION_END, "")

	b.addKeyValuePair(&eventBuilder, BEACON_KEY_PARENT_ACTION_ID, strconv.Itoa(0))
	b.addKeyValuePair(&eventBuilder, BEACON_KEY_START_SEQUENCE_NUMBER, strconv.Itoa(b.createSequenceNumber()))
	b.addKeyValuePair(&eventBuilder, BEACON_KEY_TIME_0, strconv.Itoa(b.getTimeSinceSessionStartTime(session.endTime)))

	b.addEventData(session.endTime, &eventBuilder)

}

func (b *Beacon) addAction(action *action) {

	var actionBuilder strings.Builder

	b.buildBasicEventData(&actionBuilder, EventTypeACTION, action.name)

	b.addKeyValuePair(&actionBuilder, BEACON_KEY_ACTION_ID, strconv.Itoa(action.ID))
	b.addKeyValuePair(&actionBuilder, BEACON_KEY_PARENT_ACTION_ID, strconv.Itoa(action.getParentActionID()))
	b.addKeyValuePair(&actionBuilder, BEACON_KEY_START_SEQUENCE_NUMBER, strconv.Itoa(action.startSequenceNo))
	b.addKeyValuePair(&actionBuilder, BEACON_KEY_TIME_0, strconv.Itoa(b.getTimeSinceSessionStartTime(action.startTime)))
	b.addKeyValuePair(&actionBuilder, BEACON_KEY_END_SEQUENCE_NUMBER, strconv.Itoa(action.endSequenceNo))
	b.addKeyValuePair(&actionBuilder, BEACON_KEY_TIME_1, strconv.Itoa(action.endTime-action.startTime))

	b.addActionData(action.startTime, &actionBuilder)
}

func (b *Beacon) addActionData(timestamp int, sb *strings.Builder) {
	b.beaconCache.addActionData(b.sessionNumber, timestamp, sb.String())
}

func (b *Beacon) getTimeSinceSessionStartTime(timestamp int) int {
	return timestamp - b.sessionStartTime
}

func (b *Beacon) createSequenceNumber() int {
	atomic.AddUint64(&b.nextSequenceNumber, 1)
	return int(b.nextSequenceNumber)
}

func (b *Beacon) createID() int {
	atomic.AddUint64(&b.nextID, 1)
	return int(b.nextID)
}

func (b *Beacon) buildBasicEventData(sb *strings.Builder, eventType EventType, name string) {

	b.addKeyValuePair(sb, BEACON_KEY_EVENT_TYPE, strconv.Itoa(int(eventType)))

	if len(name) != 0 {
		b.addKeyValuePair(sb, BEACON_KEY_NAME, b.truncate(name))
	}

	// TODO - Replace "1" with getThreadID()
	b.addKeyValuePair(sb, BEACON_KEY_THREAD_ID, "1")

}

func (b *Beacon) addEventData(timestamp int, sb *strings.Builder) {
	b.beaconCache.addEventData(b.sessionNumber, timestamp, sb.String())
}

func (b *Beacon) addKeyValuePair(sb *strings.Builder, key string, value string) {

	encodedValue := encodeWithReservedChars(value, CHARSET, nil)
	b.appendKey(sb, key)
	sb.WriteString(encodedValue)
}

func (b *Beacon) appendKey(sb *strings.Builder, key string) {
	if sb.Len() != 0 {
		sb.WriteString("&")
	}
	sb.WriteString(key)
	sb.WriteString("=")
}

func (b *Beacon) truncate(name string) string {
	name = strings.Trim(name, " ")
	if len(name) > MAX_NAME_LEN {
		runes := []rune(name)
		name = string(runes[:MAX_NAME_LEN])
	}

	return name
}

func (b *Beacon) addKeyValuePairIfNotNull(sb *strings.Builder, key string, value *string) {

	if value != nil {
		b.addKeyValuePair(sb, key, *value)
	}
}

func (b *Beacon) createImmutableBasicBeaconData() string {

	var basicBeaconBuilder strings.Builder

	// version and application information
	b.addKeyValuePair(&basicBeaconBuilder, BEACON_KEY_PROTOCOL_VERSION, strconv.Itoa(PROTOCOL_VERSION))
	b.addKeyValuePair(&basicBeaconBuilder, BEACON_KEY_OPENKIT_VERSION, OPENKIT_VERSION)
	b.addKeyValuePair(&basicBeaconBuilder, BEACON_KEY_APPLICATION_ID, b.config.applicationID)
	b.addKeyValuePair(&basicBeaconBuilder, BEACON_KEY_APPLICATION_NAME, b.config.applicationName)
	b.addKeyValuePairIfNotNull(&basicBeaconBuilder, BEACON_KEY_APPLICATION_VERSION, &b.config.applicationVersion)
	b.addKeyValuePair(&basicBeaconBuilder, BEACON_KEY_PLATFORM_TYPE, strconv.Itoa(PLATFORM_TYPE_OPENKIT))
	b.addKeyValuePair(&basicBeaconBuilder, BEACON_KEY_AGENT_TECHNOLOGY_TYPE, AGENT_TECHNOLOGY_TYPE)

	// device/visitor ID, session number and IP address
	b.addKeyValuePair(&basicBeaconBuilder, BEACON_KEY_VISITOR_ID, b.config.deviceID)
	b.addKeyValuePair(&basicBeaconBuilder, BEACON_KEY_SESSION_NUMBER, strconv.Itoa(b.sessionNumber))
	b.addKeyValuePair(&basicBeaconBuilder, BEACON_KEY_CLIENT_IP_ADDRESS, b.clientIPAddress)

	// platform information
	b.addKeyValuePairIfNotNull(&basicBeaconBuilder, BEACON_KEY_DEVICE_OS, &b.config.device.operatingSystem)
	b.addKeyValuePairIfNotNull(&basicBeaconBuilder, BEACON_KEY_DEVICE_MANUFACTURER, &b.config.device.manufacturer)
	b.addKeyValuePairIfNotNull(&basicBeaconBuilder, BEACON_KEY_DEVICE_MODEL, &b.config.device.modelID)

	beaconConfig := b.beaconConfiguration
	b.addKeyValuePair(&basicBeaconBuilder, BEACON_KEY_DATA_COLLECTION_LEVEL, strconv.Itoa(beaconConfig.dataCollectionLevel))
	b.addKeyValuePair(&basicBeaconBuilder, BEACON_KEY_CRASH_REPORTING_LEVEL, strconv.Itoa(beaconConfig.crashReportingLevel))

	return basicBeaconBuilder.String()
}

func (b *Beacon) appendMutableBeaconData(immutableBasicBeaconData *string) string {

	var mutableBeaconDataBuilder strings.Builder

	if immutableBasicBeaconData != nil && len(*immutableBasicBeaconData) > 0 {

		mutableBeaconDataBuilder.WriteString(*immutableBasicBeaconData)
		mutableBeaconDataBuilder.WriteString(BEACON_DATA_DELIMITER)
	}

	// append timestamp data
	mutableBeaconDataBuilder.WriteString(b.createTimestampData())

	// append multiplicity
	mutableBeaconDataBuilder.WriteString(BEACON_DATA_DELIMITER)
	mutableBeaconDataBuilder.WriteString(b.createMultiplicityData())

	return mutableBeaconDataBuilder.String()

}

func (b *Beacon) createTimestampData() string {
	var sb strings.Builder

	b.addKeyValuePair(&sb, BEACON_KEY_TRANSMISSION_TIME, strconv.Itoa(b.config.makeTimestamp()))
	b.addKeyValuePair(&sb, BEACON_KEY_SESSION_START_TIME, strconv.Itoa(b.sessionStartTime))

	return sb.String()
}

func (b *Beacon) createMultiplicityData() string {
	var sb strings.Builder

	b.addKeyValuePair(&sb, BEACON_KEY_MULTIPLICITY, strconv.Itoa(b.beaconConfiguration.multiplicity))

	return sb.String()
}

func (b *Beacon) send(client *HttpClient) *StatusResponse {

	// TODO remove all this horrible chunk stuff

	var response *StatusResponse

	prefix := b.appendMutableBeaconData(&b.immutableBasicBeaconData)

	for {
		chunk := b.beaconCache.getNextBeaconChunk(b.sessionNumber, prefix, b.config.maxBeaconSize-1024, BEACON_DATA_DELIMITER)
		if chunk == nil || *chunk == "" {
			return response

		}

		encodedBeacon := []byte(*chunk)

		response = client.sendBeaconRequest(b.clientIPAddress, encodedBeacon)

		if response == nil || response.responseCode >= 400 {
			b.beaconCache.resetChunkedData(b.sessionNumber)
		} else {
			b.beaconCache.removeChunkedData(b.sessionNumber)
		}
	}

	return response

}
