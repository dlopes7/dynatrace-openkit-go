package core

import (
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo/caching"
	log "github.com/sirupsen/logrus"
)

type SessionProxy struct {
	log *log.Logger
	// TODO openKitConfiguration
	// TODO privacyConfiguration
	// TODO beaconCache
	beaconCache     *caching.BeaconCache
	clientIPAddress string
	serverID        int

	sessionSequenceNumber uint32

	// TODO clientIpAddress
}

func (p *SessionProxy) GetSessionSequenceNumber() uint32 {
	return p.sessionSequenceNumber
}
