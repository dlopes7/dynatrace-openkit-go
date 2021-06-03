package core

import (
	"time"
)

type NullAction struct{}

func NewNullAction() NullAction {
	return NullAction{}
}

func (a NullAction) ReportEvent(eventName string) IAction {
	return a
}

func (a NullAction) ReportEventAt(eventName string, timestamp time.Time) IAction {
	return a
}

func (a NullAction) ReportInt64Value(valueName string, value int64) IAction {
	return a
}

func (a NullAction) ReportInt64ValueAt(valueName string, value int64, timestamp time.Time) IAction {
	return a
}

func (a NullAction) ReportStringValue(valueName string, value string) IAction {
	return a
}

func (a NullAction) ReportStringValueAt(valueName string, value string, timestamp time.Time) IAction {
	return a
}

func (a NullAction) ReportFloat64Value(valueName string, value float64) IAction {
	return a
}

func (a NullAction) ReportFloat64ValueAt(valueName string, value float64, timestamp time.Time) IAction {
	return a
}

func (a NullAction) ReportError(errorName string, errorCode int) IAction {
	return a
}

func (a NullAction) ReportErrorAt(errorName string, errorCode int, timestamp time.Time) IAction {
	return a
}

func (a NullAction) ReportException(errorName string, causeName string, causeDescription string, causeStack string) IAction {
	return a
}

func (a NullAction) ReportExceptionAt(errorName string, causeName string, causeDescription string, causeStack string, timestamp time.Time) IAction {
	return a
}

func (a NullAction) LeaveAction() IAction {
	return a
}

func (a NullAction) LeaveActionAt(timestamp time.Time) IAction {
	return a
}

func (a NullAction) CancelAction() IAction {
	return a
}

func (a NullAction) CancelActionAt(timestamp time.Time) IAction {
	panic("implement me")
}

func (a NullAction) GetDuration() time.Duration {
	panic("implement me")
}
