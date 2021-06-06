package core

import (
	"fmt"
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo"
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo/caching"
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo/configuration"
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo/providers"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

type ServerConfigurationUpdateCallback interface {
	onServerConfigurationUpdate(serverConfiguration *configuration.ServerConfiguration)
}

type SessionProxy struct {

	// From java SessionProxyImpl
	log                  *log.Logger
	parent               OpenKitComposite
	openKitConfiguration *configuration.OpenKitConfiguration
	privacyConfiguration *configuration.PrivacyConfiguration
	beaconSender         *BeaconSender
	// TODO SessionWatchdog
	currentSession      *Session
	topLevelActionCount int
	lastInteractionTime time.Time
	serverConfiguration *configuration.ServerConfiguration
	isFinished          bool
	lastUserTag         string

	// From java SessionCreatorImpl
	beaconCache           *caching.BeaconCache
	clientIPAddress       string
	serverID              int
	sessionSequenceNumber int32

	children []OpenKitObject
	mutex    sync.Mutex
}

func NewSessionProxy(
	log *log.Logger,
	parent OpenKitComposite,
	beaconSender *BeaconSender,
	// TODO sessionWatchdog
	input *OpenKit,
	clientIPAddress string,
	timestamp time.Time,
) *SessionProxy {
	p := &SessionProxy{
		// Proxy
		log:          log,
		parent:       parent,
		beaconSender: beaconSender,

		// Creator
		openKitConfiguration: input.openKitConfiguration,
		privacyConfiguration: input.privacyConfiguration,
		beaconCache:          input.beaconCache,
		clientIPAddress:      clientIPAddress,
		serverID:             beaconSender.GetCurrentServerId(),

		currentSession:      nil,
		topLevelActionCount: 0,
		lastInteractionTime: time.Time{},
		isFinished:          false,
		lastUserTag:         "",

		children: []OpenKitObject{},
	}

	currentServerConfig := beaconSender.GetLastServerConfiguration()
	p.createInitialSessionAndMakeCurrent(currentServerConfig, timestamp)

	return p
}

func (p *SessionProxy) storeChildInList(child OpenKitObject) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.children = append(p.children, child)
}

func (p *SessionProxy) removeChildFromList(child OpenKitObject) bool {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	removed := false

	var keep []OpenKitObject
	for _, c := range p.children {
		if c != child {
			keep = append(keep, c)
		} else {
			removed = true
		}
	}
	p.children = keep
	return removed
}

func (p *SessionProxy) getCopyOfChildObjects() []OpenKitObject {
	return p.children[:]
}

func (p *SessionProxy) getChildCount() int {
	return len(p.children)
}

func (p *SessionProxy) onChildClosed(child OpenKitObject) {
	p.removeChildFromList(child)
	/*
		TODO
		if (childObject instanceof SessionImpl) {
			sessionWatchdog.dequeueFromClosing((SessionImpl) childObject);
		}
	*/

}

func (p *SessionProxy) getActionID() int {
	return DEFAULT_ACTION_ID
}

func (p *SessionProxy) EnterAction(actionName string) openkitgo.Action {
	return p.EnterActionAt(actionName, time.Now())
}

func (p *SessionProxy) EnterActionAt(actionName string, timestamp time.Time) openkitgo.Action {
	if actionName == "" {
		p.log.Warning("actionName must not be empty")
		return NewNullAction()
	}
	p.log.WithFields(log.Fields{"actionName": actionName}).Debug("SessionProxy.EnterAction()")

	p.mutex.Lock()
	defer p.mutex.Unlock()
	if !p.isFinished {
		session := p.getOrSplitCurrentSessionByEvents(timestamp)
		p.topLevelActionCount++
		return session.EnterActionAt(actionName, timestamp)
	}

	return NewNullAction()
}

func (p *SessionProxy) IdentifyUser(userTag string) {
	p.IdentifyUserAt(userTag, time.Now())
}

func (p *SessionProxy) IdentifyUserAt(userTag string, timestamp time.Time) {
	p.log.WithFields(log.Fields{"userTag": userTag}).Debug("SessionProxy.IdentifyUser()")

	if !p.isFinished {
		s := p.getOrSplitCurrentSessionByEvents(timestamp)
		p.lastInteractionTime = time.Now()
		s.IdentifyUserAt(userTag, timestamp)
		p.lastUserTag = userTag

	}
}

func (p *SessionProxy) ReportCrash(errorName string, reason string, stacktrace string) {
	p.ReportCrashAt(errorName, reason, stacktrace, time.Now())
}

func (p *SessionProxy) ReportCrashAt(errorName string, reason string, stacktrace string, timestamp time.Time) {
	p.log.WithFields(log.Fields{"errorName": errorName, "reason": reason, "stacktrace": stacktrace, "timestamp": timestamp}).Debug("SessionProxy.ReportCrash()")

	if !p.isFinished {
		s := p.getOrSplitCurrentSessionByEvents(timestamp)
		p.topLevelActionCount++
		s.ReportCrashAt(errorName, reason, stacktrace, timestamp)
		p.splitAndCreateNewInitialSession()
	}
}
func (p *SessionProxy) End() {
	p.EndAt(time.Now())
}

func (p *SessionProxy) EndAt(timestamp time.Time) {
	p.log.Debug("SessionProxy.End()")

	p.mutex.Lock()
	if p.isFinished {
		p.mutex.Unlock()
		return
	}
	p.isFinished = true
	p.mutex.Unlock()

	p.closeChildObjects(timestamp)
	p.parent.onChildClosed(p)
	// TODO sessionWatchdog.removeFromSplitByTimeout(this);

}

