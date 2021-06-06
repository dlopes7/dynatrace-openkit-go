package core

import (
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo/interfaces"
	"time"
)

type NullAction struct{}

func (a NullAction) TraceWebRequest(url string) interfaces.WebRequestTracer {
	return NewNullWebRequestTracer()
}

func (a NullAction) TraceWebRequestAt(url string, timestamp time.Time) interfaces.WebRequestTracer {
	return NewNullWebRequestTracer()
}

func NewNullAction() NullAction {
	return NullAction{}
}

func (a NullAction) ReportEvent(eventName string) interfaces.Action {
	return a
}

func (a NullAction) ReportEventAt(eventName string, timestamp time.Time) interfaces.Action {
	return a
}

func (a NullAction) ReportValue(valueName string, value interface{}) interfaces.Action {
	return a
}

func (a NullAction) ReportValueAt(valueName string, value interface{}, timestamp time.Time) interfaces.Action {
	return a
}

func (a NullAction) ReportError(errorName string, causeName string, causeDescription string, causeStack string) interfaces.Action {
	return a
}

func (a NullAction) ReportErrorAt(errorName string, causeName string, causeDescription string, causeStack string, timestamp time.Time) interfaces.Action {
	return a
}

func (a NullAction) LeaveAction() interfaces.Action {
	return a
}

func (a NullAction) LeaveActionAt(timestamp time.Time) interfaces.Action {
	return a
}

func (a NullAction) CancelAction() interfaces.Action {
	return a
}

func (a NullAction) CancelActionAt(timestamp time.Time) interfaces.Action {
	return a
}

func (a NullAction) GetDuration() time.Duration {
	return 0
}
