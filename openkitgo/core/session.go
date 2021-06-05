package core

import (
	"fmt"
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo"
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

func (s *Session) EnterAction(actionName string) openkitgo.IAction {
	panic("implement me")
}

func (s *Session) EnterActionAt(actionName string, timestamp time.Time) openkitgo.IAction {
	panic("implement me")
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

func NewSession(log *log.Logger, parent OpenKitComposite, beacon *Beacon) *Session {

	s := &Session{
		log:               log,
		parent:            parent,
		beacon:            beacon,
		remainingRequests: MAX_NEW_SESSION_REQUESTS,
	}
	s.State = NewSessionState(s)

	// TODO - s.beacon.StartSession()
	return s
}

func (s *Session) getCopyOfChildObjects() []OpenKitObject {
	return s.children[:]
}

func (s *Session) onChildClosed(child OpenKitObject) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.State.mutex.Lock()
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
	s.log.WithFields(log.Fields{"session": s}).Debug("end()")

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

func (s *Session) ClearCapturedData() {
	s.beacon.ClearData()
}
