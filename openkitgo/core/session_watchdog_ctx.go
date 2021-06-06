package core

import (
	"sync/atomic"
	"time"
)

const (
	SESSION_WATCHDOG_DEFAULT_SLEEP_TIME = 5 * time.Second
)

type SessionWatchdogContext struct {
	shutdown                 int32 // atomic
	sessionsToClose          []*Session
	sessionsToSplitByTimeout []*SessionProxy
}

func NewSessionWatchdogContext() *SessionWatchdogContext {
	return &SessionWatchdogContext{}
}

func (c *SessionWatchdogContext) execute() {
	durationToNextClose := c.closeExpiredSessions()
	durationToNextSplit := c.splitTimedOutSessions()

	if durationToNextSplit < durationToNextClose {
		time.Sleep(durationToNextSplit)
	} else {
		time.Sleep(durationToNextClose)
	}
}

func (c *SessionWatchdogContext) splitTimedOutSessions() time.Duration {
	sleepTime := SESSION_WATCHDOG_DEFAULT_SLEEP_TIME

	for _, session := range c.sessionsToSplitByTimeout {
		nextSessionSplitTime := session.splitSessionByTime()
		if nextSessionSplitTime.IsZero() {
			continue
		}

		now := time.Now()
		durationToNextSplit := nextSessionSplitTime.Sub(now)
		if durationToNextSplit < 0 {
			continue
		}

		if durationToNextSplit < sleepTime {
			sleepTime = durationToNextSplit
		}
	}

	return sleepTime
}

func (c *SessionWatchdogContext) closeExpiredSessions() time.Duration {
	sleepTime := SESSION_WATCHDOG_DEFAULT_SLEEP_TIME

	var sessionsToEnd []*Session

	for _, session := range c.sessionsToClose {
		now := time.Now()
		gracePeriodEndTime := session.getSplitByEventsGracePeriodEndTime()
		gracePeriodExpired := gracePeriodEndTime.Before(now)
		if gracePeriodExpired {
			sessionsToEnd = append(sessionsToEnd, session)
			continue
		}
		sleepTimeToGracePeriodEnd := gracePeriodEndTime.Sub(now)
		if sleepTimeToGracePeriodEnd < sleepTime {
			sleepTime = sleepTimeToGracePeriodEnd
		}
	}

	for _, session := range sessionsToEnd {
		session.endWithEvent(false, time.Now())
	}

	return sleepTime
}

func (c *SessionWatchdogContext) requestShutdown() {
	atomic.StoreInt32(&c.shutdown, 1)
}

func (c *SessionWatchdogContext) isShutdownRequested() bool {
	return atomic.LoadInt32(&c.shutdown) == 1
}

func (c *SessionWatchdogContext) closeOrEnqueueForClosing(session *Session, closeGracePeriod time.Duration) {
	if session.tryEnd() {
		return
	}
	closeTime := time.Now().Add(closeGracePeriod)
	session.setSplitByEventsGracePeriodEndTime(closeTime)
	c.sessionsToClose = append(c.sessionsToClose, session)
}

func (c *SessionWatchdogContext) dequeueFromClosing(session *Session) {
	var keep []*Session

	for _, s := range c.sessionsToClose {
		if s != session {
			keep = append(keep, s)
		}
	}
	c.sessionsToClose = keep

}
func (c *SessionWatchdogContext) addToSplitByTimeout(session *SessionProxy) {
	if session.isFinished {
		return
	}
	c.sessionsToSplitByTimeout = append(c.sessionsToSplitByTimeout, session)
}

func (c *SessionWatchdogContext) removeFromSplitByTimeout(session *SessionProxy) {
	var keep []*SessionProxy

	for _, s := range c.sessionsToSplitByTimeout {
		if s != session {
			keep = append(keep, s)
		}
	}
	c.sessionsToSplitByTimeout = keep

}
