package core

import (
	log "github.com/sirupsen/logrus"
	"time"
)

type SessionWatchdog struct {
	log *log.Logger
	ctx *SessionWatchdogContext
}

func NewSessionWatchdog(log *log.Logger, ctx *SessionWatchdogContext) *SessionWatchdog {
	return &SessionWatchdog{
		log: log,
		ctx: ctx,
	}
}

func sessionWatchdogGoRoutine(log *log.Logger, ctx *SessionWatchdogContext) {

	go func() {
		log.Debug("SessionWatchdogGoRoutine.run()")
		for !ctx.isShutdownRequested() {
			ctx.execute()
		}
	}()

}

func (w *SessionWatchdog) Initialize() {
	sessionWatchdogGoRoutine(w.log, w.ctx)
}

func (w *SessionWatchdog) Shutdown() {
	w.log.Debug("sessionWatchdog.Shutdown()")
	w.ctx.requestShutdown()
}

func (w *SessionWatchdog) CloseOrEnqueueForClosing(session *Session, closeGracePeriod time.Duration) {
	w.ctx.closeOrEnqueueForClosing(session, closeGracePeriod)
}

func (w *SessionWatchdog) DequeueFromClosing(session *Session) {
	w.ctx.dequeueFromClosing(session)
}

func (w *SessionWatchdog) AddToSplitByTimeout(session *SessionProxy) {
	w.ctx.addToSplitByTimeout(session)
}
func (w *SessionWatchdog) RemoveFromSplitByTimeout(session *SessionProxy) {
	w.ctx.removeFromSplitByTimeout(session)
}
