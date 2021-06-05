package openkitgo

import "time"

type Action interface {
	ReportEvent(eventName string) Action
	ReportEventAt(eventName string, timestamp time.Time) Action

	ReportInt64Value(valueName string, value int64) Action
	ReportInt64ValueAt(valueName string, value int64, timestamp time.Time) Action

	ReportStringValue(valueName string, value string) Action
	ReportStringValueAt(valueName string, value string, timestamp time.Time) Action

	ReportFloat64Value(valueName string, value float64) Action
	ReportFloat64ValueAt(valueName string, value float64, timestamp time.Time) Action

	ReportError(errorName string, errorCode int) Action
	ReportErrorAt(errorName string, errorCode int, timestamp time.Time) Action

	ReportException(errorName string, causeName string, causeDescription string, causeStack string) Action
	ReportExceptionAt(errorName string, causeName string, causeDescription string, causeStack string, timestamp time.Time) Action

	// TODO TraceWebRequest()
	// TODO TraceWebRequestAt()

	LeaveAction() Action
	LeaveActionAt(timestamp time.Time) Action

	CancelAction() Action
	CancelActionAt(timestamp time.Time) Action

	GetDuration() time.Duration
}
