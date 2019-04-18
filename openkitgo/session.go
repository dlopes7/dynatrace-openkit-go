package openkitgo

import "github.com/op/go-logging"

type Session interface {
	EnterAction(string) Action
	IdentifyUser(string)
	ReportCrash(string, string, string)
	TraceWebRequest(string)
	End()
}

type session struct {
	endTime uint64

	beaconSender BeaconSender
	beacon       Beacon
	logger       logging.Logger
}

/*
func NewSession(logger logging.Logger, beaconSender BeaconSender, beacon Beacon) Session {
	s := new(session)
	s.logger = logger
	s.beaconSender = beaconSender
	s.beacon = beacon

	// TODO         beaconSender.startSession(this);
	// TODO         beacon.startSession();

	return s
}*/
