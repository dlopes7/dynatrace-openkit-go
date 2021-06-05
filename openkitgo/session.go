package openkitgo

import "time"

type ISession interface {
	EnterAction(actionName string) IAction
	EnterActionAt(actionName string, timestamp time.Time) IAction

	IdentifyUser(userTag string)
	IdentifyUserAt(userTag string, timestamp time.Time)

	ReportCrash(errorName string, reason string, stacktrace string)
	ReportCrashAt(errorName string, reason string, stacktrace string, timestamp time.Time)

	// TODO TraceWebRequest()
	// TODO TraceWebRequestAt()

	End()

	String() string
}
