package communication

import (
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo/configuration"
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo/core"
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo/protocol"
	log "github.com/sirupsen/logrus"
	"sync"
	"sync/atomic"
	"time"
)

const (
	DEFAULT_SLEEP_TIME_MILLISECONDS = 1 * time.Second
)

type BeaconSendingContext struct {
	log                     *log.Logger
	mutex                   sync.Mutex
	serverConfiguration     configuration.ServerConfiguration
	lastResponseAttributes  protocol.ResponseAttributes
	httpClientConfiguration configuration.HttpClientConfiguration
	sessions                []*core.Session

	shutdown int32 // atomic
	initWg   sync.WaitGroup

	currentState BeaconState
	nextState    BeaconState

	lastOpenSessionSent time.Time
	lastStatusCheck     time.Time
	initOk              bool
}

func NewBeaconSendingContext(log *log.Logger,
	httpClientConfiguration configuration.HttpClientConfiguration) *BeaconSendingContext {

	return &BeaconSendingContext{
		log:                     log,
		serverConfiguration:     configuration.DefaultServerConfiguration(),
		lastResponseAttributes:  protocol.UndefinedResponseAttributes(),
		httpClientConfiguration: httpClientConfiguration,
		initWg:                  sync.WaitGroup{},
		currentState:            NewStateInit(),
	}

}

func (c *BeaconSendingContext) executeCurrentState() {
	c.nextState = nil
	c.currentState.execute(c)

	if c.nextState != nil && c.nextState != c.currentState {
		c.log.WithFields(log.Fields{"currentState": c.currentState, "nextState": c.nextState}).Debug("changing state")
	}
	c.currentState = c.nextState
}

func (c *BeaconSendingContext) getCurrentTimestamp() time.Time {
	return time.Now()
}

func (c *BeaconSendingContext) getHttpClient() HttpClient {
	return NewHttpClient(c.log, c.httpClientConfiguration)
}

func (c *BeaconSendingContext) GetConfigurationTimestamp() time.Time {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.lastResponseAttributes.Timestamp
}

func (c *BeaconSendingContext) isCaptureOn() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.serverConfiguration.Capture
}

func (c *BeaconSendingContext) disableCapture() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.serverConfiguration.Capture = false
}

func (c *BeaconSendingContext) IsShutdownRequested() bool {
	return atomic.LoadInt32(&c.shutdown) == 1
}

func (c *BeaconSendingContext) handleStatusResponse(statusResponse protocol.StatusResponse) {
	if statusResponse.ResponseCode >= 400 {
		c.disableCapture()
		// TODO clearAllSessionData
		return
	}

	c.updateFrom(statusResponse)

	if !c.isCaptureOn() {
		// TODO clearAllSessionData
	}

}

func (c *BeaconSendingContext) updateFrom(statusResponse protocol.StatusResponse) protocol.ResponseAttributes {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if statusResponse.ResponseCode >= 400 {
		return c.lastResponseAttributes
	}

	c.lastResponseAttributes = c.lastResponseAttributes.Merge(statusResponse.ResponseAttributes)
	c.serverConfiguration = configuration.NewServerConfiguration(c.lastResponseAttributes)
	c.httpClientConfiguration.ServerID = c.serverConfiguration.ServerID

	return c.lastResponseAttributes
}

func (c *BeaconSendingContext) requestShutDown() {
	atomic.StoreInt32(&c.shutdown, 0)
}

func (c *BeaconSendingContext) WaitForInitTimeout(timeout time.Duration) bool {
	if waitTimeout(&c.initWg, timeout) {
		c.log.WithFields(log.Fields{"timeout": timeout}).Error("timed out waiting for init")
		return c.initOk
	}
	return c.initOk
}

func (c *BeaconSendingContext) WaitForInit() bool {
	c.initWg.Wait()
	return c.initOk
}

func waitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		return false
	case <-time.After(timeout):
		return true
	}
}

func (c *BeaconSendingContext) IsInitialized() bool {
	return c.initOk
}

func (c *BeaconSendingContext) IsInTerminalState() bool {
	return c.currentState.terminal()
}

func (c *BeaconSendingContext) GetSendInterval() time.Duration {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.serverConfiguration.SendInterval
}

func (c *BeaconSendingContext) GetLastServerConfiguration() configuration.ServerConfiguration {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.serverConfiguration
}

func (c *BeaconSendingContext) disableCaptureAndClear() {
	c.disableCapture()
	c.clearAllSessionData()
}

func (c *BeaconSendingContext) clearAllSessionData() {

	var keepSessions []*core.Session

	c.mutex.Lock()
	defer c.mutex.Unlock()
	for _, session := range c.sessions {
		session.ClearCapturedData()
		if !session.State.IsFinished() {
			keepSessions = append(keepSessions, session)
		}
	}
	c.sessions = keepSessions
}

func (c *BeaconSendingContext) getAllNotConfiguredSessions() []*core.Session {

	var filtered []*core.Session

	for _, session := range c.sessions {
		if !session.State.IsConfigured() {
			filtered = append(filtered, session)
		}
	}
	return filtered
}

func (c *BeaconSendingContext) getAllOpenAndConfiguredSessions() []*core.Session {

	var filtered []*core.Session

	for _, session := range c.sessions {
		if !session.State.IsConfiguredAndOpen() {
			filtered = append(filtered, session)
		}
	}
	return filtered
}

func (c *BeaconSendingContext) getAllFinishedAndConfiguredSessions() []*core.Session {

	var filtered []*core.Session

	for _, session := range c.sessions {
		if !session.State.IsConfiguredAndFinished() {
			filtered = append(filtered, session)
		}
	}
	return filtered
}

func (c *BeaconSendingContext) GetCurrentServerId() int {
	return c.httpClientConfiguration.ServerID
}

func (c *BeaconSendingContext) AddSession(session *core.Session) {
	c.sessions = append(c.sessions, session)
}

func (c *BeaconSendingContext) RemoveSession(session *core.Session) {
	var keep []*core.Session

	for _, s := range c.sessions {
		if s != session {
			keep = append(keep, s)
		}
	}
	c.sessions = keep
}
