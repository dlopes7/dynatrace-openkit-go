package openkitgo

import (
	"github.com/op/go-logging"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

// TODO - Implement BeaconSender

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

type BeaconSenderContext struct {
	logger          logging.Logger
	httpClient      *HttpClient
	config          *Configuration
	isTerminalState bool

	currentState BeaconSendingState
	nextState    BeaconSendingState

	sessions map[int]*session
}

func (b BeaconSenderContext) removeSession(session *session) {

	delete(b.sessions, session.ID)

}

func (b BeaconSenderContext) isCapture() bool {
	return b.config.capture
}

func (BeaconSenderContext) sleep() {
	time.Sleep(1 * time.Second)
}

func (b *BeaconSenderContext) executeCurrentState() {
	b.nextState = nil

	b.currentState.execute(b)

	if b.nextState != nil && b.nextState != b.currentState {
		b.logger.Infof("executeCurrentState() - State change from %s to %s", b.currentState.ToString(), b.nextState.ToString())
		b.currentState = b.nextState
	}
}

func (b *BeaconSenderContext) startSession(session *session) {
	b.sessions[session.ID] = session
}

func (b *BeaconSenderContext) finishSession(session Session) {
	session.finishSession()
}

func (b *BeaconSenderContext) getAllNewSessions() []Session {
	newSessions := make([]Session, 0)

	for _, session := range b.sessions {

		if !session.isBeaconConfigurationSet() {
			newSessions = append(newSessions, session)
		}
	}

	return newSessions
}

func (b *BeaconSenderContext) handleStatusResponse(statusResponse *StatusResponse) {

	b.config.updateSettings(statusResponse)

}

func (b *BeaconSenderContext) getAllFinishedAndConfiguredSessions() []*session {

	finishedSessions := make([]*session, 0)

	for _, session := range b.sessions {

		if session.isBeaconConfigurationSet() && session.isSessionFinished() {
			finishedSessions = append(finishedSessions, session)
		}
	}

	return finishedSessions

}

type BeaconSendingState interface {
	execute(*BeaconSenderContext)
	isTerminalState() bool
	getShutdownState() BeaconSendingState
	ToString() string
}

// ------------------------------------------------------------
// beaconSendingCaptureOnState -> beaconSendingFlushSessionsState

type beaconSendingCaptureOnState struct{}

func (b *beaconSendingCaptureOnState) execute(context *BeaconSenderContext) {
	context.sleep()
	context.logger.Info("Executed state beaconSendingCaptureOnState")

	newSessionsResponse := b.sendNewSessionRequests(context)
	finishedSessionsResponse := b.sendFinishedSessions(context)

	lastStatusResponse := newSessionsResponse
	if finishedSessionsResponse != nil {
		lastStatusResponse = finishedSessionsResponse
	}

	b.handleStatusResponse(context, lastStatusResponse)

}

func (b *beaconSendingCaptureOnState) handleStatusResponse(context *BeaconSenderContext, statusResponse *StatusResponse) {
	if statusResponse == nil {
		return
	}
	context.handleStatusResponse(statusResponse)
}

func (b *beaconSendingCaptureOnState) sendNewSessionRequests(context *BeaconSenderContext) *StatusResponse {

	var statusResponse *StatusResponse

	for _, session := range context.getAllNewSessions() {

		if !session.canSendNewSessionRequest() {
			currentConfiguration := session.getBeaconConfiguration()
			newConfiguration := &BeaconConfiguration{
				multiplicity:        0,
				dataCollectionLevel: currentConfiguration.dataCollectionLevel,
				crashReportingLevel: currentConfiguration.crashReportingLevel,
			}
			session.updateBeaconConfiguration(newConfiguration)
			continue
		}

		statusResponse = context.httpClient.sendNewSessionRequest()

		if statusResponse.responseCode < 400 {
			currentConfiguration := session.getBeaconConfiguration()
			newConfiguration := &BeaconConfiguration{
				multiplicity:        statusResponse.multiplicity,
				dataCollectionLevel: currentConfiguration.dataCollectionLevel,
				crashReportingLevel: currentConfiguration.crashReportingLevel,
			}
			session.updateBeaconConfiguration(newConfiguration)

		}

	}

	return statusResponse

}

func (b *beaconSendingCaptureOnState) sendFinishedSessions(context *BeaconSenderContext) *StatusResponse {

	var statusResponse *StatusResponse

	for _, finishedSession := range context.getAllFinishedAndConfiguredSessions() {
		context.logger.Debug("Found finished session! Sending!")

		if finishedSession.isDataSendingAllowed() {
			context.removeSession(finishedSession)
			statusResponse = finishedSession.sendBeacon(context.httpClient)
			finishedSession.clearCapturedData()
			finishedSession.End()

		}
	}

	return statusResponse

}

func (b *beaconSendingCaptureOnState) getShutdownState() BeaconSendingState {
	return &beaconSendingFlushSessionsState{}
}

func (beaconSendingCaptureOnState) isTerminalState() bool {
	return false
}

func (beaconSendingCaptureOnState) ToString() string {
	return "beaconSendingCaptureOnState"
}

// ------------------------------------------------------------
// beaconSendingInitState -> beaconSendingCaptureOnState

type beaconSendingInitState struct{}

func (beaconSendingInitState) execute(context *BeaconSenderContext) {
	context.sleep()

	// TODO - Only set this if StatusRequest is OK
	context.nextState = &beaconSendingCaptureOnState{}

	context.logger.Info("Executed state beaconSendingInitState")

}

func (beaconSendingInitState) ToString() string {
	return "beaconSendingInitState"
}

func (beaconSendingInitState) isTerminalState() bool {
	return false
}

func (b *beaconSendingInitState) getShutdownState() BeaconSendingState {
	return &beaconSendingCaptureOnState{}
}

// ------------------------------------------------------------
// beaconSendingFlushSessionsState -> beaconSendingTerminalState

type beaconSendingFlushSessionsState struct{}

func (beaconSendingFlushSessionsState) execute(context *BeaconSenderContext) {

	// first get all sessions that do not have any multiplicity se
	for _, newSession := range context.getAllNewSessions() {
		currentConfiguration := newSession.getBeaconConfiguration()
		newSession.updateBeaconConfiguration(&BeaconConfiguration{
			multiplicity:        1,
			dataCollectionLevel: currentConfiguration.dataCollectionLevel,
			crashReportingLevel: currentConfiguration.crashReportingLevel,
		})

	}

	context.nextState = &beaconSendingTerminalState{}

}

func (beaconSendingFlushSessionsState) ToString() string {
	return "beaconSendingInitState"
}

func (beaconSendingFlushSessionsState) isTerminalState() bool {
	return false
}

func (b *beaconSendingFlushSessionsState) getShutdownState() BeaconSendingState {
	return &beaconSendingCaptureOnState{}
}

// ------------------------------------------------------------
// beaconSendingTerminalState ->

type beaconSendingTerminalState struct{}

func (beaconSendingTerminalState) execute(context *BeaconSenderContext) {
	context.sleep()

	// TODO - Only set this if StatusRequest is OK
	context.nextState = &beaconSendingCaptureOnState{}

	context.logger.Info("Executed state beaconSendingInitState")

}

func (beaconSendingTerminalState) ToString() string {
	return "beaconSendingInitState"
}

func (beaconSendingTerminalState) isTerminalState() bool {
	return false
}

func (b *beaconSendingTerminalState) getShutdownState() BeaconSendingState {
	return &beaconSendingCaptureOnState{}
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
	b.buildBasicEventData(eventBuilder, EventTypeSESSION_START, "")

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

	b.buildBasicEventData(eventBuilder, EventTypeSESSION_END, "")

	b.addKeyValuePair(&eventBuilder, BEACON_KEY_PARENT_ACTION_ID, strconv.Itoa(0))
	b.addKeyValuePair(&eventBuilder, BEACON_KEY_START_SEQUENCE_NUMBER, strconv.Itoa(b.createSequenceNumber()))
	b.addKeyValuePair(&eventBuilder, BEACON_KEY_TIME_0, strconv.Itoa(b.getTimeSinceSessionStartTime(session.endTime)))

	b.addEventData(session.endTime, &eventBuilder)

}

func (b *Beacon) addAction(action *action) {

	var actionBuilder strings.Builder

	b.buildBasicEventData(actionBuilder, EventTypeACTION, action.name)

	b.addKeyValuePair(&actionBuilder, BEACON_KEY_ACTION_ID, strconv.Itoa(action.ID))
	b.addKeyValuePair(&actionBuilder, BEACON_KEY_PARENT_ACTION_ID, strconv.Itoa(action.parentAction.ID))
	b.addKeyValuePair(&actionBuilder, BEACON_KEY_START_SEQUENCE_NUMBER, strconv.Itoa(action.startSequenceNo))
	b.addKeyValuePair(&actionBuilder, BEACON_KEY_TIME_0, strconv.Itoa(b.getTimeSinceSessionStartTime(action.startTime)))
	b.addKeyValuePair(&actionBuilder, BEACON_KEY_END_SEQUENCE_NUMBER, strconv.Itoa(action.endSequenceNo))
	b.addKeyValuePair(&actionBuilder, BEACON_KEY_TIME_1, strconv.Itoa(action.endTime-action.startTime))

	b.addActionData(action.startTime, actionBuilder)
}

func (b *Beacon) addActionData(timestamp int, sb strings.Builder) {
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

func (b *Beacon) buildBasicEventData(sb strings.Builder, eventType EventType, name string) {

	b.addKeyValuePair(&sb, BEACON_KEY_EVENT_TYPE, strconv.Itoa(int(eventType)))

	if len(name) != 0 {
		b.addKeyValuePair(&sb, BEACON_KEY_NAME, b.truncate(name))
	}

	// TODO - Replace "1" with getThreadID()
	b.addKeyValuePair(&sb, BEACON_KEY_THREAD_ID, "1")

}

func (b *Beacon) addEventData(timestamp int, sb *strings.Builder) {
	b.beaconCache.addEventData(b.sessionNumber, timestamp, sb.String())
}

func (b *Beacon) addKeyValuePair(sb *strings.Builder, key string, value string) {
	b.appendKey(sb, key)
	sb.WriteString(value)
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
