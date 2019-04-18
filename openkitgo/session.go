package openkitgo

import "github.com/op/go-logging"

type Session interface {
	enterAction(string) Action
	// identifyUser(string)
	// reportCrash(string, string, string)
	// traceWebRequest(string)
	End()
	finishSession()

	isBeaconConfigurationSet() bool
	canSendNewSessionRequest() bool
	isSessionFinished() bool
	isDataSendingAllowed() bool

	getBeaconConfiguration() *BeaconConfiguration
	updateBeaconConfiguration(*BeaconConfiguration)

	sendBeacon() *StatusResponse
}

type session struct {
	endTime int

	beaconSender *BeaconSender
	beacon       *Beacon
	logger       *logging.Logger

	openRootActions map[int]Action

	sessionFinished           bool
	beaconConfigurationSet    bool
	numNewSessionRequestsLeft int
}

func NewSession(logger *logging.Logger, beaconSender *BeaconSender, beacon *Beacon) Session {
	s := new(session)

	s.logger = logger
	s.beaconSender = beaconSender
	s.beacon = beacon
	s.openRootActions = make(map[int]Action)

	s.numNewSessionRequestsLeft = 4
	beaconSender.startSession(s)
	beacon.startSession()

	return s
}

func (s *session) enterAction(actionName string) Action {
	s.logger.Debugf("enterAction(%s)", actionName)

	return NewAction(s.logger, s.beacon, actionName, s.openRootActions)

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
