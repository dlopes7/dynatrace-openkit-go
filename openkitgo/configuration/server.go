package configuration

import (
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo/protocol"
	"time"
)

type ServerConfiguration struct {
	Capture                       bool
	CrashReporting                bool
	ErrorReporting                bool
	ServerID                      int
	BeaconSizeInBytes             int
	Multiplicity                  int
	SendInterval                  time.Duration
	MaxSessionDuration            time.Duration
	SessionSplitBySessionDuration bool
	MaxEventsPerSession           int
	SessionSplitByEvents          bool
	SessionTimeout                time.Duration
	SessionSplitByIdleTimeout     bool
	VisitStoreVersion             int
	TrafficControlPercentage      int
}

func NewServerConfiguration(attributes protocol.ResponseAttributes) *ServerConfiguration {
	return &ServerConfiguration{
		Capture:                       attributes.Capture,
		CrashReporting:                attributes.CaptureCrashes,
		ErrorReporting:                attributes.CaptureErrors,
		ServerID:                      attributes.ServerID,
		BeaconSizeInBytes:             attributes.MaxBeaconSizeInBytes,
		Multiplicity:                  attributes.Multiplicity,
		SendInterval:                  attributes.SendInterval,
		MaxSessionDuration:            attributes.MaxSessionDuration,
		SessionSplitBySessionDuration: attributes.MaxSessionDuration != 0,
		MaxEventsPerSession:           attributes.MaxEventsPerSession,
		SessionSplitByEvents:          attributes.MaxEventsPerSession != 0,
		SessionTimeout:                attributes.SessionTimeout,
		SessionSplitByIdleTimeout:     attributes.SessionTimeout != 0,
		VisitStoreVersion:             attributes.VisitStoreVersion,
		TrafficControlPercentage:      attributes.TrafficControlPercentage,
	}
}

func DefaultServerConfiguration() *ServerConfiguration {
	return NewServerConfiguration(protocol.UndefinedResponseAttributes())
}

func (c *ServerConfiguration) IsSendingDataAllowed() bool {
	return c.Capture && c.Multiplicity > 0
}
