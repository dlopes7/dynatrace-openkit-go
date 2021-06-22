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
	sessionSplitBySessionDuration bool
	MaxEventsPerSession           int
	sessionSplitByEvents          bool
	SessionTimeout                time.Duration
	sessionSplitByIdleTimeout     bool
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
		sessionSplitBySessionDuration: attributes.MaxSessionDuration != 0,
		MaxEventsPerSession:           attributes.MaxEventsPerSession,
		sessionSplitByEvents:          attributes.MaxEventsPerSession != 0,
		SessionTimeout:                attributes.SessionTimeout,
		sessionSplitByIdleTimeout:     attributes.SessionTimeout != 0,
		VisitStoreVersion:             attributes.VisitStoreVersion,
		TrafficControlPercentage:      attributes.TrafficControlPercentage,
	}
}

func DefaultServerConfiguration() *ServerConfiguration {
	return NewServerConfiguration(protocol.DefaultResponseAttributes())
}

func (c *ServerConfiguration) IsSendingDataAllowed() bool {
	return c.Capture && c.Multiplicity > 0
}

func (c *ServerConfiguration) IsSendingErrorsAllowed() bool {
	return c.IsSendingDataAllowed() && c.ErrorReporting

}

func (c *ServerConfiguration) IsSessionSplitByEventsEnabled() bool {
	return c.sessionSplitByEvents && c.MaxEventsPerSession > 0
}

func (c *ServerConfiguration) IsSessionSplitBySessionDurationEnabled() bool {
	return c.sessionSplitBySessionDuration && c.MaxSessionDuration > 0

}

func (c *ServerConfiguration) IsSessionSplitByIdleTimeoutEnabled() bool {
	return c.sessionSplitByIdleTimeout && c.SessionTimeout > 0

}
