package core

import (
	"fmt"
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo/caching"
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo/configuration"
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo/interfaces"
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo/protocol"
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo/providers"
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo/utils"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"strings"
	"sync/atomic"
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

type EventType int

const (
	ACTION        EventType = 1
	VALUE_STRING  EventType = 11
	VALUE_INT     EventType = 12
	VALUE_DOUBLE  EventType = 13
	NAMED_EVENT   EventType = 10
	SESSION_START EventType = 18
	SESSION_END   EventType = 19
	WEB_REQUEST   EventType = 30
	ERROR         EventType = 40
	EXCEPTION     EventType = 42
	CRASH         EventType = 50
	IDENTIFY_USER EventType = 60
)

type Beacon struct {
	nextID             int32 // Atomic
	nextSequenceNumber int32 // Atomic
	key                caching.BeaconKey
	sessionStartTime   time.Time
	deviceID           int64
	clientIPAddress    string

	immutableBasicBeaconData string
	configuration            *configuration.BeaconConfiguration
	trafficControlValue      int
	log                      *log.Logger
	cache                    *caching.BeaconCache
	sessionIDProvider        *providers.SessionIDProvider
}

func NewBeacon(
	log *log.Logger,
	beaconCache *caching.BeaconCache,
	sessionIDProvider *providers.SessionIDProvider,
	sessionProxy *SessionProxy,
	beaconConfiguration *configuration.BeaconConfiguration,
	sessionStartTime time.Time,
	deviceID int64,
	ipAddress string,

) *Beacon {
	sessionNumber := sessionIDProvider.GetNextSessionID()
	sessionSequenceNumber := sessionProxy.GetSessionSequenceNumber()

	b := &Beacon{
		nextID:              0,
		nextSequenceNumber:  0,
		key:                 caching.NewBeaconKey(sessionNumber, sessionSequenceNumber),
		sessionStartTime:    sessionStartTime,
		deviceID:            deviceID,
		clientIPAddress:     ipAddress,
		configuration:       beaconConfiguration,
		trafficControlValue: rand.Intn(100),
		log:                 log,
		cache:               beaconCache,
		sessionIDProvider:   sessionIDProvider,
	}
	b.immutableBasicBeaconData = b.createImmutableBasicBeaconData()

	return b

}
func (b *Beacon) EndSession() {
	b.EndSessionAt(time.Now())
}

func (b *Beacon) CreateID() int32 {
	return atomic.AddInt32(&b.nextID, 1)
}
func (b *Beacon) CreateSequenceNumber() int32 {
	return atomic.AddInt32(&b.nextSequenceNumber, 1)
}

func (b *Beacon) GetSessionStartTime() time.Time {
	return b.sessionStartTime
}

func (b *Beacon) CreateTag(parentActionID int, tracerSeqNo int) string {

	if !b.configuration.PrivacyConfiguration.IsWebRequestTracingAllowed() {
		return ""
	}

	serverID := b.configuration.HttpClientConfiguration.ServerID
	var builder strings.Builder

	builder.WriteString(TAG_PREFIX)
	builder.WriteString(fmt.Sprintf("_%d", protocol.PROTOCOL_VERSION))
	builder.WriteString(fmt.Sprintf("_%d", serverID))
	builder.WriteString(fmt.Sprintf("_%d", b.deviceID))
	builder.WriteString(fmt.Sprintf("_%d", b.GetSessionNumber()))
	if b.GetVisitStoreVersion() > 1 {
		builder.WriteString(fmt.Sprintf("-%d", b.key.BeaconSeqNo))
	}
	builder.WriteString(fmt.Sprintf("_%s", b.configuration.OpenKitConfiguration.PercentEncodedApplicationID))
	builder.WriteString(fmt.Sprintf("_%d", parentActionID))
	builder.WriteString("_1")
	builder.WriteString(fmt.Sprintf("_%d", tracerSeqNo))

	return builder.String()
}

func (b *Beacon) AddAction(action *Action) {
	b.AddActionAt(action, time.Now())
}

func (b *Beacon) AddActionAt(action *Action, timestamp time.Time) {
	if action == nil || action.name == "" {
		b.log.Error("Attempted to start an action without a name.")
		return
	}

	if !b.configuration.PrivacyConfiguration.IsActionReportingAllowed() {
		return
	}

	if !b.isDataCapturingEnabled() {
		return
	}

	var builder strings.Builder

	b.buildBasicEventData(&builder, ACTION, action.name)

	b.addKeyValuePair(&builder, BEACON_KEY_ACTION_ID, action.id)
	b.addKeyValuePair(&builder, BEACON_KEY_PARENT_ACTION_ID, action.parentActionID)
	b.addKeyValuePair(&builder, BEACON_KEY_START_SEQUENCE_NUMBER, action.startSequenceNo)
	b.addKeyValuePair(&builder, BEACON_KEY_TIME_0, timestamp.Sub(b.sessionStartTime).Milliseconds())
	b.addKeyValuePair(&builder, BEACON_KEY_END_SEQUENCE_NUMBER, action.endSequenceNo)
	b.addKeyValuePair(&builder, BEACON_KEY_TIME_1, action.endTime.Sub(action.startTime).Milliseconds())

	b.addActionData(timestamp, &builder)

}

func (b *Beacon) addActionData(timestamp time.Time, builder *strings.Builder) {
	if b.isDataCapturingEnabled() {
		b.cache.AddActionData(b.key, timestamp, builder.String())
	}
}

func (b *Beacon) GetSessionNumber() int {
	if b.configuration.PrivacyConfiguration.IsSessionNumberReportingAllowed() {
		return int(b.key.BeaconId)
	}
	return 1
}

func (b *Beacon) GetVisitStoreVersion() int {
	return b.configuration.GetServerConfiguration().VisitStoreVersion

}

func (b *Beacon) EndSessionAt(timestamp time.Time) {

	if !b.isDataCapturingEnabled() {
		return
	}

	var builder strings.Builder
	b.buildBasicEventDataWithoutName(&builder, SESSION_END)
	b.addKeyValuePair(&builder, BEACON_KEY_PARENT_ACTION_ID, 0)
	b.addKeyValuePair(&builder, BEACON_KEY_START_SEQUENCE_NUMBER, atomic.AddInt32(&b.nextSequenceNumber, 1))

	sessionDuration := timestamp.Sub(b.sessionStartTime).Milliseconds()
	b.addKeyValuePair(&builder, BEACON_KEY_TIME_0, sessionDuration)

	b.addEventData(timestamp, &builder)

}

func (b *Beacon) isDataCapturingEnabled() bool {
	s := b.configuration.GetServerConfiguration()
	return s.IsSendingDataAllowed() && b.trafficControlValue < s.TrafficControlPercentage

}

func (b *Beacon) isErrorCapturingEnabled() bool {
	s := b.configuration.GetServerConfiguration()
	return s.IsSendingErrorsAllowed() && b.trafficControlValue < s.TrafficControlPercentage

}

func (b *Beacon) buildBasicEventDataWithoutName(builder *strings.Builder, eventType EventType) {

	b.addKeyValuePair(builder, BEACON_KEY_EVENT_TYPE, eventType)
	b.addKeyValuePair(builder, BEACON_KEY_THREAD_ID, 1)

}

func (b *Beacon) addKeyValuePair(builder *strings.Builder, key string, value interface{}) {
	b.appendKey(builder, key)
	builder.WriteString(fmt.Sprintf("%v", value))
}

func (b *Beacon) appendKey(builder *strings.Builder, key string) {
	if builder.Len() > 0 {
		builder.WriteRune('&')
	}
	builder.WriteString(key)
	builder.WriteRune('=')

}

func (b *Beacon) addEventData(timestamp time.Time, builder *strings.Builder) {

	if b.isDataCapturingEnabled() {
		b.cache.AddEventData(b.key, timestamp, builder.String())
	}
}

func (b *Beacon) ClearData() {
	b.cache.DeleteCacheEntry(b.key)
}

func (b *Beacon) buildEvent(builder *strings.Builder, eventType EventType, name string, parentActionID int, timestamp time.Time) {
	b.buildBasicEventData(builder, eventType, name)

	b.addKeyValuePair(builder, BEACON_KEY_PARENT_ACTION_ID, parentActionID)
	b.addKeyValuePair(builder, BEACON_KEY_START_SEQUENCE_NUMBER, atomic.AddInt32(&b.nextSequenceNumber, 1))
	b.addKeyValuePair(builder, BEACON_KEY_TIME_0, timestamp.Sub(b.sessionStartTime).Milliseconds())

}

func (b *Beacon) buildBasicEventData(builder *strings.Builder, eventType EventType, name string) {
	b.buildBasicEventDataWithoutName(builder, eventType)
	b.addKeyValuePair(builder, BEACON_KEY_NAME, truncate(name))
}

func (b *Beacon) createImmutableBasicBeaconData() string {

	config := b.configuration.OpenKitConfiguration

	var builder strings.Builder

	b.addKeyValuePair(&builder, BEACON_KEY_PROTOCOL_VERSION, protocol.PROTOCOL_VERSION)
	b.addKeyValuePair(&builder, BEACON_KEY_OPENKIT_VERSION, protocol.OPENKIT_VERSION)
	b.addKeyValuePair(&builder, BEACON_KEY_APPLICATION_ID, config.ApplicationID)
	b.addKeyValuePairIfNotNull(&builder, BEACON_KEY_APPLICATION_NAME, config.ApplicationName)
	b.addKeyValuePair(&builder, BEACON_KEY_APPLICATION_VERSION, config.ApplicationVersion)
	b.addKeyValuePair(&builder, BEACON_KEY_PLATFORM_TYPE, protocol.PLATFORM_TYPE_OPENKIT)
	b.addKeyValuePair(&builder, BEACON_KEY_AGENT_TECHNOLOGY_TYPE, protocol.AGENT_TECHNOLOGY_TYPE)

	b.addKeyValuePair(&builder, BEACON_KEY_VISITOR_ID, b.deviceID)
	b.addKeyValuePair(&builder, BEACON_KEY_SESSION_NUMBER, b.GetSessionNumber())
	b.addKeyValuePair(&builder, BEACON_KEY_CLIENT_IP_ADDRESS, b.clientIPAddress)

	b.addKeyValuePairIfNotNull(&builder, BEACON_KEY_DEVICE_OS, config.OperatingSystem)
	b.addKeyValuePairIfNotNull(&builder, BEACON_KEY_DEVICE_MANUFACTURER, config.Manufacturer)
	b.addKeyValuePairIfNotNull(&builder, BEACON_KEY_DEVICE_MODEL, config.ModelID)

	privacyConfig := b.configuration.PrivacyConfiguration

	b.addKeyValuePair(&builder, BEACON_KEY_DATA_COLLECTION_LEVEL, privacyConfig.DataCollectionLevel)
	b.addKeyValuePair(&builder, BEACON_KEY_CRASH_REPORTING_LEVEL, privacyConfig.CrashReportingLevel)

	return builder.String()
}

func (b *Beacon) createTimestampData() string {

	var builder strings.Builder

	b.addKeyValuePair(&builder, BEACON_KEY_TRANSMISSION_TIME, utils.TimeToMillis(time.Now()))
	b.addKeyValuePair(&builder, BEACON_KEY_SESSION_START_TIME, utils.TimeToMillis(b.sessionStartTime))

	return builder.String()
}

func (b *Beacon) createMultiplicityData() string {

	var builder strings.Builder

	b.addKeyValuePair(&builder, BEACON_KEY_MULTIPLICITY, b.configuration.ServerConfiguration.Multiplicity)

	return builder.String()
}

func (b *Beacon) addKeyValuePairIfNotNull(builder *strings.Builder, key string, value interface{}) {

	send := true
	switch value.(type) {
	case string:
		if value == "" {
			send = false
		}
	case int:
		if value == 0 {
			send = false
		}
	case int64:
		if value == 0 {
			send = false
		}
	}

	if send {
		b.addKeyValuePair(builder, key, value)
	}
}

func (b *Beacon) addKeyValuePairIfNotNegative(builder *strings.Builder, key string, value interface{}) {

	switch v := value.(type) {
	case int:
		if v > 0 {
			b.addKeyValuePair(builder, key, value)
		}
	}

}

func (b *Beacon) IsEmpty() bool {
	return b.cache.IsEmpty(b.key)
}

func (b *Beacon) isServerConfigurationSet() bool {
	return b.configuration.IsServerConfigurationSet()
}

func (b *Beacon) disableCapture() {
	b.configuration.DisableCapture()
}

func (b *Beacon) updateServerConfiguration(config *configuration.ServerConfiguration) {
	b.configuration.UpdateServerConfiguration(config)
}

func (b *Beacon) send(ctx *BeaconSendingContext) protocol.StatusResponse {
	httpClient := ctx.getHttpClient()

	statusResponse := protocol.StatusResponse{}

	b.cache.PrepareDataForSending(b.key)
	for b.cache.HasDataForSending(b.key) {
		prefix := b.appendMutableBeaconData(b.immutableBasicBeaconData)

		chunk := b.cache.GetNextBeaconChunk(b.key, prefix, b.configuration.ServerConfiguration.BeaconSizeInBytes-1024, BEACON_DATA_DELIMITER)

		if chunk == "" {
			return statusResponse
		}

		statusResponse := httpClient.sendBeaconRequest(b.clientIPAddress, []byte(chunk), ctx)
		if statusResponse.ResponseCode > 400 {
			b.cache.ResetChunkedData(b.key)
			break
		} else {
			b.cache.RemoveChunkedData(b.key)
		}
	}
	return statusResponse
}

func (b *Beacon) appendMutableBeaconData(immutableBasicBeaconData string) string {

	var builder strings.Builder

	builder.WriteString(immutableBasicBeaconData)

	b.addKeyValuePair(&builder, BEACON_KEY_VISIT_STORE_VERSION, b.GetVisitStoreVersion())
	if b.GetVisitStoreVersion() > 1 {
		b.addKeyValuePair(&builder, BEACON_KEY_SESSION_SEQUENCE, b.key.BeaconSeqNo)
	}
	builder.WriteRune(BEACON_DATA_DELIMITER)
	builder.WriteString(b.createTimestampData())
	builder.WriteRune(BEACON_DATA_DELIMITER)
	builder.WriteString(b.createMultiplicityData())
	return builder.String()
}

func (b *Beacon) setServerConfigurationUpdateCallback(callback ServerConfigurationUpdateCallback) {
	b.configuration.SetServerConfigurationUpdateCallback(callback.onServerConfigurationUpdate)
}

func (b *Beacon) startSession() {
	if !b.isDataCapturingEnabled() {
		return
	}

	var builder strings.Builder

	b.buildBasicEventDataWithoutName(&builder, SESSION_START)

	b.addKeyValuePair(&builder, BEACON_KEY_PARENT_ACTION_ID, 0)
	b.addKeyValuePair(&builder, BEACON_KEY_START_SEQUENCE_NUMBER, b.CreateSequenceNumber())
	b.addKeyValuePair(&builder, BEACON_KEY_TIME_0, 0)

	b.addEventData(b.sessionStartTime, &builder)

}

func (b *Beacon) reportEvent(parentActionID int, eventName string, timestamp time.Time) {
	if !b.isDataCapturingEnabled() {
		return
	}

	var builder strings.Builder
	b.buildEvent(&builder, NAMED_EVENT, eventName, parentActionID, timestamp)
	b.addEventData(timestamp, &builder)
}
func (b *Beacon) reportValue(parentActionID int, valueName string, value interface{}, timestamp time.Time) {
	if !b.isDataCapturingEnabled() {
		return
	}

	valueType := VALUE_STRING
	switch value.(type) {
	case string:
		valueType = VALUE_STRING
	case int, int64, int32, uint32, uint64:
		valueType = VALUE_INT
	case float64, float32:
		valueType = VALUE_DOUBLE
	}

	var builder strings.Builder
	b.buildEvent(&builder, valueType, valueName, parentActionID, timestamp)
	b.addKeyValuePair(&builder, BEACON_KEY_VALUE, value)
	b.addEventData(timestamp, &builder)

}

func (b *Beacon) reportError(parentActionID int, errorName string, causeName string, causeDescription string, causeStackTrace string, timestamp time.Time) {
	if !b.isDataCapturingEnabled() {
		return
	}

	var builder strings.Builder

	b.buildBasicEventData(&builder, EXCEPTION, errorName)

	b.addKeyValuePair(&builder, BEACON_KEY_PARENT_ACTION_ID, parentActionID)
	b.addKeyValuePair(&builder, BEACON_KEY_START_SEQUENCE_NUMBER, b.CreateSequenceNumber())
	b.addKeyValuePair(&builder, BEACON_KEY_TIME_0, timestamp.Sub(b.sessionStartTime).Milliseconds())
	b.addKeyValuePairIfNotNull(&builder, BEACON_KEY_ERROR_VALUE, causeName)
	b.addKeyValuePairIfNotNull(&builder, BEACON_KEY_ERROR_REASON, causeDescription)
	b.addKeyValuePairIfNotNull(&builder, BEACON_KEY_ERROR_STACKTRACE, causeStackTrace)
	b.addKeyValuePair(&builder, BEACON_KEY_ERROR_TECHNOLOGY_TYPE, protocol.ERROR_TECHNOLOGY_TYPE)

	b.addEventData(timestamp, &builder)
}
func (b *Beacon) addWebRequest(parentActionID int, tracer interfaces.WebRequestTracer) {

	if !b.isErrorCapturingEnabled() {
		return
	}

	var builder strings.Builder

	b.buildBasicEventData(&builder, WEB_REQUEST, tracer.(*WebRequestTracer).url)

	b.addKeyValuePair(&builder, BEACON_KEY_PARENT_ACTION_ID, parentActionID)
	b.addKeyValuePair(&builder, BEACON_KEY_START_SEQUENCE_NUMBER, tracer.(*WebRequestTracer).startSequenceNo)
	b.addKeyValuePair(&builder, BEACON_KEY_TIME_0, tracer.(*WebRequestTracer).startTime.Sub(b.sessionStartTime).Milliseconds())
	b.addKeyValuePair(&builder, BEACON_KEY_END_SEQUENCE_NUMBER, tracer.(*WebRequestTracer).endSequenceNo)
	b.addKeyValuePair(&builder, BEACON_KEY_TIME_1, tracer.(*WebRequestTracer).endTime.Sub(tracer.(*WebRequestTracer).startTime).Milliseconds())

	b.addKeyValuePairIfNotNegative(&builder, BEACON_KEY_WEBREQUEST_BYTES_SENT, tracer.(*WebRequestTracer).bytesSent)
	b.addKeyValuePairIfNotNegative(&builder, BEACON_KEY_WEBREQUEST_BYTES_RECEIVED, tracer.(*WebRequestTracer).bytesReceived)
	b.addKeyValuePairIfNotNegative(&builder, BEACON_KEY_WEBREQUEST_RESPONSECODE, tracer.(*WebRequestTracer).responseCode)

	b.addEventData(tracer.(*WebRequestTracer).startTime, &builder)
}

func (b *Beacon) identifyUser(userTag string, timestamp time.Time) {
	if !b.isDataCapturingEnabled() {
		return
	}
	var builder strings.Builder

	if userTag != "" {
		b.buildBasicEventData(&builder, IDENTIFY_USER, userTag)
	} else {
		b.buildBasicEventDataWithoutName(&builder, IDENTIFY_USER)
	}

	b.addKeyValuePair(&builder, BEACON_KEY_PARENT_ACTION_ID, 0)
	b.addKeyValuePair(&builder, BEACON_KEY_START_SEQUENCE_NUMBER, b.CreateSequenceNumber())
	b.addKeyValuePair(&builder, BEACON_KEY_TIME_0, timestamp.Sub(b.sessionStartTime).Milliseconds())

	b.addEventData(timestamp, &builder)
}

func (b *Beacon) initializeServerConfiguration(c *configuration.ServerConfiguration) {
	b.configuration.InitializeServerConfiguration(c)
}
func (b *Beacon) enableCapture() {
	b.configuration.EnableCapture()
}

func (b *Beacon) reportCrash(name string, reason string, stacktrace string, timestamp time.Time) {

	if !b.isDataCapturingEnabled() {
		return
	}

	var builder strings.Builder

	b.buildBasicEventData(&builder, CRASH, name)

	b.addKeyValuePair(&builder, BEACON_KEY_PARENT_ACTION_ID, 0)
	b.addKeyValuePair(&builder, BEACON_KEY_START_SEQUENCE_NUMBER, b.CreateSequenceNumber())
	b.addKeyValuePair(&builder, BEACON_KEY_TIME_0, timestamp.Sub(b.sessionStartTime).Milliseconds())
	b.addKeyValuePairIfNotNull(&builder, BEACON_KEY_ERROR_REASON, reason)
	b.addKeyValuePairIfNotNull(&builder, BEACON_KEY_ERROR_STACKTRACE, stacktrace)
	b.addKeyValuePair(&builder, BEACON_KEY_ERROR_TECHNOLOGY_TYPE, protocol.ERROR_TECHNOLOGY_TYPE)

	b.addEventData(timestamp, &builder)
}

func truncate(name string) string {
	name = strings.TrimSpace(name)
	if len(name) > MAX_NAME_LEN {
		name = name[:MAX_NAME_LEN]
	}
	return name
}
