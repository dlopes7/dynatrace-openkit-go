package openkitgo

import (
	"github.com/op/go-logging"
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
}

type action struct {

	// These belong to all Actions
	logger *logging.Logger

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

// 	return NewAction(s.logger, s.beacon, actionName, s.openRootActions)
func newAction(logger *logging.Logger, beacon *Beacon, actionName string, parentAction *action, thisLevelActions map[int]Action) *action {
	a := new(action)

	a.logger = logger
	a.beacon = beacon
	a.name = actionName
	a.parentAction = parentAction
	a.thisLevelActions = thisLevelActions

	a.startTime = beacon.getCurrentTimestamp()
	a.endTime = -1
	a.startSequenceNo = beacon.createSequenceNumber()
	a.ID = beacon.createID()

	if parentAction != nil {
		a.thisLevelActions[a.ID] = a
	}

	return a
}

func newRootAction(logger *logging.Logger, beacon *Beacon, actionName string, openChildActions map[int]Action) Action {

	a := new(rootAction)

	a.openChildActions = openChildActions
	a.action = newAction(logger, beacon, actionName, nil, openChildActions)

	return a

}

func (a *rootAction) LeaveAction() {
	a.action.logger.Debug("RootAction.leaveAction()")

	for len(a.openChildActions) > 0 {
		for _, child := range a.openChildActions {
			child.LeaveAction()
		}
	}

	a.action.LeaveAction()

}

func (a *rootAction) EnterAction(actionName string) Action {
	a.action.logger.Debugf("EnterAction(%s)\n", actionName)

	if a.action.endTime == -1 {
		return newAction(a.action.logger, a.action.beacon, actionName, a.action, a.openChildActions)
	}

	return nil
}

func (a *action) LeaveAction() {
	a.logger.Debugf("Action(%s).leaveAction()", a.name)

	a.endTime = a.beacon.getCurrentTimestamp()
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
