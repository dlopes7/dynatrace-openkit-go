package core

import (
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

type IAction interface {
	ReportEvent(eventName string) IAction
	ReportEventAt(eventName string, timestamp time.Time) IAction

	ReportInt64Value(valueName string, value int64) IAction
	ReportInt64ValueAt(valueName string, value int64, timestamp time.Time) IAction

	ReportStringValue(valueName string, value string) IAction
	ReportStringValueAt(valueName string, value string, timestamp time.Time) IAction

	ReportFloat64Value(valueName string, value float64) IAction
	ReportFloat64ValueAt(valueName string, value float64, timestamp time.Time) IAction

	ReportError(errorName string, errorCode int) IAction
	ReportErrorAt(errorName string, errorCode int, timestamp time.Time) IAction

	ReportException(errorName string, causeName string, causeDescription string, causeStack string) IAction
	ReportExceptionAt(errorName string, causeName string, causeDescription string, causeStack string, timestamp time.Time) IAction

	// TODO TraceWebRequest()
	// TODO TraceWebRequestAt()

	LeaveAction() IAction
	LeaveActionAt(timestamp time.Time) IAction

	CancelAction() IAction
	CancelActionAt(timestamp time.Time) IAction

	GetDuration() time.Duration
}

type Action struct {
	log             *log.Logger
	parent          OpenKitComposite
	parentActionID  int
	mutex           sync.Mutex
	id              int
	name            string
	startTime       time.Time
	endTime         time.Time
	startSequenceNo int
	endSequenceNo   int
	actionLeft      bool
	// TODO beacon *Beacon
}

func NewAction(log *log.Logger, parent OpenKitComposite, name string /* TODO beacon Beacon */, startTime time.Time) Action {

	return Action{
		log:             log,
		parent:          parent,
		parentActionID:  parent.GetActionID(),
		id:              0, // TODO beacon.CreateID(),
		name:            name,
		startTime:       startTime,
		startSequenceNo: 0, // TODO beacon.CreateSequenceNumber()
		endSequenceNo:   -1,
		actionLeft:      false,
		// TODO beacon: beacon
	}

}

func (a *Action) ReportEvent(eventName string) IAction {
	panic("implement me")
}

func (a *Action) ReportEventAt(eventName string, timestamp time.Time) IAction {
	panic("implement me")
}

func (a *Action) ReportInt64Value(valueName string, value int64) IAction {
	panic("implement me")
}

func (a *Action) ReportInt64ValueAt(valueName string, value int64, timestamp time.Time) IAction {
	panic("implement me")
}

func (a *Action) ReportStringValue(valueName string, value string) IAction {
	panic("implement me")
}

func (a *Action) ReportStringValueAt(valueName string, value string, timestamp time.Time) IAction {
	panic("implement me")
}

func (a *Action) ReportFloat64Value(valueName string, value float64) IAction {
	panic("implement me")
}

func (a *Action) ReportFloat64ValueAt(valueName string, value float64, timestamp time.Time) IAction {
	panic("implement me")
}

func (a *Action) ReportError(errorName string, errorCode int) IAction {
	panic("implement me")
}

func (a *Action) ReportErrorAt(errorName string, errorCode int, timestamp time.Time) IAction {
	panic("implement me")
}

func (a *Action) ReportException(errorName string, causeName string, causeDescription string, causeStack string) IAction {
	panic("implement me")
}

func (a *Action) ReportExceptionAt(errorName string, causeName string, causeDescription string, causeStack string, timestamp time.Time) IAction {
	panic("implement me")
}

func (a *Action) LeaveAction() IAction {
	panic("implement me")
}

func (a *Action) LeaveActionAt(timestamp time.Time) IAction {
	panic("implement me")
}

func (a *Action) CancelAction() IAction {
	panic("implement me")
}

func (a *Action) CancelActionAt(timestamp time.Time) IAction {
	panic("implement me")
}

func (a *Action) GetDuration() time.Duration {
	panic("implement me")
}
