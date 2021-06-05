package openkitgo

import "time"

type Session interface {
	EnterAction(actionName string) Action
	EnterActionAt(actionName string, timestamp time.Time) Action

	IdentifyUser(userTag string)
	IdentifyUserAt(userTag string, timestamp time.Time)

	ReportCrash(errorName string, reason string, stacktrace string)
	ReportCrashAt(errorName string, reason string, stacktrace string, timestamp time.Time)

	// TODO TraceWebRequest()
	// TODO TraceWebRequestAt()

	End()
	EndAt(timestamp time.Time)

	String() string
}
