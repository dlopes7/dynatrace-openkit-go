package openkitgo

import (
	"github.com/op/go-logging"
	"net/http"
	"sync"
	"time"
)

// TODO - Implement BeaconSender

var wg sync.WaitGroup

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
	wg.Add(1)
	go func() {
		b.logger.Info("BeaconSender goroutine initialized")
		for !b.context.isTerminalState {
			b.context.executeCurrentState()
		}
		wg.Done()
	}()
	wg.Wait()
}

type BeaconSenderContext struct {
	logger          logging.Logger
	httpClient      *http.Client
	config          *Configuration
	isTerminalState bool

	currentState BeaconSendingState
	nextState    BeaconSendingState
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

// TODO - Implement Beacon
// TODO - Implement BeaconCache
type Beacon struct {
}
