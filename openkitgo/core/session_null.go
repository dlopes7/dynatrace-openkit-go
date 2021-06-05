package core

import (
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo"
	"time"
)

type NullSession struct {
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

func NewNullSession() openkitgo.Session {
	return &NullSession{}
}

func (n NullSession) EnterAction(actionName string) openkitgo.Action {
	return NewNullAction()
}

func (n NullSession) EnterActionAt(actionName string, timestamp time.Time) openkitgo.Action {
	return NewNullAction()
}

func (n NullSession) IdentifyUser(userTag string) {}

func (n NullSession) IdentifyUserAt(userTag string, timestamp time.Time) {}

func (n NullSession) ReportCrash(errorName string, reason string, stacktrace string) {}

func (n NullSession) ReportCrashAt(errorName string, reason string, stacktrace string, timestamp time.Time) {
}

func (n NullSession) End() {}

func (n NullSession) String() string {
	panic("NullSession")
}
