package openkitgo

import "github.com/op/go-logging"

type Session struct {
	endTime uint64

	beaconSender BeaconSender
	beacon       Beacon
	logger       logging.Logger
}

func NewSession(logger logging.Logger, beaconSender BeaconSender, beacon Beacon) *Session {
	s := new(Session)
	s.logger = logger
	s.beaconSender = beaconSender
	s.beacon = beacon

	// TODO         beaconSender.startSession(this);
	// TODO         beacon.startSession();

	return s
}
