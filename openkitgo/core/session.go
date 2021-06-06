package core

import (
	"fmt"
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo/configuration"
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo/interfaces"
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo/protocol"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

const (
	MAX_NEW_SESSION_REQUESTS = 4
)

type Session struct {
	log                             *log.Logger
	parent                          OpenKitComposite
	beacon                          *Beacon
	State                           *SessionState
	remainingRequests               int
	children                        []OpenKitObject
	mutex                           sync.RWMutex
	splitByEventsGracePeriodEndTime time.Time
}

func (s *Session) EnterAction(actionName string) interfaces.Action {
	return s.EnterActionAt(actionName, time.Now())
}

func (s *Session) EnterActionAt(actionName string, timestamp time.Time) interfaces.Action {
	s.log.WithFields(log.Fields{"actionName": actionName, "timestamp": timestamp}).Debug("Session.EnterActionAt()")

	if !s.State.IsFinishingOrFinished() {
		action := NewAction(s.log, s, nil, actionName, s.beacon, timestamp)
		s.storeChildInList(action)
		return action
	}
	return NewNullAction()
}

func (s *Session) IdentifyUser(userTag string) {
	s.IdentifyUserAt(userTag, time.Now())
}

func (s *Session) IdentifyUserAt(userTag string, timestamp time.Time) {
	s.log.WithFields(log.Fields{"session": s, "userTag": userTag, "timestamp": timestamp}).Debug("Session.IdentifyUser()")

	if !s.State.IsFinishingOrFinished() {
		s.beacon.identifyUser(userTag, timestamp)
	}
}

func (s *Session) ReportCrash(errorName string, reason string, stacktrace string) {
	s.ReportCrashAt(errorName, reason, stacktrace, time.Now())
}

func (s *Session) ReportCrashAt(errorName string, reason string, stacktrace string, timestamp time.Time) {
	s.log.WithFields(log.Fields{"session": s, "errorName": errorName, "reason": reason, "stacktrace": stacktrace, "timestamp": timestamp}).Debug("Session.ReportCrash()")

	if !s.State.IsFinishingOrFinished() {
		s.beacon.reportCrash(errorName, reason, stacktrace, timestamp)
	}
}

func (s *Session) String() string {
	return fmt.Sprintf("Session(%d)", s.beacon.GetSessionNumber())
}

func NewSession(log *log.Logger, parent OpenKitComposite, beacon *Beacon, timestamp time.Time) *Session {

	s := &Session{
		log:               log,
		parent:            parent,
		beacon:            beacon,
		remainingRequests: MAX_NEW_SESSION_REQUESTS,
	}
	s.State = NewSessionState(s)

	beacon.startSession()
	return s
}

func (s *Session) getCopyOfChildObjects() []OpenKitObject {
	return s.children[:]
}

func (s *Session) onChildClosed(child OpenKitObject) {
	s.removeChildFromList(child)

	if s.State.WasTriedForEnding() && s.getChildCount() == 0 {
		s.endWithEvent(false, time.Now())
	}

}

func (s *Session) storeChildInList(child OpenKitObject) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	s.children = append(s.children, child)

}

func (s *Session) removeChildFromList(child OpenKitObject) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	removed := false

	var keep []OpenKitObject
	for _, c := range s.children {
		if c != child {
			keep = append(keep, c)
		} else {
			removed = true
		}
	}
	s.children = keep
	return removed
}

func (s *Session) getChildCount() int {
	return len(s.children)
}

func (s *Session) getActionID() int {
	return DEFAULT_ACTION_ID
}

func (s *Session) close() {
	s.closeAt(time.Now())
}

func (s *Session) closeAt(timestamp time.Time) {
	s.endWithEvent(true, timestamp)
}

func (s *Session) End() {
	s.EndAt(time.Now())
}

func (s *Session) EndAt(timestamp time.Time) {
	s.endWithEvent(true, timestamp)
}

func (s *Session) endWithEvent(sendEvent bool, timestamp time.Time) {
	s.log.WithFields(log.Fields{"session": s, "sendEvent": sendEvent, "timestamp": timestamp}).Debug("Session.end()")

	// End was already called before
	if !s.State.MarkAsIsFinishing() {
		return
	}

	for _, child := range s.getCopyOfChildObjects() {
		child.closeAt(timestamp)
	}

	if sendEvent {
		s.beacon.EndSession()
	}

	s.State.MarkAsFinished()
	s.parent.onChildClosed(s)
	s.parent = nil
}

func (s *Session) clearCapturedData() {
	s.beacon.ClearData()
}

func (s *Session) canSendNewSessionRequest() bool {
	return s.remainingRequests > 0
}

func (s *Session) disableCapture() {
	s.beacon.disableCapture()
}

func (s *Session) updateServerConfiguration(config *configuration.ServerConfiguration) {
	s.beacon.updateServerConfiguration(config)
}

func (s *Session) sendBeacon(ctx *BeaconSendingContext) protocol.StatusResponse {
	return s.beacon.send(ctx)
}

func (s *Session) isDataSendingAllowed() bool {
	return s.State.IsConfigured() && s.beacon.isDataCapturingEnabled()
}

func (s *Session) isEmpty() bool {
	return s.beacon.IsEmpty()

}

func (s *Session) TraceWebRequest(url string) interfaces.WebRequestTracer {
	return s.TraceWebRequestAt(url, time.Now())
}

func (s *Session) TraceWebRequestAt(url string, timestamp time.Time) interfaces.WebRequestTracer {
	s.log.WithFields(log.Fields{"session": s, "url": url, "timestamp": timestamp}).Debug("Session.TraceWebRequest()")
	if !s.State.IsFinishingOrFinished() {
		tracer := NewWebRequestTracer(s.log, s, url, s.beacon, timestamp)
		s.storeChildInList(tracer)
		return tracer
	}
	return NewNullWebRequestTracer()
}

func (s *Session) enableCapture() {
	s.beacon.enableCapture()
}

func (s *Session) getSplitByEventsGracePeriodEndTime() time.Time {
	return s.splitByEventsGracePeriodEndTime
}

func (s *Session) setSplitByEventsGracePeriodEndTime(timestamp time.Time) {
	s.splitByEventsGracePeriodEndTime = timestamp
}

func (s *Session) tryEnd() bool {
	if s.State.IsConfiguredAndFinished() {
		return true
	}
	if s.getChildCount() == 0 {
		s.endWithEvent(false, time.Now())
		return true
	}
	s.State.MarkAsWasTriedForEnding()
	return false

}

func (s *Session) initializeServerConfiguration(config *configuration.ServerConfiguration) {
	s.beacon.initializeServerConfiguration(config)
}
