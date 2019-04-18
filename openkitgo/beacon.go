package openkitgo

import (
	"github.com/op/go-logging"
	"net/http"
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

func NewBeaconSender(logger logging.Logger, config *Configuration, client *http.Client) *BeaconSender {
	b := new(BeaconSender)
	b.logger = logger

	b.context = BeaconSenderContext{
		logger:       logger,
		config:       config,
		httpClient:   client,
		currentState: new(beaconSendingInitState),
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

func (b *BeaconSender) startSession(session Session) {
	b.logger.Debug("BeaconSender startSession()")

	b.context.startSession(session)

}

func (b *BeaconSender) finishSession(session Session) {
	b.logger.Debug("BeaconSender finishSession()")

	b.context.finishSession(session)

}

type BeaconSenderContext struct {
	logger          logging.Logger
	httpClient      *http.Client
	config          *Configuration
	isTerminalState bool

	currentState BeaconSendingState
	nextState    BeaconSendingState

	sessions []Session
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

func (b *BeaconSenderContext) startSession(session Session) {
	b.sessions = append(b.sessions, session)
}

func (b *BeaconSenderContext) finishSession(session Session) {
	session.finishSession()
}

type BeaconSendingState interface {
	execute(*BeaconSenderContext)
	isTerminalState() bool
	ToString() string
}

type beaconSendingCaptureOnState struct{}

func (beaconSendingCaptureOnState) execute(context *BeaconSenderContext) {
	context.sleep()
	context.logger.Info("Executed state beaconSendingCaptureOnState")
}

func (beaconSendingCaptureOnState) isTerminalState() bool {
	return false
}

func (beaconSendingCaptureOnState) ToString() string {
	return "beaconSendingCaptureOnState"
}

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

const (
	BEACON_KEY_EVENT_TYPE            = "et"
	BEACON_KEY_NAME                  = "na"
	BEACON_KEY_THREAD_ID             = "it"
	BEACON_KEY_PARENT_ACTION_ID      = "pa"
	BEACON_KEY_START_SEQUENCE_NUMBER = "s0"
	BEACON_KEY_TIME_0                = "t0"
	BEACON_KEY_ACTION_ID             = "ca"
	BEACON_KEY_END_SEQUENCE_NUMBER   = "s1"
	BEACON_KEY_TIME_1                = "t1"

	MAX_NAME_LEN = 250
)

type Beacon struct {
	logger          logging.Logger
	beaconCache     *beaconCache
	config          *Configuration
	clientIPAddress string
	ops             uint64

	sessionNumber    int
	sessionStartTime int
}

func NewBeacon(logger logging.Logger, beaconCache *beaconCache, config *Configuration, clientIPAddress string) *Beacon {
	b := new(Beacon)

	b.sessionNumber = b.config.createSessionNumber()
	b.sessionStartTime = b.config.makeTimestamp()
	b.logger = logger
	b.beaconCache = beaconCache
	b.config = config
	b.clientIPAddress = clientIPAddress

	return b

}

func (b *Beacon) startSession() {

	var eventBuilder strings.Builder
	b.buildBasicEventData(eventBuilder, EventTypeSESSION_START, "")

	b.addKeyValuePair(eventBuilder, BEACON_KEY_PARENT_ACTION_ID, "0")
	b.addKeyValuePair(eventBuilder, BEACON_KEY_START_SEQUENCE_NUMBER, strconv.Itoa(b.createSequenceNumber()))
	b.addKeyValuePair(eventBuilder, BEACON_KEY_TIME_0, "0")

	b.addEventData(b.sessionStartTime, eventBuilder)

}

func (b *Beacon) getCurrentTimestamp() int {
	return b.config.makeTimestamp()

}

func (b *Beacon) endSession(session *session) {

	var eventBuilder strings.Builder

	b.buildBasicEventData(eventBuilder, EventTypeSESSION_END, "")

	b.addKeyValuePair(eventBuilder, BEACON_KEY_PARENT_ACTION_ID, strconv.Itoa(0))
	b.addKeyValuePair(eventBuilder, BEACON_KEY_START_SEQUENCE_NUMBER, strconv.Itoa(b.createSequenceNumber()))
	b.addKeyValuePair(eventBuilder, BEACON_KEY_TIME_0, strconv.Itoa(b.getTimeSinceSessionStartTime(session.endTime)))

	b.addEventData(session.endTime, eventBuilder)

}

func (b *Beacon) addAction(action *action) {

	var actionBuilder strings.Builder

	b.buildBasicEventData(actionBuilder, EventTypeACTION, action.name)

	b.addKeyValuePair(actionBuilder, BEACON_KEY_ACTION_ID, strconv.Itoa(action.ID))
	b.addKeyValuePair(actionBuilder, BEACON_KEY_PARENT_ACTION_ID, strconv.Itoa(action.parentAction.ID))
	b.addKeyValuePair(actionBuilder, BEACON_KEY_START_SEQUENCE_NUMBER, strconv.Itoa(action.startSequenceNo))
	b.addKeyValuePair(actionBuilder, BEACON_KEY_TIME_0, strconv.Itoa(b.getTimeSinceSessionStartTime(action.startTime)))
	b.addKeyValuePair(actionBuilder, BEACON_KEY_END_SEQUENCE_NUMBER, strconv.Itoa(action.endSequenceNo))
	b.addKeyValuePair(actionBuilder, BEACON_KEY_TIME_1, strconv.Itoa(action.endTime-action.startTime))

	b.addActionData(action.startTime, actionBuilder)
}

func (b *Beacon) addActionData(timestamp int, sb strings.Builder) {
	b.beaconCache.addActionData(b.sessionNumber, timestamp, sb.String())
}

/*
   public void addAction(ActionImpl action) {

       if (isCapturingDisabled()) {
           return;
       }

       if (getBeaconConfiguration().getDataCollectionLevel() == DataCollectionLevel.OFF) {
           return;
       }

       StringBuilder actionBuilder = new StringBuilder();

       buildBasicEventData(actionBuilder, EventType.ACTION, action.getName());

       addKeyValuePair(actionBuilder, BEACON_KEY_ACTION_ID, action.getID());
       addKeyValuePair(actionBuilder, BEACON_KEY_PARENT_ACTION_ID, action.getParentID());
       addKeyValuePair(actionBuilder, BEACON_KEY_START_SEQUENCE_NUMBER, action.getStartSequenceNo());
       addKeyValuePair(actionBuilder, BEACON_KEY_TIME_0, getTimeSinceSessionStartTime(action.getStartTime()));
       addKeyValuePair(actionBuilder, BEACON_KEY_END_SEQUENCE_NUMBER, action.getEndSequenceNo());
       addKeyValuePair(actionBuilder, BEACON_KEY_TIME_1, action.getEndTime() - action.getStartTime());

       addActionData(action.getStartTime(), actionBuilder);
   }
*/

func (b *Beacon) getTimeSinceSessionStartTime(timestamp int) int {
	return timestamp - b.sessionStartTime
}

func (b *Beacon) createSequenceNumber() int {
	atomic.AddUint64(&b.ops, 1)
	return int(b.ops)
}

func (b *Beacon) buildBasicEventData(sb strings.Builder, eventType EventType, name string) {

	b.addKeyValuePair(sb, BEACON_KEY_EVENT_TYPE, strconv.Itoa(int(eventType)))

	if len(name) != 0 {
		b.addKeyValuePair(sb, BEACON_KEY_NAME, b.truncate(name))
	}

	// TODO - Replace "1" with getThreadID()
	b.addKeyValuePair(sb, BEACON_KEY_THREAD_ID, "1")

}

func (b *Beacon) addEventData(timestamp int, sb strings.Builder) {
	b.beaconCache.addEventData(b.sessionNumber, timestamp, sb.String())
}

func (b *Beacon) addKeyValuePair(sb strings.Builder, key string, value string) {
	b.appendKey(sb, key)
	sb.WriteString(value)
}

func (b *Beacon) appendKey(sb strings.Builder, key string) {
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
