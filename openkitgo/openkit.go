package openkitgo

import (
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo/core"
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo/interfaces"
)

func NewOpenKitBuilder(endpointURL string, applicationID string, deviceID int64) interfaces.OpenKitBuilder {
	return core.NewOpenKitBuilder(endpointURL, applicationID, deviceID)
}
