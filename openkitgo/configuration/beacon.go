package configuration

import "sync"

type BeaconConfiguration struct {
	OpenKitConfiguration       *OpenKitConfiguration
	ServerConfiguration        *ServerConfiguration
	HttpClientConfiguration    *HttpClientConfiguration
	PrivacyConfiguration       *PrivacyConfiguration
	serverConfigurationSet     bool
	serverConfigUpdateCallback func(configuration *ServerConfiguration)

	mutex sync.Mutex
}

func NewBeaconConfiguration(
	openKitConfiguration *OpenKitConfiguration,
	privacyConfiguration *PrivacyConfiguration,
	serverID int,
) *BeaconConfiguration {

	// TODO create from config
	h := &HttpClientConfiguration{
		ServerID: serverID,
	}
	return &BeaconConfiguration{
		OpenKitConfiguration:    openKitConfiguration,
		HttpClientConfiguration: h,
		PrivacyConfiguration:    privacyConfiguration,
	}
}

func (c *BeaconConfiguration) IsServerConfigurationSet() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.serverConfigurationSet
}

func (c *BeaconConfiguration) SetServerConfigurationUpdateCallback(callback func(configuration *ServerConfiguration)) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.serverConfigUpdateCallback = callback
}

func (c *BeaconConfiguration) notifyServerConfigurationUpdate(configuration *ServerConfiguration) {
	if c.serverConfigUpdateCallback != nil {
		c.serverConfigUpdateCallback(configuration)
	}
}
