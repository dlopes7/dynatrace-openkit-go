package openkitgo

import "time"

type IAction interface {
	ReportEvent(eventName string) IAction
	ReportEventAt(eventName string, timestamp time.Time) IAction

	ReportInt64Value(valueName string, value int64) IAction
	ReportInt64ValueAt(valueName string, value int64, timestamp time.Time) IAction

	ReportStringValue(valueName string, value string) IAction
	ReportStringValueAt(valueName string, value string, timestamp time.Time) IAction

	ReportFloat64Value(valueName string, value float64) IAction
	ReportFloat64ValueAt(valueName string, value float64, timestamp time.Time) IAction

	ReportError(errorName string, errorCode int) IAction
	ReportErrorAt(errorName string, errorCode int, timestamp time.Time) IAction

	ReportException(errorName string, causeName string, causeDescription string, causeStack string) IAction
	ReportExceptionAt(errorName string, causeName string, causeDescription string, causeStack string, timestamp time.Time) IAction

	// TODO TraceWebRequest()
	// TODO TraceWebRequestAt()

	LeaveAction() IAction
	LeaveActionAt(timestamp time.Time) IAction

	CancelAction() IAction
	CancelActionAt(timestamp time.Time) IAction

	GetDuration() time.Duration
}
