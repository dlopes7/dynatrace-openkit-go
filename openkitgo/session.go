package openkitgo

import (
	log "github.com/sirupsen/logrus"
	"sync"
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

const MAX_NEW_SESSION_REQUESTS = 4

type sessionState struct {
	session        *session
	finishing      bool
	finished       bool
	triedForEnding bool
	lock           sync.Mutex
}

func newSessionState(s *session) *sessionState {
	st := new(sessionState)
	st.session = s
	return st
}

func (st *sessionState) wasTriedForEnding() bool {
	st.lock.Lock()
	defer st.lock.Unlock()
	return st.triedForEnding
}

func (st *sessionState) isConfigured() bool {
	st.lock.Lock()
	defer st.lock.Unlock()
	return st.session.beacon.config.isServerConfigurationSet()
}

func (st *sessionState) isConfiguredAndFinished() bool {
	st.lock.Lock()
	defer st.lock.Unlock()
	return st.isConfigured() && st.finished
}

func (st *sessionState) isConfiguredAndOpen() bool {
	st.lock.Lock()
	defer st.lock.Unlock()
	return st.isConfigured() && !st.finished
}

func (st *sessionState) isFinished() bool {
	st.lock.Lock()
	defer st.lock.Unlock()
	return st.finished
}

func (st *sessionState) isFinishingOrFinished() bool {
	st.lock.Lock()
	defer st.lock.Unlock()
	return st.finishing || st.finished
}

func (st *sessionState) markAsIsFinishing() bool {

	if st.isFinishingOrFinished() {
		return false
	}
	st.finishing = true
	return true
}

func (st *sessionState) markAsFinished() {
	st.lock.Lock()
	defer st.lock.Unlock()
	st.finished = true
}

func (st *sessionState) markAsWasTriedForEnding() {
	st.lock.Lock()
	defer st.lock.Unlock()
	st.triedForEnding = true
}

type session struct {
	ID      int
	endTime time.Time

	beaconSender *BeaconSender
	beacon       *Beacon
	log          *log.Logger
	children     []OpenKitObject

	openRootActions map[int]Action

	sessionFinished           bool
	beaconConfigurationSet    bool
	numNewSessionRequestsLeft int

	position      int
	sessionNumber int

	state *sessionState

	lock sync.Mutex
}

func newSession(log *log.Logger, beaconSender *BeaconSender, beacon *Beacon) Session {
	s := new(session)

	s.log = log
	s.beacon = beacon
	s.beaconSender = beaconSender
	s.state = newSessionState(s)
	s.ID = beacon.sessionNumber
	s.numNewSessionRequestsLeft = 4
	beaconSender.startSession(s)
	beacon.startSession()

	return s
}

func (s *session) clearCapturedData() {
	s.beacon.beaconCache.deleteCacheEntry(s.beacon.sessionNumber)
}

func (s *session) EnterAction(actionName string) Action {
	s.log.WithFields(log.Fields{"actionName": actionName}).Debugf("Session.EnterAction")

	return newRootAction(s.log, s, actionName, s.beacon)

}

func (s *session) EnterActionAt(actionName string, timestamp time.Time) Action {
	s.log.Debugf("Session.EnterActionAt(%s, %s)", actionName, timestamp.String())

	return newRootActionAt(s.log, s, actionName, s.beacon, timestamp)

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

	if !s.state.markAsIsFinishing() {
		return
	}

	children := s.getCopyOfChildObjects()
	for _, c := range children {
		c.close()
	}

	s.beacon.endSession(s)
	s.beaconSender.finishSession(s)
	s.state.markAsFinished()

}

func (s *session) EndAt(timestamp time.Time) {
	s.log.Debugf("Session.End()")

	if !s.state.markAsIsFinishing() {
		return
	}

	children := s.getCopyOfChildObjects()
	for _, c := range children {
		c.close()
	}

	s.endTime = timestamp
	s.beacon.endSession(s)
	s.beaconSender.finishSession(s)
	s.state.markAsFinished()
}

func (s *session) getActionID() int {
	return 0
}

func (s *session) TraceWebRequest(url string) *WebRequestTracer {
	w := NewWebRequestTracer(s.log, s, url, s.beacon)
	s.storeChildInList(w)
	return w
}

func (s *session) TraceWebRequestAt(url string, timestamp time.Time) *WebRequestTracer {
	w := NewWebRequestTracerAt(s.log, s, url, s.beacon, timestamp)
	s.storeChildInList(w)
	return w
}

func (s *session) storeChildInList(child OpenKitObject) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.children = append(s.children, child)
}

func (s *session) removeChildFromList(child OpenKitObject) bool {
	tmp := s.children[:0]
	for _, c := range s.children {
		if child != c {
			tmp = append(tmp, c)
		}
	}
	s.children = tmp
	return true
}

func (s *session) getCopyOfChildObjects() []OpenKitObject {
	s.lock.Lock()
	defer s.lock.Unlock()
	copied := make([]OpenKitObject, len(s.children))
	copy(copied, s.children)
	return copied
}

func (s *session) close() {
	s.End()
}

func (s *session) onChildClosed(child OpenKitObject) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.removeChildFromList(child)

}

func (s *session) getChildCount() int {
	s.lock.Lock()
	s.lock.Unlock()
	return len(s.children)
}
