package core

import (
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

type Action struct {
	log             *log.Logger
	parent          OpenKitComposite
	parentActionID  int
	mutex           sync.Mutex
	id              uint32
	name            string
	startTime       time.Time
	endTime         time.Time
	startSequenceNo int
	endSequenceNo   int
	actionLeft      bool
	beacon          *Beacon
}

func NewAction(log *log.Logger, parent OpenKitComposite, name string, beacon *Beacon, startTime time.Time) *Action {

	return &Action{
		log:             log,
		parent:          parent,
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

func (a *Action) ReportEvent(eventName string) openkitgo.IAction {
	panic("implement me")
}

func (a *Action) ReportEventAt(eventName string, timestamp time.Time) openkitgo.IAction {
	panic("implement me")
}

func (a *Action) ReportInt64Value(valueName string, value int64) openkitgo.IAction {
	panic("implement me")
}

func (a *Action) ReportInt64ValueAt(valueName string, value int64, timestamp time.Time) openkitgo.IAction {
	panic("implement me")
}

func (a *Action) ReportStringValue(valueName string, value string) openkitgo.IAction {
	panic("implement me")
}

func (a *Action) ReportStringValueAt(valueName string, value string, timestamp time.Time) openkitgo.IAction {
	panic("implement me")
}

func (a *Action) ReportFloat64Value(valueName string, value float64) openkitgo.IAction {
	panic("implement me")
}

func (a *Action) ReportFloat64ValueAt(valueName string, value float64, timestamp time.Time) openkitgo.IAction {
	panic("implement me")
}

func (a *Action) ReportError(errorName string, errorCode int) openkitgo.IAction {
	panic("implement me")
}

func (a *Action) ReportErrorAt(errorName string, errorCode int, timestamp time.Time) openkitgo.IAction {
	panic("implement me")
}

func (a *Action) ReportException(errorName string, causeName string, causeDescription string, causeStack string) openkitgo.IAction {
	panic("implement me")
}

func (a *Action) ReportExceptionAt(errorName string, causeName string, causeDescription string, causeStack string, timestamp time.Time) openkitgo.IAction {
	panic("implement me")
}

func (a *Action) LeaveAction() openkitgo.IAction {
	panic("implement me")
}

func (a *Action) LeaveActionAt(timestamp time.Time) openkitgo.IAction {
	panic("implement me")
}

func (a *Action) CancelAction() openkitgo.IAction {
	panic("implement me")
}

func (a *Action) CancelActionAt(timestamp time.Time) openkitgo.IAction {
	panic("implement me")
}

func (a *Action) GetDuration() time.Duration {
	panic("implement me")
}
