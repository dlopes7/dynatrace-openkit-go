package core

import (
	log "github.com/sirupsen/logrus"
	"time"
)

const (
	MAX_NEW_SESSION_REQUESTS = 4
)

type Session struct {
	log    *log.Logger
	parent OpenKitComposite
	// TODO beacon *Beacon
	state             SessionState
	remainingRequests int
	splitEndTime      time.Time
}

func NewSession(log *log.Logger, parent OpenKitComposite /*TODO beacon *Beacon*/) *Session {

	s := &Session{
		log:    log,
		parent: parent,
		// TODO - beacon: beacon
		remainingRequests: MAX_NEW_SESSION_REQUESTS,
	}
	s.state = NewSessionState(s)

	// TODO - s.beacon.startSession()
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
	if !s.state.markAsIsFinishing() {
		return
	}

	for _, child := range s.GetCopyOfChildObjects() {
		child.closeAt(timestamp)
	}

	if sendEvent {
		// TODO - s.beacon.endSession()
	}

	s.state.markAsFinished()
	s.parent.OnChildClosed(s)
	s.parent = nil
}
