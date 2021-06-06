package core

import (
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo/interfaces"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

type Action struct {
	log             *log.Logger
	parent          OpenKitComposite
	parentAction    interfaces.Action
	parentActionID  int
	mutex           sync.RWMutex
	id              int32
	name            string
	startTime       time.Time
	endTime         time.Time
	startSequenceNo int
	endSequenceNo   int32
	actionLeft      bool
	beacon          *Beacon
	children        []OpenKitObject
}

func (a *Action) storeChildInList(child OpenKitObject) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.children = append(a.children, child)
}

func (a *Action) removeChildFromList(child OpenKitObject) bool {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	removed := false

	var keep []OpenKitObject
	for _, c := range a.children {
		if c != child {
			keep = append(keep, c)
		} else {
			removed = true
		}
	}
	a.children = keep
	return removed
}

func (a *Action) getCopyOfChildObjects() []OpenKitObject {
	return a.children[:]
}

func (a *Action) getChildCount() int {
	return len(a.children)
}

func (a *Action) onChildClosed(child OpenKitObject) {
	a.removeChildFromList(child)
}

func (a *Action) getActionID() int {
	return int(a.id)
}

func NewAction(log *log.Logger, parent OpenKitComposite, parentAction interfaces.Action, name string, beacon *Beacon, startTime time.Time) *Action {

	return &Action{
		log:             log,
		parent:          parent,
		parentAction:    parentAction,
		parentActionID:  parent.getActionID(),
		id:              beacon.CreateID(),
		name:            name,
		startTime:       startTime,
		startSequenceNo: int(beacon.CreateSequenceNumber()),
		endSequenceNo:   -1,
		actionLeft:      false,
		beacon:          beacon,
	}

}

func (a *Action) closeAt(timestamp time.Time) {
	a.LeaveActionAt(timestamp)
}

func (a *Action) close() {
	a.closeAt(time.Now())
}

func (a *Action) ReportEvent(eventName string) interfaces.Action {
	return a.ReportEventAt(eventName, time.Now())
}

func (a *Action) ReportEventAt(eventName string, timestamp time.Time) interfaces.Action {
	if eventName == "" {
		a.log.Warning("eventName must not be empty")
		return a
	}

	a.log.WithFields(log.Fields{"actionName": a.name, "eventName": eventName, "timestamp": timestamp}).Debug("ReportEvent()")

	if !a.actionLeft {
		a.beacon.reportEvent(int(a.id), eventName, timestamp)
	}

	return a
}

func (a *Action) ReportValue(valueName string, value interface{}) interfaces.Action {
	return a.ReportValueAt(valueName, value, time.Now())
}

func (a *Action) ReportValueAt(valueName string, value interface{}, timestamp time.Time) interfaces.Action {
	a.log.WithFields(log.Fields{"actionName": a.name, "valueName": valueName, "value": value, "timestamp": timestamp}).Debug("ReportValue()")
	if !a.actionLeft {
		a.beacon.reportValue(int(a.id), valueName, value, timestamp)
	}
	return a
}

func (a *Action) ReportError(errorName string, causeName string, causeDescription string, causeStack string) interfaces.Action {
	return a.ReportErrorAt(errorName, causeName, causeDescription, causeStack, time.Now())
}

func (a *Action) ReportErrorAt(errorName string, causeName string, causeDescription string, causeStack string, timestamp time.Time) interfaces.Action {
	a.log.WithFields(log.Fields{"actionName": a.name, "errorName": errorName, "causeName": causeName, "timestamp": timestamp}).Debug("ReportError()")
	if !a.actionLeft {
		a.beacon.reportError(int(a.id), errorName, causeName, causeDescription, causeStack, timestamp)
	}
	return a
}

func (a *Action) LeaveAction() interfaces.Action {
	return a.LeaveActionAt(time.Now())
}

func (a *Action) LeaveActionAt(timestamp time.Time) interfaces.Action {
	a.log.WithFields(log.Fields{"actionName": a.name, "timestamp": timestamp}).Debug("Action.LeaveAction()")

	return a.doLeaveAction(false, timestamp)
}

func (a *Action) CancelAction() interfaces.Action {
	return a.CancelActionAt(time.Now())
}

func (a *Action) CancelActionAt(timestamp time.Time) interfaces.Action {
	a.log.WithFields(log.Fields{"actionName": a.name}).Debug("CancelAction()")

	return a.doLeaveAction(true, timestamp)

}

func (a *Action) doLeaveAction(discardData bool, timestamp time.Time) interfaces.Action {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	if a.actionLeft {
		return a.parentAction
	}
	a.actionLeft = true

	for _, child := range a.getCopyOfChildObjects() {
		if discardData {
			child.(*Action).CancelActionAt(timestamp)
		} else {
			child.closeAt(timestamp)
		}
	}

	a.endTime = timestamp
	a.endSequenceNo = a.beacon.CreateSequenceNumber()

	if !discardData {
		a.beacon.AddAction(a)
	}

	a.parent.onChildClosed(a)
	a.parent = nil

	return a.parentAction
}

func (a *Action) GetDuration() time.Duration {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	if a.actionLeft {
		return a.endTime.Sub(a.startTime)
	}
	return time.Now().Sub(a.startTime)
}

func (a *Action) TraceWebRequest(url string) interfaces.WebRequestTracer {
	return a.TraceWebRequestAt(url, time.Now())
}

func (a *Action) TraceWebRequestAt(url string, timestamp time.Time) interfaces.WebRequestTracer {
	a.log.WithFields(log.Fields{"actionName": a.name, "url": url, "timestamp": timestamp}).Debug("Action.TraceWebRequest()")

	if !a.actionLeft {
		t := NewWebRequestTracer(a.log, a, url, a.beacon, timestamp)
		a.storeChildInList(t)
		return t
	}
	return NewNullWebRequestTracer()
}
