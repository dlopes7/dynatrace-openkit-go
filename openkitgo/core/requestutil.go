package core

import (
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo/protocol"
	"net/http"
	"time"
)

type BeaconSendingRequestUtil struct{}

func sendStatusRequest(ctx *BeaconSendingContext, numRetries int, initialRetryDelay time.Duration) protocol.StatusResponse {

	var statusResponse protocol.StatusResponse

	sleepTime := initialRetryDelay
	retry := 0

	httpClient := ctx.getHttpClient()
	for {
		statusResponse = httpClient.SendStatusRequest(ctx)
		if statusResponse.ResponseCode < http.StatusBadRequest ||
			statusResponse.ResponseCode == 429 ||
			retry >= numRetries ||
			ctx.IsShutdownRequested() {

			// If we get here, stop trying
			// Everything either worked, or someone else asked us to stop
			break
		}
		time.Sleep(sleepTime)
		sleepTime *= 2
		retry++

	}
	return statusResponse
}
