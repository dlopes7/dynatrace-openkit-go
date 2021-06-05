package core

import "sync"

type SessionState struct {
	session        *Session
	finishing      bool
	finished       bool
	triedForEnding bool
	mutex          sync.Mutex
}

func NewSessionState(session *Session) SessionState {
	return SessionState{
		session: session,
	}
}

func (s *SessionState) IsConfigured() bool {
	return s.session.beacon.isServerConfigurationSet()

}

func (s *SessionState) IsConfiguredAndFinished() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.IsConfigured() && s.finished
}

func (s *SessionState) IsConfiguredAndOpen() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.IsConfigured() && !s.finished
}

func (s *SessionState) IsFinishingOrFinished() bool {
	return s.finishing || s.finished
}

func (s *SessionState) MarkAsIsFinishing() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.IsFinishingOrFinished() {
		return false
	}
	s.finishing = true
	return true
}

func (s *SessionState) MarkAsFinished() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.finished = true
}

func (s *SessionState) MarkAsWasTriedForEnding() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.triedForEnding = true
}

func (s *SessionState) WasTriedForEnding() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.triedForEnding
}

func (s *SessionState) IsFinished() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.finished
}
