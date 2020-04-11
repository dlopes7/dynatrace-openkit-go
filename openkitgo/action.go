package openkitgo

import (
	log "github.com/sirupsen/logrus"
)

type Action interface {
	// TODO - Implement these
	// ReportEvent(string)
	// ReportIntValue(string, int)
	// ReportDoubleValue(string, float64)
	// ReportStringValue(string, string)
	// ReportError(string, int, string)
	// TraceWebRequest(string)
	EnterAction(string) Action
	LeaveAction()
	LeaveActionAt(int)
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
}

type rootAction struct {
	action           *action
	openChildActions map[int]Action
}

// 	return NewAction(s.log, s.beacon, actionName, s.openRootActions)
func newAction(log *log.Logger, beacon *Beacon, actionName string, parentAction *action, thisLevelActions map[int]Action) *action {
	a := new(action)

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

	return a
}

func newRootAction(log *log.Logger, beacon *Beacon, actionName string, openChildActions map[int]Action) Action {

	a := new(rootAction)

	a.openChildActions = openChildActions
	a.action = newAction(log, beacon, actionName, nil, openChildActions)

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

func (a *rootAction) LeaveActionAt(endTime int) {
	a.action.log.Debug("RootAction.leaveActionAt()")

	for len(a.openChildActions) > 0 {
		for _, child := range a.openChildActions {
			child.LeaveActionAt(endTime)
		}
	}

	a.action.LeaveActionAt(endTime)

}

func (a *rootAction) EnterAction(actionName string) Action {
	a.action.log.Debugf("EnterAction(%s)\n", actionName)

	if a.action.endTime == -1 {
		return newAction(a.action.log, a.action.beacon, actionName, a.action, a.openChildActions)
	}

	return nil
}

func (a *action) LeaveAction() {
	a.log.Debugf("Action(%s).leaveAction()", a.name)

	a.endTime = a.beacon.getCurrentTimestamp()
	a.endSequenceNo = a.beacon.createSequenceNumber()

	a.beacon.addAction(a)

	delete(a.thisLevelActions, a.ID)

}
func (a *action) LeaveActionAt(endTime int) {
	a.log.Debugf("Action(%s).LeaveActionAt()", a.name)

	a.endTime = endTime
	a.endSequenceNo = a.beacon.createSequenceNumber()

	a.beacon.addAction(a)

	delete(a.thisLevelActions, a.ID)

}

func (a *action) EnterAction(actionName string) Action {
	return nil

}

func (a *action) getParentActionID() int {

	if a.parentAction == nil {
		return 0
	}

	return a.parentAction.ID

}
