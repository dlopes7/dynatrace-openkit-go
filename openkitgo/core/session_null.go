package core

import (
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo/interfaces"
	"time"
)

type NullSession struct {
}

func (n NullSession) TraceWebRequest(url string) interfaces.WebRequestTracer {
	return NewNullWebRequestTracer()
}

func (n NullSession) TraceWebRequestAt(url string, timestamp time.Time) interfaces.WebRequestTracer {
	return NewNullWebRequestTracer()
}

func (n NullSession) storeChildInList(child OpenKitObject) {

}

func (n NullSession) removeChildFromList(child OpenKitObject) bool {
	return false
}

func (n NullSession) getCopyOfChildObjects() []OpenKitObject {
	return []OpenKitObject{}
}

func (n NullSession) getChildCount() int {
	return 0
}

func (n NullSession) onChildClosed(child OpenKitObject) {

}

func (n NullSession) getActionID() int {
	return 0
}

func NewNullSession() interfaces.Session {
	return &NullSession{}
}

func (n NullSession) EnterAction(actionName string) interfaces.Action {
	return NewNullAction()
}

func (n NullSession) EnterActionAt(actionName string, timestamp time.Time) interfaces.Action {
	return NewNullAction()
}

func (n NullSession) IdentifyUser(userTag string)                                    {}
func (n NullSession) IdentifyUserAt(userTag string, timestamp time.Time)             {}
func (n NullSession) ReportCrash(errorName string, reason string, stacktrace string) {}
func (n NullSession) ReportCrashAt(errorName string, reason string, stacktrace string, timestamp time.Time) {
}
func (n NullSession) End()                      {}
func (n NullSession) EndAt(timestamp time.Time) {}
func (n NullSession) String() string            { return "NullSession" }
