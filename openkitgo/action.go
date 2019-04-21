package openkitgo

import "github.com/op/go-logging"

type Action interface {
	// ReportEvent(string)
	// ReportIntValue(string, int)
	// ReportDoubleValue(string, float64)
	// ReportStringValue(string, string)
	// ReportError(string, int, string)
	// TraceWebRequest(string)
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

	// This is only for RootAction
	openChildActions map[int]Action
}

type rootAction struct {
	action *action
}

// 	return NewAction(s.logger, s.beacon, actionName, s.openRootActions)
func NewAction(logger *logging.Logger, beacon *Beacon, actionName string, openChildActions map[int]Action) Action {
	a := new(action)

	a.logger = logger
	a.beacon = beacon
	a.name = actionName

	a.startTime = beacon.getCurrentTimestamp()
	a.startSequenceNo = beacon.createSequenceNumber()
	a.ID = beacon.createID()

	a.thisLevelActions = make(map[int]Action)
	a.openChildActions = openChildActions

	return a
}

func (a *action) LeaveAction() {
	a.logger.Debug("Action.leaveAction()")

	a.endTime = a.beacon.getCurrentTimestamp()
	a.endSequenceNo = a.beacon.createSequenceNumber()

	a.beacon.addAction(a)

	delete(a.thisLevelActions, a.ID)

}

func (a *action) getParentActionID() int {

	if a.parentAction == nil {
		return 0
	}

	return a.parentAction.ID

}