func (p *SessionProxy) String() string {
	return fmt.Sprintf("SessionProxy")
}

func (p *SessionProxy) closeAt(timestamp time.Time) {
	p.EndAt(timestamp)
}

func (p *SessionProxy) close() {
	p.closeAt(time.Now())
}

func (p *SessionProxy) GetSessionSequenceNumber() int32 {
	return p.sessionSequenceNumber
}

func (p *SessionProxy) getOrSplitCurrentSessionByEvents(timestamp time.Time) openkitgo.Session {
	if p.isSessionSplitByEventsRequired() {
		p.closeOrEnqueueCurrentSessionForClosing()
		p.createSplitSessionAndMakeCurrent(p.serverConfiguration, timestamp)
		p.reTagCurrentSession()
	}

	return p.currentSession
}

func (p *SessionProxy) createInitialSessionAndMakeCurrent(initialServerConfig *configuration.ServerConfiguration, timestamp time.Time) {
	p.createAndAssignCurrentSession(initialServerConfig, nil, timestamp)
}

func (p *SessionProxy) createAndAssignCurrentSession(initialServerConfig *configuration.ServerConfiguration, updatedServerConfig *configuration.ServerConfiguration, timestamp time.Time) {

	session := p.createSessionAt(p, timestamp)
	beacon := session.(*Session).beacon
	beacon.setServerConfigurationUpdateCallback(p)
	p.storeChildInList(session.(*Session))

	p.lastInteractionTime = beacon.GetSessionStartTime()
	p.topLevelActionCount = 0

	if initialServerConfig != nil {
		// TODO session.(*Session).initializeServerConfiguration(initialServerConfig)
	}

	if updatedServerConfig != nil {
		// TODO session.(*Session).updateServerConfiguration(updatedServerConfig)
	}

	p.mutex.Lock()
	p.currentSession = session.(*Session)
	p.mutex.Unlock()

	p.beaconSender.AddSession(session.(*Session))

}

func (p *SessionProxy) createSessionAt(parent OpenKitComposite, timestamp time.Time) openkitgo.Session {

	config := configuration.NewBeaconConfiguration(
		p.openKitConfiguration,
		p.privacyConfiguration,
		p.serverID)

	beacon := NewBeacon(
		p.log,
		p.beaconCache,
		providers.NewSessionIDProvider(),
		p,
		config,
		timestamp,
		p.openKitConfiguration.DeviceID,
		p.clientIPAddress,
	)

	session := NewSession(p.log, parent, beacon, timestamp)
	p.sessionSequenceNumber++

	return session
}

func (p *SessionProxy) closeChildObjects(timestamp time.Time) {
	for _, child := range p.getCopyOfChildObjects() {
		switch child.(type) {
		case openkitgo.Session:
			child.(*Session).endWithEvent(child == p.currentSession, timestamp)
		default:
			child.closeAt(timestamp)
		}
	}
}

func (p *SessionProxy) isSessionSplitByEventsRequired() bool {

	if p.serverConfiguration == nil || !p.serverConfiguration.IsSendingDataAllowed() {
		return false
	}
	return p.serverConfiguration.MaxEventsPerSession <= p.topLevelActionCount

}

func (p *SessionProxy) closeOrEnqueueCurrentSessionForClosing() {
	if p.serverConfiguration == nil {
		p.serverConfiguration = configuration.DefaultServerConfiguration()
	}
	closeGracePeriod := p.serverConfiguration.SessionTimeout
	if closeGracePeriod != 0 {
		closeGracePeriod = closeGracePeriod / 2
	} else {
		closeGracePeriod = p.serverConfiguration.SendInterval
	}

	// TODO         sessionWatchdog.closeOrEnqueueForClosing(currentSession, closeGracePeriodInMillis);
}

func (p *SessionProxy) createSplitSessionAndMakeCurrent(serverConfiguration *configuration.ServerConfiguration, timestamp time.Time) {
	p.createAndAssignCurrentSession(nil, serverConfiguration, timestamp)

}

func (p *SessionProxy) onServerConfigurationUpdate(serverConfiguration *configuration.ServerConfiguration) {
	p.serverConfiguration = serverConfiguration

	if p.isFinished {
		return
	}

	if p.serverConfiguration.SessionSplitByIdleTimeout || p.serverConfiguration.SessionSplitBySessionDuration {
		// TODO sessionWatchdog.addToSplitByTimeout(this);
	}

}

func (p *SessionProxy) splitAndCreateNewInitialSession() {
	p.closeOrEnqueueCurrentSessionForClosing()
	p.sessionSequenceNumber = 0
	p.createInitialSessionAndMakeCurrent(p.serverConfiguration, time.Now())
	p.reTagCurrentSession()

}

func (p *SessionProxy) reTagCurrentSession() {
	if p.lastUserTag != "" {
		p.currentSession.IdentifyUser(p.lastUserTag)
	}

}

func (p *SessionProxy) TraceWebRequest(url string) openkitgo.WebRequestTracer {
	return p.TraceWebRequestAt(url, time.Now())
}

func (p *SessionProxy) TraceWebRequestAt(url string, timestamp time.Time) openkitgo.WebRequestTracer {
	p.log.WithFields(log.Fields{"url": url, "timestamp": timestamp}).Debug("SessionProxy.TraceWebRequest()")

	if !p.isFinished {
		s := p.getOrSplitCurrentSessionByEvents(timestamp)
		p.topLevelActionCount++
		return s.TraceWebRequestAt(url, timestamp)
	}

	return NewNullWebRequestTracer()
}
