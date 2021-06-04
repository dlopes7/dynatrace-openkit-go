package communication

import (
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo/protocol"
)

type StateCaptureOn struct{}

//func (s *StateCaptureOn) execute(ctx *BeaconSendingContext) {
//	time.Sleep(DEFAULT_SLEEP_TIME_MILLISECONDS)
//
//	newSessionsResponse := s.sendNewSessionRequests(ctx)
//
//	/*
//	   if (BeaconSendingResponseUtil.isTooManyRequestsResponse(newSessionsResponse)) {
//	       // server is currently overloaded, temporarily switch to capture off
//	       context.setNextState(new BeaconSendingCaptureOffState(newSessionsResponse.getRetryAfterInMilliseconds()));
//	       return;
//	   }
//
//	   // send all finished sessions
//	   StatusResponse finishedSessionsResponse = sendFinishedSessions(context);
//	   if (BeaconSendingResponseUtil.isTooManyRequestsResponse(finishedSessionsResponse)) {
//	       // server is currently overloaded, temporarily switch to capture off
//	       context.setNextState(new BeaconSendingCaptureOffState(finishedSessionsResponse.getRetryAfterInMilliseconds()));
//	       return;
//	   }
//
//	   // check if we need to send open sessions & do it if necessary
//	   StatusResponse openSessionsResponse = sendOpenSessions(context);
//	   if (BeaconSendingResponseUtil.isTooManyRequestsResponse(openSessionsResponse)) {
//	       // server is currently overloaded, temporarily switch to capture off
//	       context.setNextState(new BeaconSendingCaptureOffState(openSessionsResponse.getRetryAfterInMilliseconds()));
//	       return;
//	   }
//
//	   // collect the last status response
//	   StatusResponse lastStatusResponse = newSessionsResponse;
//	   if (openSessionsResponse != null) {
//	       lastStatusResponse = openSessionsResponse;
//	   } else if (finishedSessionsResponse != null) {
//	       lastStatusResponse = finishedSessionsResponse;
//	   }
//
//	   // handle the last statusResponse received (or null if none was received) from the server
//	   handleStatusResponse(context, lastStatusResponse);
//	*/
//
//}

func (s *StateCaptureOn) terminal() bool {
	return false
}

func (s *StateCaptureOn) onInterrupted(ctx *BeaconSendingContext) {}

func (s *StateCaptureOn) getShutdownState() BeaconState {
	panic("Implement me")
}

func (s *StateCaptureOn) String() string {
	return "StateCaptureOn"
}

func (s *StateCaptureOn) sendNewSessionRequests(ctx *BeaconSendingContext) protocol.StatusResponse {
	panic("Implement me")
}
