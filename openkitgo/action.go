package openkitgo

type Action interface {
	ReportEvent(string)
	ReportIntValue(string, int)
	ReportDoubleValue(string, float64)
	ReportStringValue(string, string)
	ReportError(string, int, string)
	TraceWebRequest(string)
	LeaveAction()
}
