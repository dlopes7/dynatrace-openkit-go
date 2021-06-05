package core

import (
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo"
	"time"
)

type NullAction struct{}

func NewNullAction() NullAction {
	return NullAction{}
}

func (a NullAction) ReportEvent(eventName string) openkitgo.Action {
	return a
}

func (a NullAction) ReportEventAt(eventName string, timestamp time.Time) openkitgo.Action {
	return a
}

func (a NullAction) ReportInt64Value(valueName string, value int64) openkitgo.Action {
	return a
}

func (a NullAction) ReportInt64ValueAt(valueName string, value int64, timestamp time.Time) openkitgo.Action {
	return a
}

func (a NullAction) ReportStringValue(valueName string, value string) openkitgo.Action {
	return a
}

func (a NullAction) ReportStringValueAt(valueName string, value string, timestamp time.Time) openkitgo.Action {
	return a
}

func (a NullAction) ReportFloat64Value(valueName string, value float64) openkitgo.Action {
	return a
}

func (a NullAction) ReportFloat64ValueAt(valueName string, value float64, timestamp time.Time) openkitgo.Action {
	return a
}

func (a NullAction) ReportError(errorName string, errorCode int) openkitgo.Action {
	return a
}

func (a NullAction) ReportErrorAt(errorName string, errorCode int, timestamp time.Time) openkitgo.Action {
	return a
}

func (a NullAction) ReportException(errorName string, causeName string, causeDescription string, causeStack string) openkitgo.Action {
	return a
}

func (a NullAction) ReportExceptionAt(errorName string, causeName string, causeDescription string, causeStack string, timestamp time.Time) openkitgo.Action {
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
