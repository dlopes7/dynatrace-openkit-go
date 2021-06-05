package core

import (
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo/caching"
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo/configuration"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

type SessionProxy struct {

	// From java SessionProxyImpl
	log                  *log.Logger
	parent               OpenKitComposite
	openKitConfiguration configuration.OpenKitConfiguration
	privacyConfiguration configuration.PrivacyConfiguration
	// TODO BeaconSender
	// TODO SessionWatchdog
	currentSession      *Session
	topLevelActionCount int
	lastInteractionTime time.Time
	serverConfiguration configuration.ServerConfiguration
	isFinished          bool
	lastUserTag         string

	// From java SessionCreatorImpl
	beaconCache           *caching.BeaconCache
	clientIPAddress       string
	serverID              int
	sessionSequenceNumber uint32

	mutex sync.Mutex
}

func NewSessionProxy() *SessionProxy {
	return &SessionProxy{
		log:                   nil,
		parent:                nil,
		openKitConfiguration:  configuration.OpenKitConfiguration{},
		privacyConfiguration:  configuration.PrivacyConfiguration{},
		currentSession:        nil,
		topLevelActionCount:   0,
		lastInteractionTime:   time.Time{},
		serverConfiguration:   configuration.ServerConfiguration{},
		isFinished:            false,
		lastUserTag:           "",
		beaconCache:           nil,
		clientIPAddress:       "",
		serverID:              0,
		sessionSequenceNumber: 0,
		mutex:                 sync.Mutex{},
	}
}

func (p *SessionProxy) GetSessionSequenceNumber() uint32 {
	return p.sessionSequenceNumber
}
