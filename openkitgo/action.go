package openkitgo

import (
	log "github.com/sirupsen/logrus"
	"sync"
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
}

type action struct {

	// These belong to all Actions
	log *log.Logger

	ID   int
	name string

	parentAction *action

	startTime       int
	endTime         int
	startSequenceNo int
	endSequenceNo   int

	beacon *Beacon

	thisLevelActions map[int]Action
	lock             sync.Mutex
}

type rootAction struct {
	action           *action
	openChildActions map[int]Action
}

// 	return NewAction(s.log, s.beacon, actionName, s.openRootActions)
func newAction(log *log.Logger, beacon *Beacon, actionName string, parentAction *action, thisLevelActions map[int]Action) *action {
	a := new(action)

	a.lock.Lock()
	a.log = log
	a.beacon = beacon
	a.name = actionName
	a.parentAction = parentAction

	a.thisLevelActions = thisLevelActions

	a.startTime = beacon.getCurrentTimestamp()
	a.endTime = -1
	a.startSequenceNo = beacon.createSequenceNumber()
	a.ID = beacon.createID()
	a.thisLevelActions[a.ID] = a
	a.lock.Unlock()

	return a
}

func newActionAt(log *log.Logger, beacon *Beacon, actionName string, parentAction *action, thisLevelActions map[int]Action, timestamp time.Time) *action {
	a := new(action)

	a.lock.Lock()
	a.log = log
	a.beacon = beacon
	a.name = actionName
	a.parentAction = parentAction
	a.thisLevelActions = thisLevelActions

	a.startTime = TimeToMillis(timestamp)
	a.endTime = -1
	a.startSequenceNo = beacon.createSequenceNumber()
	a.ID = beacon.createID()

	a.thisLevelActions[a.ID] = a
	a.lock.Unlock()

	return a
}

func newRootAction(log *log.Logger, beacon *Beacon, actionName string, openChildActions map[int]Action) Action {

	a := new(rootAction)

	a.openChildActions = openChildActions
	a.action = newAction(log, beacon, actionName, nil, openChildActions)

	return a

}

func newRootActionAt(log *log.Logger, beacon *Beacon, actionName string, openChildActions map[int]Action, timestamp time.Time) Action {

	a := new(rootAction)

	a.openChildActions = openChildActions
	a.action = newActionAt(log, beacon, actionName, nil, openChildActions, timestamp)

	return a

}

func (a *rootAction) LeaveAction() {
	a.action.log.Debugf("RootAction.leaveAction()")

	for len(a.openChildActions) > 0 {
		for _, child := range a.openChildActions {
			child.LeaveAction()
		}
	}
	a.action.LeaveAction()
}

func (a *rootAction) LeaveActionAt(endTime time.Time) {
	a.action.log.Debugf("RootAction.LeaveActionAt(%s)", endTime.String())

	for len(a.openChildActions) > 0 {
		for _, child := range a.openChildActions {
			child.LeaveActionAt(endTime)
		}
	}

	a.action.LeaveActionAt(endTime)
}

func (a *rootAction) EnterAction(actionName string) Action {
	a.action.log.Debugf("EnterAction(%s)", actionName)

	if a.action.endTime == -1 {
		return newAction(a.action.log, a.action.beacon, actionName, a.action, a.openChildActions)
	}

	return nil
}

func (a *rootAction) EnterActionAt(actionName string, timestamp time.Time) Action {
	a.action.log.Debugf("EnterActionAt(%s, %s)", actionName, timestamp)
	return newActionAt(a.action.log, a.action.beacon, actionName, a.action, a.openChildActions, timestamp)

}

func (a *rootAction) getActionID() int {
	return a.action.ID
}

func (a *rootAction) TraceWebRequest(url string) *WebRequestTracer {
	w := NewWebRequestTracer(a.action.log, a, url, a.action.beacon)
	return w
}

func (a *rootAction) TraceWebRequestAt(url string, timestamp time.Time) *WebRequestTracer {
	w := NewWebRequestTracerAt(a.action.log, a, url, a.action.beacon, timestamp)
	return w
}

func (a *rootAction) ReportStringValueAt(key string, value string, timestamp time.Time) {
	a.action.log.WithFields(log.Fields{"key": key, "value": value, "timestamp": timestamp}).Debugf("ReportStringValueAt")
	a.action.beacon.reportValueAt(a.action.ID, key, value, timestamp)
}

func (a *action) LeaveAction() {
	a.lock.Lock()
	a.log.Debugf("Action(%s).leaveAction()", a.name)

	a.endTime = a.beacon.getCurrentTimestamp()
	a.endSequenceNo = a.beacon.createSequenceNumber()

	a.beacon.addAction(a)
	delete(a.thisLevelActions, a.ID)
	a.lock.Unlock()

}
func (a *action) LeaveActionAt(endTime time.Time) {
	a.lock.Lock()
	a.log.Debugf("Action(%s).LeaveActionAt(%s)", a.name, endTime.String())
	a.endTime = TimeToMillis(endTime)
	a.endSequenceNo = a.beacon.createSequenceNumber()
	a.beacon.addAction(a)
	delete(a.thisLevelActions, a.ID)
	a.lock.Unlock()

}

func (a *action) EnterAction(actionName string) Action {
	return nil
}
func (a *action) EnterActionAt(actionName string, timestamp time.Time) Action {
	return nil
}

func (a *action) getParentActionID() int {

	if a.parentAction == nil {
		return 0
	}

	return a.parentAction.ID

}

func (a *action) getActionID() int {
	return a.ID
}

func (a *action) TraceWebRequest(url string) *WebRequestTracer {
	w := NewWebRequestTracer(a.log, a, url, a.beacon)
	return w
}

func (a *action) TraceWebRequestAt(url string, timestamp time.Time) *WebRequestTracer {
	w := NewWebRequestTracerAt(a.log, a, url, a.beacon, timestamp)
	return w
}

func (a *action) ReportStringValueAt(key string, value string, timestamp time.Time) {
	a.log.WithFields(log.Fields{"key": key, "value": value, "timestamp": timestamp}).Debugf("ReportStringValueAt")
	a.beacon.reportValueAt(a.ID, key, value, timestamp)
}
