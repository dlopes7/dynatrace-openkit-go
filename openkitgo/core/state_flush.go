package core

import (
	"net/http"
)

type StateFlush struct{}

func (s *StateFlush) terminal() bool {
	return false
}

func (s *StateFlush) execute(ctx *BeaconSendingContext) {
	for _, session := range ctx.getAllNotConfiguredSessions() {
		session.enableCapture()
	}

	for _, session := range ctx.getAllOpenAndConfiguredSessions() {
		session.End()
	}

	tooManyRequestsReceived := false
	for _, session := range ctx.getAllFinishedAndConfiguredSessions() {
		if !tooManyRequestsReceived && session.isDataSendingAllowed() {
			resp := session.sendBeacon(ctx)
			if resp.ResponseCode == http.StatusTooManyRequests {
				tooManyRequestsReceived = true
			}
		}
		session.clearCapturedData()
		session.close()
		ctx.RemoveSession(session)
	}

	ctx.nextState = &StateTerminal{}

}

func (s *StateFlush) getShutdownState() BeaconState {
	return &StateTerminal{}
}

func (s *StateFlush) String() string {
	return "StateFlush"
}
