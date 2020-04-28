package openkitgo

import (
	log "github.com/sirupsen/logrus"
	"time"
)

type Session interface {
	EnterAction(string) Action
	EnterActionAt(string, time.Time) Action
	IdentifyUser(string)
	// reportCrash(string, string, string)
	TraceWebRequest(string) *WebRequestTracer
	TraceWebRequestAt(string, time.Time) *WebRequestTracer
	End()
	EndAt(time.Time)
	finishSession()

	isBeaconConfigurationSet() bool
	canSendNewSessionRequest() bool
	isSessionFinished() bool
	isDataSendingAllowed() bool

	getBeaconConfiguration() *BeaconConfiguration
	updateBeaconConfiguration(*BeaconConfiguration)

	sendBeacon(*HttpClient) *StatusResponse

	clearCapturedData()
}

type session struct {
	ID      int
	endTime int

	beaconSender *BeaconSender
	beacon       *Beacon
	log          *log.Logger

	openRootActions map[int]Action

	sessionFinished           bool
	beaconConfigurationSet    bool
	numNewSessionRequestsLeft int

	position      int
	sessionNumber int
}

func newSession(log *log.Logger, beaconSender *BeaconSender, beacon *Beacon) Session {
	s := new(session)

	s.log = log
	s.beaconSender = beaconSender
	s.beacon = beacon
	s.ID = s.beacon.config.createSessionNumber()
	s.openRootActions = make(map[int]Action)

	s.numNewSessionRequestsLeft = 4
	beaconSender.startSession(s)
	beacon.startSession()

	return s
}

func (s *session) clearCapturedData() {
	s.beacon.beaconCache.deleteCacheEntry(s.beacon.sessionNumber)
}

func (s *session) EnterAction(actionName string) Action {
	s.log.Debugf("Session.EnterAction(%s)", actionName)

	return newRootAction(s.log, s.beacon, actionName, s.openRootActions)

}

func (s *session) EnterActionAt(actionName string, timestamp time.Time) Action {
	s.log.Debugf("Session.EnterActionAt(%s, %s)", actionName, timestamp.String())

	return newRootActionAt(s.log, s.beacon, actionName, s.openRootActions, timestamp)

}

func (s *session) finishSession() {
	s.sessionFinished = true
}

func (s *session) isBeaconConfigurationSet() bool {
	return s.beaconConfigurationSet
}

func (s *session) getBeaconConfiguration() *BeaconConfiguration {
	return &s.beacon.beaconConfiguration
}

func (s *session) updateBeaconConfiguration(beaconConfiguration *BeaconConfiguration) {
	s.beacon.beaconConfiguration = *beaconConfiguration
	s.beaconConfigurationSet = true
}

func (s *session) canSendNewSessionRequest() bool {
	return s.numNewSessionRequestsLeft > 0
}

func (s *session) isSessionFinished() bool {
	return s.sessionFinished
}

func (s *session) isDataSendingAllowed() bool {
	return s.isBeaconConfigurationSet() && s.beacon.beaconConfiguration.multiplicity > 0
}

func (s *session) sendBeacon(httpClient *HttpClient) *StatusResponse {
	return s.beacon.send(httpClient)
}

func (s *session) IdentifyUser(userTag string) {
	s.log.Debugf("Session.IdentifyUser(%s)", userTag)
	s.beacon.identifyUser(userTag)
}

func (s *session) End() {
	s.log.Debugf("Session.End()")

	s.endTime = s.beacon.getCurrentTimestamp()

	for len(s.openRootActions) != 0 {
		for _, a := range s.openRootActions {
			a.LeaveAction()
		}
	}

	s.beacon.endSession(s)
	s.beaconSender.finishSession(s)
}

func (s *session) EndAt(endTime time.Time) {
	s.log.Debugf("Session.EndAt(%s)", endTime)

	s.endTime = TimeToMillis(endTime)

	for len(s.openRootActions) != 0 {
		for _, a := range s.openRootActions {
			a.LeaveActionAt(endTime)
		}
	}

	s.beacon.endSession(s)
	s.beaconSender.finishSession(s)
}

func (s *session) getActionID() int {
	return 0
}

func (s *session) TraceWebRequest(url string) *WebRequestTracer {
	// TODO : Store in children
	w := NewWebRequestTracer(s.log, s, url, s.beacon)
	return w
}

func (s *session) TraceWebRequestAt(url string, timestamp time.Time) *WebRequestTracer {
	// TODO : Store in children
	w := NewWebRequestTracerAt(s.log, s, url, s.beacon, timestamp)
	return w
}
