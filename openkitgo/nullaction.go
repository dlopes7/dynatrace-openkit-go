package openkitgo

import "time"

type NullAction struct {
	parent *RootAction
}

func (n NullAction) ReportStringValueAt(s string, s2 string, time time.Time) {
}

func (n NullAction) TraceWebRequest(s string) *WebRequestTracer {
	return nil
}

func (n NullAction) TraceWebRequestAt(s string, time time.Time) *WebRequestTracer {
	return nil
}

func (n NullAction) EnterAction(s string) Action {
	return n
}

func (n NullAction) EnterActionAt(s string, time time.Time) Action {
	return n
}

func (n NullAction) LeaveAction() {

}

func (n NullAction) LeaveActionAt(time time.Time) {

}

func (n NullAction) getName() string {
	return ""
}

func (n NullAction) getParentActionID() int {
	return 0
}

func (n NullAction) getStartTime() time.Time {
	return time.Now()
}

func (n NullAction) getEndTime() time.Time {
	return time.Now()
}

type NullRootAction struct {
}

func (n NullRootAction) ReportStringValueAt(s string, s2 string, time time.Time) {
}

func (n NullRootAction) TraceWebRequest(s string) *WebRequestTracer {
	return nil
}

func (n NullRootAction) TraceWebRequestAt(s string, time time.Time) *WebRequestTracer {
	return nil
}

func (n NullRootAction) EnterAction(s string) Action {
	return n
}

func (n NullRootAction) EnterActionAt(s string, time time.Time) Action {
	return n
}

func (n NullRootAction) LeaveAction() {

}

func (n NullRootAction) LeaveActionAt(time time.Time) {

}

func (n NullRootAction) getName() string {
	return ""
}

func (n NullRootAction) getParentActionID() int {
	return 0
}

func (n NullRootAction) getStartTime() time.Time {
	return time.Now()
}

func (n NullRootAction) getEndTime() time.Time {
	return time.Now()
}
