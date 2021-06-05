package core

import (
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo/protocol"
	"math"
	"net/http"
	"time"
)

const (
	MAX_INITIAL_STATUS_REQUEST_RETRIES    = 5
	INITIAL_RETRY_SLEEP_TIME_MILLISECONDS = 1 * time.Second
)

type StateInit struct {
	reInitDelayMilliseconds []time.Duration
	reInitDelayIndex        int
}

func NewStateInit() *StateInit {
	r := []time.Duration{
		1 * time.Minute,
		5 * time.Minute,
		15 * time.Minute,
		1 * time.Hour,
		2 * time.Hour,
	}
	return &StateInit{
		reInitDelayMilliseconds: r,
	}
}

func (s *StateInit) execute(ctx *BeaconSendingContext) {

	statusResponse := s.executeStatusRequest(ctx)

	if ctx.IsShutdownRequested() {
		ctx.initWg.Done()
		ctx.initOk = false
		ctx.nextState = s.getShutdownState()
	} else if statusResponse.ResponseCode < http.StatusBadRequest {
		ctx.handleStatusResponse(statusResponse)

		if ctx.isCaptureOn() {
			ctx.nextState = NewStateCaptureOn()
		} else {
			ctx.nextState = NewStateCaptureOff(0)
		}
		ctx.initWg.Done()
		ctx.initOk = true
	}

}

func (s *StateInit) executeStatusRequest(ctx *BeaconSendingContext) protocol.StatusResponse {

	var statusResponse protocol.StatusResponse

	for {

		currentTimestamp := ctx.getCurrentTimestamp()
		ctx.lastOpenSessionSent = currentTimestamp
		ctx.lastStatusCheck = currentTimestamp

		statusResponse = sendStatusRequest(ctx, MAX_INITIAL_STATUS_REQUEST_RETRIES, INITIAL_RETRY_SLEEP_TIME_MILLISECONDS)
		if ctx.IsShutdownRequested() || statusResponse.ResponseCode < http.StatusBadRequest {
			// We are done, we are either shutting down or we got a good response
			break
		}

		sleepTime := s.reInitDelayMilliseconds[s.reInitDelayIndex]

		if statusResponse.ResponseCode == 429 {
			sleepTime = statusResponse.GetRetryAfter()
			ctx.disableCaptureAndClear()
		}
		time.Sleep(sleepTime)
		s.reInitDelayIndex = int(math.Min(float64(s.reInitDelayIndex+1), float64(len(s.reInitDelayMilliseconds)-1)))
	}

	return statusResponse

}

func (s *StateInit) terminal() bool {
	return false
}

func (s *StateInit) getShutdownState() BeaconState {
	return &StateTerminal{}
}

func (*StateInit) String() string {
	return "StateInit"
}
