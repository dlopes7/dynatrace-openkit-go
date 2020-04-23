package openkitgo

import (
	log "github.com/sirupsen/logrus"
	"time"
)

type BeaconSenderContext struct {
	log        *log.Logger
	httpClient *HttpClient
	config     *Configuration

	isTerminalState bool
	initCompleted   bool

	currentState BeaconSendingState
	nextState    BeaconSendingState

	lastOpenSessionBeaconSendTime time.Time
	lastStatusCheckTime           time.Time

	shutdown bool

	sessions map[int]*session
}

func (b BeaconSenderContext) removeSession(session *session) {

	delete(b.sessions, session.ID)

}

func (b BeaconSenderContext) isCapture() bool {
	return b.config.capture
}

func (BeaconSenderContext) sleep(timeToSleep time.Duration) {
	time.Sleep(timeToSleep)
}

func (b *BeaconSenderContext) executeCurrentState() {
	b.nextState = nil

	b.currentState.execute(b)

	if b.nextState != nil && b.nextState != b.currentState {
		b.log.Debugf("executeCurrentState() - State change from %s to %s", b.currentState.String(), b.nextState.String())
		b.currentState = b.nextState
	}
}

func (b *BeaconSenderContext) startSession(session *session) {
	b.sessions[session.ID] = session
}

func (b *BeaconSenderContext) finishSession(session Session) {
	session.finishSession()
}

func (b *BeaconSenderContext) getAllNewSessions() []Session {
	newSessions := make([]Session, 0)

	for _, session := range b.sessions {

		if !session.isBeaconConfigurationSet() {
			newSessions = append(newSessions, session)
		}
	}

	return newSessions
}

func (b *BeaconSenderContext) handleStatusResponse(statusResponse *StatusResponse) {

	b.config.updateSettings(statusResponse)

}

func (b *BeaconSenderContext) getAllFinishedAndConfiguredSessions() []*session {

	finishedSessions := make([]*session, 0)

	for _, session := range b.sessions {

		if session.isBeaconConfigurationSet() && session.isSessionFinished() {
			finishedSessions = append(finishedSessions, session)
		}
	}

	return finishedSessions

}

func (b *BeaconSenderContext) getAllOpenAndConfiguredSessions() []*session {

	openSessions := make([]*session, 0)

	for _, session := range b.sessions {

		if session.isBeaconConfigurationSet() && !session.isSessionFinished() {
			openSessions = append(openSessions, session)
		}
	}

	return openSessions

}
