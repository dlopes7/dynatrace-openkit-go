package protocol

import "time"

//
//type ResponseAttributes interface {
//	GetMaxBeaconSizeInBytes() int
//	GetMaxSessionDuration() time.Duration
//	GetMaxEventsPerSession() int
//	GetSessionTimeout() time.Duration
//	GetSendInterval() time.Duration
//	GetVisitStoreVersion() int
//	Capture() bool
//	CaptureCrashes() bool
//	CaptureErrors() bool
//	GetApplicationId() string
//	GetMultiplicity() int
//	GetServerId() int
//	GetStatus() string
//	GetTimestamp() time.Time
//
//	// TODO boolean isAttributeSet(ResponseAttribute attribute);
//	// TODO ResponseAttributes merge(ResponseAttributes responseAttributes);
//}

type ResponseAttributes struct {
	MaxBeaconSizeInBytes     int
	MaxSessionDuration       time.Duration
	MaxEventsPerSession      int
	SessionTimeout           time.Duration
	SendInterval             time.Duration
	VisitStoreVersion        int
	Capture                  bool
	CaptureCrashes           bool
	CaptureErrors            bool
	TrafficControlPercentage int
	ApplicationID            string
	Multiplicity             int
	ServerID                 int
	Status                   string
	Timestamp                time.Time
}

func DefaultResponseAttributes() ResponseAttributes {
	return ResponseAttributes{
		MaxBeaconSizeInBytes:     30 * 1024,
		MaxSessionDuration:       -1 * time.Millisecond,
		MaxEventsPerSession:      -1,
		SessionTimeout:           -1 * time.Millisecond,
		SendInterval:             2 * time.Minute,
		VisitStoreVersion:        1,
		Capture:                  true,
		CaptureCrashes:           true,
		CaptureErrors:            true,
		TrafficControlPercentage: 100,
		ApplicationID:            "",
		Multiplicity:             1,
		ServerID:                 1,
		Timestamp:                time.Time{},
	}
}

func UndefinedResponseAttributes() ResponseAttributes {

	return ResponseAttributes{
		MaxBeaconSizeInBytes: 30 * 1024,
		MaxSessionDuration:   -1 * time.Millisecond,
		MaxEventsPerSession:  -1,
		SessionTimeout:       -1 * time.Millisecond,
		ServerID:             -1,
	}

}

func (a *ResponseAttributes) Merge(attributes ResponseAttributes) ResponseAttributes {

	// TODO - Apparently only set these if they were not set before?
	a.ServerID = attributes.ServerID
	a.MaxBeaconSizeInBytes = attributes.MaxBeaconSizeInBytes
	a.Capture = attributes.Capture
	a.ApplicationID = attributes.ApplicationID
	a.Timestamp = attributes.Timestamp
	a.Multiplicity = attributes.Multiplicity
	a.Status = attributes.Status
	a.TrafficControlPercentage = attributes.TrafficControlPercentage
	a.CaptureErrors = attributes.CaptureErrors
	a.CaptureCrashes = attributes.CaptureCrashes
	a.VisitStoreVersion = attributes.VisitStoreVersion
	a.MaxSessionDuration = attributes.MaxSessionDuration
	a.MaxEventsPerSession = attributes.MaxEventsPerSession
	a.SendInterval = attributes.SendInterval
	a.Timestamp = attributes.Timestamp

	return *a
}
