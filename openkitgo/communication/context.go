package communication

import (
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo"
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo/configuration"
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
	httpClientConfiguration configuration.HttpClient
	sessions                chan openkitgo.Session

	shutdown int32 // atomic
	initWg   sync.WaitGroup

	currentState BeaconState
	nextState    BeaconState

	lastOpenSessionSent time.Time
	lastStatusCheck     time.Time
	initOk              bool
}

func NewBeaconSendingContext(log *log.Logger,
	httpClientConfiguration configuration.HttpClient) *BeaconSendingContext {

	return &BeaconSendingContext{
		log:                     log,
		serverConfiguration:     configuration.DefaultServerConfiguration(),
		lastResponseAttributes:  protocol.UndefinedResponseAttributes(),
		httpClientConfiguration: httpClientConfiguration,
		initWg:                  sync.WaitGroup{},
		currentState:            NewStateInit(),
	}

}

func (b *BeaconSendingContext) executeCurrentState() {
	b.nextState = nil
	b.currentState.execute(b)

	if b.nextState != nil && b.nextState != b.currentState {
		b.log.WithFields(log.Fields{"currentState": b.currentState, "nextState": b.nextState}).Debug("changing state")
	}
	b.currentState = b.nextState
}

func (b *BeaconSendingContext) getCurrentTimestamp() time.Time {
	return time.Now()
}

func (b *BeaconSendingContext) getHttpClient() HttpClient {
	return NewHttpClient(b.log, b.httpClientConfiguration)
}

func (b *BeaconSendingContext) GetConfigurationTimestamp() time.Time {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	return b.lastResponseAttributes.Timestamp
}

func (b *BeaconSendingContext) isCaptureOn() bool {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	return b.serverConfiguration.Capture
}

func (b *BeaconSendingContext) disableCapture() {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.serverConfiguration.Capture = false
}

func (b *BeaconSendingContext) IsShutdownRequested() bool {
	return atomic.LoadInt32(&b.shutdown) == 1
}

func (b *BeaconSendingContext) handleStatusResponse(statusResponse protocol.StatusResponse) {
	if statusResponse.ResponseCode >= 400 {
		b.disableCapture()
		// TODO clearAllSessionData
		return
	}

	b.updateFrom(statusResponse)

	if !b.isCaptureOn() {
		// TODO clearAllSessionData
	}

}

func (b *BeaconSendingContext) updateFrom(statusResponse protocol.StatusResponse) protocol.ResponseAttributes {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	if statusResponse.ResponseCode >= 400 {
		return b.lastResponseAttributes
	}

	b.lastResponseAttributes = b.lastResponseAttributes.Merge(statusResponse.ResponseAttributes)
	b.serverConfiguration = configuration.NewServerConfiguration(b.lastResponseAttributes)
	b.httpClientConfiguration.ServerID = b.serverConfiguration.ServerID

	return b.lastResponseAttributes
}

func (b *BeaconSendingContext) requestShutDown() {
	atomic.StoreInt32(&b.shutdown, 0)
}

func (b *BeaconSendingContext) WaitForInitTimeout(timeout time.Duration) bool {
	if waitTimeout(&b.initWg, timeout) {
		b.log.WithFields(log.Fields{"timeout": timeout}).Error("timed out waiting for init")
		return b.initOk
	}
	return b.initOk
}

func (b *BeaconSendingContext) WaitForInit() bool {
	b.initWg.Wait()
	return b.initOk
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
