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

func (s *SessionState) isConfigured() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	// TODO return session.beacon.isServerConfigurationSet();
	return true
}

func (s *SessionState) isConfiguredAndFinished() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.isConfigured() && s.finished
}

func (s *SessionState) isConfiguredAndOpen() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.isConfigured() && !s.finished
}

func (s *SessionState) isFinishingOrFinished() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.finishing || s.finished
}

func (s *SessionState) markAsIsFinishing() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.isFinishingOrFinished() {
		return false
	}
	s.finishing = true
	return true
}

func (s *SessionState) markAsFinished() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.finishing = true
}

func (s *SessionState) markAsWasTriedForEnding() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.triedForEnding = true
}

func (s *SessionState) wasTriedForEnding() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.triedForEnding
}

func (s *SessionState) isFinished() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.finished
}
