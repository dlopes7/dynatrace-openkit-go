package core

import (
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo"
	"time"
)

type NullAction struct{}

func (a NullAction) TraceWebRequest(url string) openkitgo.WebRequestTracer {
	return NewNullWebRequestTracer()
}

func (a NullAction) TraceWebRequestAt(url string, timestamp time.Time) openkitgo.WebRequestTracer {
	return NewNullWebRequestTracer()
}

func NewNullAction() NullAction {
	return NullAction{}
}

func (a NullAction) ReportEvent(eventName string) openkitgo.Action {
	return a
}

func (a NullAction) ReportEventAt(eventName string, timestamp time.Time) openkitgo.Action {
	return a
}

func (a NullAction) ReportValue(valueName string, value interface{}) openkitgo.Action {
	return a
}

func (a NullAction) ReportValueAt(valueName string, value interface{}, timestamp time.Time) openkitgo.Action {
	return a
}

func (a NullAction) ReportError(errorName string, causeName string, causeDescription string, causeStack string) openkitgo.Action {
	return a
}

func (a NullAction) ReportErrorAt(errorName string, causeName string, causeDescription string, causeStack string, timestamp time.Time) openkitgo.Action {
	return a
}

func (a NullAction) LeaveAction() openkitgo.Action {
	return a
}

func (a NullAction) LeaveActionAt(timestamp time.Time) openkitgo.Action {
	return a
}

func (a NullAction) CancelAction() openkitgo.Action {
	return a
}

func (a NullAction) CancelActionAt(timestamp time.Time) openkitgo.Action {
	return a
}

func (a NullAction) GetDuration() time.Duration {
	return 0
}
