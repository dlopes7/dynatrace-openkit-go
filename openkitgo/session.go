package openkitgo

import "github.com/op/go-logging"

type Session interface {
	enterAction(string) Action
	// identifyUser(string)
	// reportCrash(string, string, string)
	// traceWebRequest(string)
	End()
	finishSession()
}

type session struct {
	endTime int

	beaconSender *BeaconSender
	beacon       *Beacon
	logger       *logging.Logger

	openRootActions map[int]Action
	sessionFinished bool
}

func (s *session) enterAction(actionName string) Action {
	s.logger.Debugf("enterAction(%s)", actionName)

	return NewAction(s.logger, s.beacon, actionName, s.openRootActions)

}

func (s *session) finishSession() {
	s.sessionFinished = true
}

func NewSession(logger *logging.Logger, beaconSender *BeaconSender, beacon *Beacon) Session {

	s := &session{
		logger:          logger,
		beaconSender:    beaconSender,
		beacon:          beacon,
		openRootActions: make(map[int]Action),
	}

	beaconSender.startSession(s)
	beacon.startSession()

	return s
}

func (s *session) End() {
	s.logger.Debug("Session.end()")

	for len(s.openRootActions) != 0 {
		for _, a := range s.openRootActions {
			a.leaveAction()
		}
	}

	s.beacon.endSession(s)
	s.beaconSender.finishSession(s)
}
