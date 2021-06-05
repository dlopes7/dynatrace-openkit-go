package core

import (
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo/configuration"
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo/protocol"
	"net/http"
	"time"
)

type StateCaptureOn struct{}

func NewStateCaptureOn() *StateCaptureOn {
	return &StateCaptureOn{}

}

func (s *StateCaptureOn) execute(ctx *BeaconSendingContext) {
	time.Sleep(DEFAULT_SLEEP_TIME_MILLISECONDS)

	newSessionsResponse := s.sendNewSessionRequests(ctx)
	if newSessionsResponse.ResponseCode == http.StatusTooManyRequests {
		ctx.nextState = NewStateCaptureOff()
		return
	}

	// TODO - Implement the rest

}

func (s *StateCaptureOn) terminal() bool {
	return false
}

func (s *StateCaptureOn) getShutdownState() BeaconState {
	panic("Implement me")
	// TODO Flush
}

func (s *StateCaptureOn) String() string {
	return "StateCaptureOn"
}

func (s *StateCaptureOn) sendNewSessionRequests(ctx *BeaconSendingContext) protocol.StatusResponse {

	var statusResponse protocol.StatusResponse

	httpClient := ctx.getHttpClient()
	for _, session := range ctx.getAllNotConfiguredSessions() {
		if !session.canSendNewSessionRequest() {
			session.disableCapture()
			continue
		}

		statusResponse = httpClient.SendNewSessionRequest(ctx)
		if statusResponse.ResponseCode < http.StatusBadRequest {
			updatedAttributes := ctx.updateFrom(statusResponse)
			newServerConfig := configuration.NewServerConfiguration(updatedAttributes)
			session.updateServerConfiguration(newServerConfig)
		} else if statusResponse.ResponseCode == http.StatusTooManyRequests {
			break
		} else {
			session.remainingRequests--
		}
	}

	return statusResponse
}
