package core

import (
	"fmt"
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo"
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo/configuration"
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo/protocol"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

const (
	MAX_NEW_SESSION_REQUESTS = 4
)

type Session struct {
	log               *log.Logger
	parent            OpenKitComposite
	beacon            *Beacon
	State             SessionState
	remainingRequests int
	splitEndTime      time.Time
	children          []OpenKitObject
	mutex             sync.Mutex
}

func (s *Session) EnterAction(actionName string) openkitgo.Action {
	return s.EnterActionAt(actionName, time.Now())
}

func (s *Session) EnterActionAt(actionName string, timestamp time.Time) openkitgo.Action {
	s.log.WithFields(log.Fields{"actionName": actionName, "timestamp": timestamp}).Debug("Session.EnterActionAt()")

	if !s.State.IsFinishingOrFinished() {
		action := NewAction(s.log, s, actionName, s.beacon, timestamp)
		s.storeChildInList(action)
		return action
	}
	return NewNullAction()
}

func (s *Session) IdentifyUser(userTag string) {
	panic("implement me")
}

func (s *Session) IdentifyUserAt(userTag string, timestamp time.Time) {
	panic("implement me")
}

func (s *Session) ReportCrash(errorName string, reason string, stacktrace string) {
	panic("implement me")
}

func (s *Session) ReportCrashAt(errorName string, reason string, stacktrace string, timestamp time.Time) {
	panic("implement me")
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
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.children = append(s.children, child)

}

func (s *Session) removeChildFromList(child OpenKitObject) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
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
