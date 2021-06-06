package openkitgo

import (
	"time"
)

type Session interface {
	EnterAction(actionName string) Action
	EnterActionAt(actionName string, timestamp time.Time) Action

	IdentifyUser(userTag string)
	IdentifyUserAt(userTag string, timestamp time.Time)

	ReportCrash(errorName string, reason string, stacktrace string)
	ReportCrashAt(errorName string, reason string, stacktrace string, timestamp time.Time)

	TraceWebRequest(url string) WebRequestTracer
	TraceWebRequestAt(url string, timestamp time.Time) WebRequestTracer

	End()
	EndAt(timestamp time.Time)

	String() string
}
