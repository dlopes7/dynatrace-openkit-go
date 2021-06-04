package configuration

import (
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo"
	"net/http"
)

const (
	ENCODING_CHARSET    = "UTF-8"
	RESERVED_CHARACTERS = '_'
)

type OpenKitConfiguration struct {
	EndpointURL                 string
	DeviceID                    int64
	OrigDeviceID                string
	OpenKitType                 string
	ApplicationID               string
	PercentEncodedApplicationID string
	ApplicationName             string
	ApplicationVersion          string
	OperatingSystem             string
	Manufacturer                string
	ModelID                     string
	DefaultServerID             string
	Transport                   http.Transport
}

func NewOpenKitConfiguration(builder openkitgo.OpenKitBuilder) *OpenKitConfiguration {
	// TODO - Create from builder
	return &OpenKitConfiguration{}

}
