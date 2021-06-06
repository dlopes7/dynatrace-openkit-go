package core

import "sync"

type SessionState struct {
	session        *Session
	finishing      bool
	finished       bool
	triedForEnding bool
	mutex          sync.RWMutex
}

func NewSessionState(session *Session) *SessionState {
	return &SessionState{
		session: session,
	}
}

func (s *SessionState) IsConfigured() bool {
	return s.session.beacon.isServerConfigurationSet()

}

func (s *SessionState) IsConfiguredAndFinished() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.IsConfigured() && s.finished
}

func (s *SessionState) IsConfiguredAndOpen() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.IsConfigured() && !s.finished
}

func (s *SessionState) IsFinishingOrFinished() bool {
	return s.finishing || s.finished
}

func (s *SessionState) MarkAsIsFinishing() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if s.IsFinishingOrFinished() {
		return false
	}
	s.finishing = true
	return true
}

func (s *SessionState) MarkAsFinished() {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	s.finished = true
}

func (s *SessionState) MarkAsWasTriedForEnding() {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	s.triedForEnding = true
}

func (s *SessionState) WasTriedForEnding() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.triedForEnding
}

func (s *SessionState) IsFinished() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.finished
}
