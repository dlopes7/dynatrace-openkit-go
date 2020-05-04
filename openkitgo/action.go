package openkitgo

import (
	"time"
)

type Action interface {
	// TODO - Implement these
	// ReportEvent(string)
	// ReportIntValue(string, int)
	// ReportDoubleValue(string, float64)
	// ReportStringValue(string, string)
	ReportStringValueAt(string, string, time.Time)
	// ReportError(string, int, string)
	TraceWebRequest(string) *WebRequestTracer
	TraceWebRequestAt(string, time.Time) *WebRequestTracer
	EnterAction(string) Action
	EnterActionAt(string, time.Time) Action
	LeaveAction()
	LeaveActionAt(time.Time)

	getName() string
	getParentActionID() int
	getStartTime() time.Time
	getEndTime() time.Time
}
