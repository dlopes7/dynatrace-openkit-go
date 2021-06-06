package configuration

import "sync"

type BeaconConfiguration struct {
	OpenKitConfiguration       *OpenKitConfiguration
	ServerConfiguration        *ServerConfiguration
	HttpClientConfiguration    *HttpClientConfiguration
	PrivacyConfiguration       *PrivacyConfiguration
	serverConfigurationSet     bool
	serverConfigUpdateCallback func(configuration *ServerConfiguration)

	mutex sync.RWMutex
}

func NewBeaconConfiguration(
	openKitConfiguration *OpenKitConfiguration,
	privacyConfiguration *PrivacyConfiguration,
	serverID int,
) *BeaconConfiguration {

	h := &HttpClientConfiguration{
		ServerID:  serverID,
		Transport: openKitConfiguration.Transport,
	}
	return &BeaconConfiguration{
		OpenKitConfiguration:    openKitConfiguration,
		HttpClientConfiguration: h,
		PrivacyConfiguration:    privacyConfiguration,
	}
}

func (c *BeaconConfiguration) IsServerConfigurationSet() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
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

func (c *BeaconConfiguration) GetServerConfiguration() *ServerConfiguration {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if c.ServerConfiguration == nil {
		return DefaultServerConfiguration()

	} else {
		return c.ServerConfiguration
	}
}

func (c *BeaconConfiguration) UpdateServerConfiguration(config *ServerConfiguration) {

	if config == nil {
		return
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.ServerConfiguration = config
	c.serverConfigurationSet = true

	c.notifyServerConfigurationUpdate(config)

}

func (c *BeaconConfiguration) InitializeServerConfiguration(config *ServerConfiguration) {

	if config == nil {
		return
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.ServerConfiguration = config
	c.serverConfigurationSet = true

	c.notifyServerConfigurationUpdate(config)

}

func (c *BeaconConfiguration) EnableCapture() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.ServerConfiguration.Capture = true

}

func (c *BeaconConfiguration) DisableCapture() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.ServerConfiguration.Capture = false

}
