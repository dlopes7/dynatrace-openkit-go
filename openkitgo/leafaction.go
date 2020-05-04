package openkitgo

import (
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

type LeafAction struct {
	log             *log.Logger
	parent          *RootAction
	children        []OpenKitObject
	parentActionID  int
	lock            sync.Mutex
	id              int
	name            string
	startTime       time.Time
	endTime         time.Time
	startSequenceNo int
	endSequenceNo   int
	isActionLeft    bool
	beacon          *Beacon
}

func (l *LeafAction) getName() string {
	return l.name
}

func (l *LeafAction) getParentActionID() int {
	return l.parentActionID
}

func (l *LeafAction) getStartTime() time.Time {
	return l.startTime
}

func (l *LeafAction) getEndTime() time.Time {
	return l.endTime
}

func newLeafAction(log *log.Logger, parent *RootAction, name string, beacon *Beacon) *LeafAction {
	l := new(LeafAction)
	l.log = log
	l.parent = parent
	l.parentActionID = parent.getActionID()
	l.id = beacon.CreateID()
	l.name = name
	l.startTime = time.Now()
	l.startSequenceNo = beacon.CreateSequenceNumber()
	l.isActionLeft = false
	l.beacon = beacon
	return l
}

func newLeafActionAt(log *log.Logger, parent *RootAction, name string, beacon *Beacon, timestamp time.Time) *LeafAction {
	l := new(LeafAction)
	l.log = log
	l.parent = parent
	l.parentActionID = parent.getActionID()
	l.id = beacon.CreateID()
	l.name = name
	l.startTime = timestamp
	l.startSequenceNo = beacon.CreateSequenceNumber()
	l.isActionLeft = false
	l.beacon = beacon
	return l
}

func (l *LeafAction) getActionID() int {
	return l.id
}

func (l *LeafAction) ReportStringValueAt(name string, value string, timestamp time.Time) {
	l.log.WithFields(log.Fields{"name": name, "value": value, "timestamp": timestamp}).Debugf("LeafAction.ReportStringValueAt")
	l.lock.Lock()
	if !l.isActionLeft {
		l.beacon.ReportValueAt(l.id, name, value, timestamp)
	}
	l.lock.Unlock()
}

func (l *LeafAction) TraceWebRequest(url string) *WebRequestTracer {
	l.lock.Lock()
	w := NewWebRequestTracer(l.log, l, url, l.beacon)
	l.lock.Unlock()
	return w
}

func (l *LeafAction) TraceWebRequestAt(url string, timestamp time.Time) *WebRequestTracer {
	l.log.WithFields(log.Fields{"url": url, "timestamp": timestamp}).Debugf("LeafAction.TraceWebRequestAt")

	l.lock.Lock()
	w := NewWebRequestTracerAt(l.log, l, url, l.beacon, timestamp)
	l.lock.Unlock()
	return w
}

func (l *LeafAction) LeaveActionAt(timestamp time.Time) {
	l.log.Debugf("LeafAction.LeaveAction")

	for _, c := range l.getCopyOfChildObjects() {
		c.close()
	}

	l.endTime = timestamp
	l.endSequenceNo = l.beacon.CreateSequenceNumber()

	l.beacon.AddLeafAction(l)
	l.parent.onChildClosed(l)
	l.parent = nil
}

func (l *LeafAction) LeaveAction() {
	l.log.Debugf("LeafAction.LeaveAction")

	for _, c := range l.getCopyOfChildObjects() {
		c.close()
	}

	l.endTime = time.Now()
	l.endSequenceNo = l.beacon.CreateSequenceNumber()

	l.beacon.AddLeafAction(l)
	l.parent.onChildClosed(l)
	l.parent = nil
}

func (l *LeafAction) storeChildInList(child OpenKitObject) {
	l.children = append(l.children, child)
}

func (l *LeafAction) removeChildFromList(child OpenKitObject) bool {
	tmp := l.children[:0]
	for _, c := range l.children {
		if child != c {
			tmp = append(tmp, c)
		}
	}
	l.children = tmp
	return true
}

func (l *LeafAction) getCopyOfChildObjects() []OpenKitObject {
	copied := make([]OpenKitObject, len(l.children))
	copy(copied, l.children)
	return copied
}

func (l *LeafAction) close() {
	l.LeaveAction()
}

func (l *LeafAction) onChildClosed(child OpenKitObject) {
	l.lock.Lock()
	l.removeChildFromList(child)
	l.lock.Unlock()
}

func (l *LeafAction) EnterAction(name string) Action {
	return l
}

func (l *LeafAction) EnterActionAt(s string, t time.Time) Action {
	return l
}

func (l *LeafAction) getParentAction() *RootAction {
	return l.parent
}

func (l *LeafAction) getChildCount() int {
	return 0
}
