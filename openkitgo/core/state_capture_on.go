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

	// send new session request for all sessions that are new
	newSessionsResponse := s.sendNewSessionRequests(ctx)
	if newSessionsResponse.ResponseCode == http.StatusTooManyRequests {
		ctx.nextState = NewStateCaptureOff(newSessionsResponse.GetRetryAfter())
		return
	}

	// send all finished sessions
	finishedSessionsResponse := s.sendFinishedSessions(ctx)
	if finishedSessionsResponse.ResponseCode == http.StatusTooManyRequests {
		ctx.nextState = NewStateCaptureOff(finishedSessionsResponse.GetRetryAfter())
		return
	}

	// check if we need to send open sessions & do it if necessary
	openSessionsResponse := s.sendOpenSessions(ctx)
	if openSessionsResponse.ResponseCode == http.StatusTooManyRequests {
		ctx.nextState = NewStateCaptureOff(openSessionsResponse.GetRetryAfter())
		return
	}

	lastStatusResponse := newSessionsResponse
	if openSessionsResponse.ResponseCode != 0 {
		lastStatusResponse = openSessionsResponse
	} else if finishedSessionsResponse.ResponseCode != 0 {
		lastStatusResponse = finishedSessionsResponse
	}

	s.handleStatusResponse(ctx, lastStatusResponse)
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

func (s *StateCaptureOn) sendFinishedSessions(ctx *BeaconSendingContext) protocol.StatusResponse {

	statusResponse := protocol.StatusResponse{}

	for _, session := range ctx.getAllFinishedAndConfiguredSessions() {
		if session.isDataSendingAllowed() {
			statusResponse = session.sendBeacon(ctx)
			if statusResponse.ResponseCode >= http.StatusBadRequest {
				if statusResponse.ResponseCode == http.StatusTooManyRequests || session.isEmpty() {
					break
				}
			}
		}
		ctx.RemoveSession(session)
		session.clearCapturedData()
		session.close()
	}

	return statusResponse
}

func (s *StateCaptureOn) sendOpenSessions(ctx *BeaconSendingContext) protocol.StatusResponse {
	statusResponse := protocol.StatusResponse{}

	currentTime := time.Now()
	if currentTime.Before(ctx.lastOpenSessionSent.Add(ctx.GetSendInterval())) {
		return statusResponse
	}

	for _, session := range ctx.getAllOpenAndConfiguredSessions() {
		if session.isDataSendingAllowed() {
			statusResponse = session.sendBeacon(ctx)
			if statusResponse.ResponseCode == http.StatusTooManyRequests {
				break
			}
		} else {
			session.clearCapturedData()
		}
	}

	ctx.lastOpenSessionSent = currentTime
	return statusResponse
}

func (s *StateCaptureOn) handleStatusResponse(ctx *BeaconSendingContext, response protocol.StatusResponse) {
	if response.ResponseCode == 0 {
		return
	}

	ctx.handleStatusResponse(response)
	if !ctx.isCaptureOn() {
		ctx.nextState = NewStateCaptureOff(0)
	}
}
