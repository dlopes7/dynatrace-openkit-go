package core

import (
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo/protocol"
	"net/http"
	"time"
)

const (
	STATUS_REQUEST_RETRIES = 5
	STATUS_CHECK_INTERVAL  = 2 * time.Hour
)

type StateCaptureOff struct {
	sleepTime time.Duration
}

func NewStateCaptureOff(sleepTime time.Duration) *StateCaptureOff {
	return &StateCaptureOff{
		sleepTime: sleepTime,
	}
}

func (s StateCaptureOff) terminal() bool {
	return false
}

func (s StateCaptureOff) execute(ctx *BeaconSendingContext) {
	ctx.disableCaptureAndClear()

	currentTime := time.Now()

	var delta time.Duration
	if s.sleepTime > 0 {
		delta = s.sleepTime
	} else {
		delta = STATUS_CHECK_INTERVAL - (currentTime.Sub(ctx.lastStatusCheck))
	}
	if delta > 0 && !ctx.IsShutdownRequested() {
		time.Sleep(delta)
	}

	statusResponse := sendStatusRequest(ctx, STATUS_REQUEST_RETRIES, INITIAL_RETRY_SLEEP_TIME_MILLISECONDS)
	s.handleStatusResponse(ctx, statusResponse)
	ctx.lastStatusCheck = currentTime
}

func (s StateCaptureOff) getShutdownState() BeaconState {
	// TODO BeaconSendingFlushSessionsState
	panic("implement me")
}

func (s StateCaptureOff) handleStatusResponse(ctx *BeaconSendingContext, response protocol.StatusResponse) {
	ctx.handleStatusResponse(response)

	if response.ResponseCode == http.StatusTooManyRequests {
		ctx.nextState = NewStateCaptureOff(response.GetRetryAfter())
	} else if response.ResponseCode < http.StatusBadRequest && ctx.isCaptureOn() {
		ctx.nextState = NewStateCaptureOn()
	}
}

func (s StateCaptureOff) String() string {
	return "StateCaptureOff"
}
