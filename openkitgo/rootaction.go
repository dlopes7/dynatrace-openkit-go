package openkitgo

import (
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

type RootAction struct {
	log             *log.Logger
	parent          OpenKitComposite
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

func newRootAction(log *log.Logger, parent OpenKitComposite, name string, beacon *Beacon) *RootAction {
	r := new(RootAction)
	r.log = log
	r.parent = parent
	r.parentActionID = parent.getActionID()
	r.id = beacon.CreateID()
	r.name = name
	r.startTime = time.Now()
	r.startSequenceNo = beacon.CreateSequenceNumber()
	r.isActionLeft = false
	r.beacon = beacon
	return r
}

func newRootActionAt(log *log.Logger, parent OpenKitComposite, name string, beacon *Beacon, timestamp time.Time) *RootAction {
	r := new(RootAction)
	r.log = log
	r.parent = parent
	r.parentActionID = parent.getActionID()
	r.id = beacon.CreateID()
	r.name = name
	r.startTime = timestamp
	r.startSequenceNo = beacon.CreateSequenceNumber()
	r.isActionLeft = false
	r.beacon = beacon
	return r
}

func (r *RootAction) getActionID() int {
	return r.id
}

func (r *RootAction) ReportStringValueAt(name string, value string, timestamp time.Time) {
	r.log.WithFields(log.Fields{"name": name, "value": value, "timestamp": timestamp}).Debugf("RootAction.ReportStringValueAt")
	r.lock.Lock()
	if !r.isActionLeft {
		r.beacon.ReportValueAt(r.id, name, value, timestamp)
	}
	r.lock.Unlock()
}

func (r *RootAction) TraceWebRequest(url string) *WebRequestTracer {
	r.lock.Lock()
	w := NewWebRequestTracer(r.log, r, url, r.beacon)
	r.lock.Unlock()
	return w
}

func (r *RootAction) TraceWebRequestAt(url string, timestamp time.Time) *WebRequestTracer {
	r.log.WithFields(log.Fields{"url": url, "timestamp": timestamp}).Debugf("RootAction.TraceWebRequestAt")

	r.lock.Lock()
	w := NewWebRequestTracerAt(r.log, r, url, r.beacon, timestamp)
	r.lock.Unlock()
	return w
}

func (r *RootAction) LeaveActionAt(timestamp time.Time) {
	r.log.Debugf("RootAction.LeaveAction")

	for _, c := range r.getCopyOfChildObjects() {
		c.closeAt(timestamp)
	}

	r.endTime = timestamp
	r.endSequenceNo = r.beacon.CreateSequenceNumber()

	r.beacon.AddRootAction(r)
	r.parent.onChildClosed(r)
	r.parent = nil
}

func (r *RootAction) LeaveAction() {
	r.log.Debugf("RootAction.LeaveAction")

	for _, c := range r.getCopyOfChildObjects() {
		c.close()
	}

	r.endTime = time.Now()
	r.endSequenceNo = r.beacon.CreateSequenceNumber()

	r.beacon.AddRootAction(r)
	r.parent.onChildClosed(r)
	r.parent = nil
}

func (r *RootAction) storeChildInList(child OpenKitObject) {
	r.children = append(r.children, child)
}

func (r *RootAction) removeChildFromList(child OpenKitObject) bool {
	tmp := make([]OpenKitObject, 0)
	for _, c := range r.children {
		if child != c {
			tmp = append(tmp, c)
		}
	}
	r.children = tmp
	return true
}

func (r *RootAction) getCopyOfChildObjects() []OpenKitObject {
	copied := make([]OpenKitObject, len(r.children))
	copy(copied, r.children)
	return copied
}

func (r *RootAction) close() {
	r.LeaveAction()
}

func (r *RootAction) closeAt(timestamp time.Time) {
	r.LeaveActionAt(timestamp)
}

func (r *RootAction) onChildClosed(child OpenKitObject) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.removeChildFromList(child)

}

func (r *RootAction) EnterAction(name string) Action {
	r.log.WithFields(log.Fields{"name": name}).Debugf("RootAction.EnterAction")
	r.lock.Lock()
	defer r.lock.Unlock()
	if !r.isActionLeft {
		child := newLeafAction(r.log, r, name, r.beacon)
		r.storeChildInList(child)
		return child
	}
	return NullAction{parent: r}
}

func (r *RootAction) EnterActionAt(name string, timestamp time.Time) Action {
	r.log.WithFields(log.Fields{"name": name}).Debugf("RootAction.EnterAction")
	r.lock.Lock()
	defer r.lock.Unlock()
	if !r.isActionLeft {
		child := newLeafActionAt(r.log, r, name, r.beacon, timestamp)
		r.storeChildInList(child)
		return child
	}

	return NullAction{parent: r}
}

func (r *RootAction) getName() string {
	return r.name
}

func (r *RootAction) getParentActionID() int {
	return 0
}

func (r *RootAction) getStartTime() time.Time {
	return r.startTime
}

func (r *RootAction) getEndTime() time.Time {
	return r.endTime
}

func (r *RootAction) getChildCount() int {
	r.lock.Lock()
	defer r.lock.Unlock()
	return len(r.children)
}
