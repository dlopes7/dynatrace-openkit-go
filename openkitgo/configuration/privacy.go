package configuration

type DataCollectionLevel int

const (
	DATA_OFF           DataCollectionLevel = 0
	DATA_PERFORMANCE   DataCollectionLevel = 1
	DATA_USER_BEHAVIOR DataCollectionLevel = 2
)

type CrashReportingLevel int

const (
	CRASH_OFF             CrashReportingLevel = 0
	CRASH_OPT_OUT_CRASHES CrashReportingLevel = 1
	CRASH_OPT_IN_CRASHES  CrashReportingLevel = 2
)

type PrivacyConfiguration struct {
	DataCollectionLevel DataCollectionLevel
	CrashReportingLevel CrashReportingLevel
}

// TODO func NewPrivacyConfiguration(builder)

func (c *PrivacyConfiguration) IsDeviceIDSendingAllowed() bool {
	return c.DataCollectionLevel == DATA_USER_BEHAVIOR
}

func (c *PrivacyConfiguration) IsSessionNumberReportingAllowed() bool {
	return c.DataCollectionLevel == DATA_USER_BEHAVIOR
}

func (c *PrivacyConfiguration) IsWebRequestTracingAllowed() bool {
	return c.DataCollectionLevel != DATA_OFF
}

func (c *PrivacyConfiguration) IsSessionReportingAllowed() bool {
	return c.DataCollectionLevel != DATA_OFF
}
func (c *PrivacyConfiguration) IsActionReportingAllowed() bool {
	return c.DataCollectionLevel != DATA_OFF
}

func (c *PrivacyConfiguration) IsValueReportingAllowed() bool {
	return c.DataCollectionLevel == DATA_USER_BEHAVIOR
}

func (c *PrivacyConfiguration) IsEventReportingAllowed() bool {
	return c.DataCollectionLevel == DATA_USER_BEHAVIOR
}

func (c *PrivacyConfiguration) IsErrorReportingAllowed() bool {
	return c.DataCollectionLevel != DATA_OFF
}
func (c *PrivacyConfiguration) IsCrashReportingAllowed() bool {
	return c.CrashReportingLevel == CRASH_OPT_IN_CRASHES
}

func (c *PrivacyConfiguration) isUserIdentificationAllowed() bool {
	return c.DataCollectionLevel == DATA_USER_BEHAVIOR
}
