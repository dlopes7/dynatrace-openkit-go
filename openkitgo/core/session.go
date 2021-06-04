package core

import (
	log "github.com/sirupsen/logrus"
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

func (s *Session) StoreChildInList(child OpenKitObject) {}
func (s *Session) RemoveChildFromList(child OpenKitObject) bool {
	return true
}
func (s *Session) GetCopyOfChildObjects() []OpenKitObject {
	return []OpenKitObject{}
}
func (s *Session) GetChildCount()                    {}
func (s *Session) OnChildClosed(child OpenKitObject) {}
func (s *Session) GetActionID() int {
	return 0
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

	for _, child := range s.GetCopyOfChildObjects() {
		child.closeAt(timestamp)
	}

	if sendEvent {
		s.beacon.EndSession()
	}

	s.State.MarkAsFinished()
	s.parent.OnChildClosed(s)
	s.parent = nil
}

func (s *Session) ClearCapturedData() {
	s.beacon.ClearData()
}
