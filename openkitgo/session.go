package openkitgo

import (
	"fmt"
	"github.com/op/go-logging"
)

type Session interface {
	EnterAction(string) Action
	IdentifyUser(string)
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

	sendBeacon(*HttpClient) *StatusResponse

	clearCapturedData()
}

type session struct {
	ID      int
	endTime int

	beaconSender *BeaconSender
	beacon       *Beacon
	logger       *logging.Logger

	openRootActions map[int]Action

	sessionFinished           bool
	beaconConfigurationSet    bool
	numNewSessionRequestsLeft int

	position      int
	sessionNumber int
}

func newSession(logger *logging.Logger, beaconSender *BeaconSender, beacon *Beacon) Session {
	s := new(session)

	s.logger = logger
	s.beaconSender = beaconSender
	s.beacon = beacon
	s.ID = s.beacon.config.createSessionNumber()
	s.openRootActions = make(map[int]Action)

	s.numNewSessionRequestsLeft = 4
	beaconSender.startSession(s)
	beacon.startSession()

	return s
}

func (s *session) clearCapturedData() {
	s.beacon.beaconCache.deleteCacheEntry(s.beacon.sessionNumber)
}

func (s *session) EnterAction(actionName string) Action {
	s.logger.Debugf("enterAction(%s)", actionName)

	return newRootAction(s.logger, s.beacon, actionName, s.openRootActions)

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

func (s *session) IdentifyUser(userTag string) {
	s.logger.Debugf("identifyUser(%s)\n", userTag)
	s.beacon.identifyUser(userTag)
}

func (s *session) End() {
	s.logger.Debug("Session.end()")

	s.endTime = s.beacon.getCurrentTimestamp()

	fmt.Println("SESSION!", len(s.openRootActions))
	for len(s.openRootActions) != 0 {
		for _, a := range s.openRootActions {
			a.LeaveAction()
		}
	}

	s.beacon.endSession(s)
	s.beaconSender.finishSession(s)
}
