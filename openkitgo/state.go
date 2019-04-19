package openkitgo

import (
	"fmt"
	"math"
	"time"
)

type BeaconSendingState interface {
	execute(*BeaconSenderContext)
	isTerminalState() bool
	getShutdownState() BeaconSendingState
	ToString() string
}

var REINIT_DELAY_MILLISECONDS = []time.Duration{
	1 * time.Minute,
	5 * time.Minute,
	15 * time.Minute,
	1 * time.Hour,
	2 * time.Hour,
}

const (
	MAX_INITIAL_STATUS_REQUEST_RETRIES    = 5
	INITIAL_RETRY_SLEEP_TIME_MILLISECONDS = 1 * time.Second
)

// ------------------------------------------------------------
// beaconSendingInitState -> beaconSendingCaptureOnState

type beaconSendingInitState struct {
	reinitializeDelayIndex int
}

func (b *beaconSendingInitState) execute(context *BeaconSenderContext) {

	statusResponse := b.executeStatusRequest(context)

	if context.shutdown {
		context.initCompleted = false
	} else if statusResponse.responseCode < 400 {
		context.handleStatusResponse(statusResponse)
		context.nextState = &beaconSendingCaptureOnState{}
		context.initCompleted = true
	}
}

func (b *beaconSendingInitState) executeStatusRequest(context *BeaconSenderContext) *StatusResponse {
	var statusResponse *StatusResponse

	for {
		currentTimestamp := context.config.makeTimestamp()
		context.lastOpenSessionBeaconSendTime = currentTimestamp
		context.lastStatusCheckTime = currentTimestamp

		statusResponse = b.sendStatusRequest(context)

		fmt.Println("Context", context, "statusResponse", statusResponse)

		if context.shutdown || (statusResponse != nil && statusResponse.responseCode < 400) {
			break
		}

		sleepTime := REINIT_DELAY_MILLISECONDS[b.reinitializeDelayIndex]
		context.sleep(sleepTime)
		b.reinitializeDelayIndex = int(math.Min(float64(b.reinitializeDelayIndex+1), float64(len(REINIT_DELAY_MILLISECONDS)-1)))

	}

	return statusResponse

}
func (b *beaconSendingInitState) sendStatusRequest(context *BeaconSenderContext) *StatusResponse {
	var statusResponse *StatusResponse

	sleepTime := INITIAL_RETRY_SLEEP_TIME_MILLISECONDS
	retry := 0

	for {
		statusResponse = context.httpClient.sendStatusRequest()
		if statusResponse != nil {
			if statusResponse.responseCode < 400 ||
				retry >= MAX_INITIAL_STATUS_REQUEST_RETRIES ||
				context.shutdown {
				break
			}
		}

		context.sleep(sleepTime)
		sleepTime *= 2
		retry += 1
	}

	return statusResponse
}

func (b *beaconSendingInitState) ToString() string {
	return "beaconSendingInitState"
}

func (b *beaconSendingInitState) isTerminalState() bool {
	return false
}

func (b *beaconSendingInitState) getShutdownState() BeaconSendingState {
	return &beaconSendingCaptureOnState{}
}

// ------------------------------------------------------------
// beaconSendingCaptureOnState -> beaconSendingFlushSessionsState

type beaconSendingCaptureOnState struct{}

func (b *beaconSendingCaptureOnState) execute(context *BeaconSenderContext) {
	context.sleep(1 * time.Second)
	context.logger.Info("Executed state beaconSendingCaptureOnState")

	newSessionsResponse := b.sendNewSessionRequests(context)
	finishedSessionsResponse := b.sendFinishedSessions(context)
	// openSessionsResponse := b.sendOpenSessions(context)

	lastStatusResponse := newSessionsResponse

	// if openSessionsResponse != nil {
	//	lastStatusResponse = openSessionsResponse
	//}

	if finishedSessionsResponse != nil {
		lastStatusResponse = finishedSessionsResponse
	}

	b.handleStatusResponse(context, lastStatusResponse)

}

func (b *beaconSendingCaptureOnState) handleStatusResponse(context *BeaconSenderContext, statusResponse *StatusResponse) {
	if statusResponse == nil {
		return
	}
	context.handleStatusResponse(statusResponse)
}

func (b *beaconSendingCaptureOnState) sendNewSessionRequests(context *BeaconSenderContext) *StatusResponse {

	var statusResponse *StatusResponse

	for _, session := range context.getAllNewSessions() {

		if !session.canSendNewSessionRequest() {
			currentConfiguration := session.getBeaconConfiguration()
			newConfiguration := &BeaconConfiguration{
				multiplicity:        0,
				dataCollectionLevel: currentConfiguration.dataCollectionLevel,
				crashReportingLevel: currentConfiguration.crashReportingLevel,
			}
			session.updateBeaconConfiguration(newConfiguration)
			continue
		}

		statusResponse = context.httpClient.sendNewSessionRequest()

		if statusResponse.responseCode < 400 {
			currentConfiguration := session.getBeaconConfiguration()
			newConfiguration := &BeaconConfiguration{
				multiplicity:        statusResponse.multiplicity,
				dataCollectionLevel: currentConfiguration.dataCollectionLevel,
				crashReportingLevel: currentConfiguration.crashReportingLevel,
			}
			session.updateBeaconConfiguration(newConfiguration)

		}

	}

	return statusResponse

}

func (b *beaconSendingCaptureOnState) sendFinishedSessions(context *BeaconSenderContext) *StatusResponse {

	var statusResponse *StatusResponse

	for _, finishedSession := range context.getAllFinishedAndConfiguredSessions() {
		context.logger.Debug("Found finished session! Sending!")

		if finishedSession.isDataSendingAllowed() {
			context.removeSession(finishedSession)
			statusResponse = finishedSession.sendBeacon(context.httpClient)
			finishedSession.clearCapturedData()
			finishedSession.End()

		}
	}

	return statusResponse

}

func (b *beaconSendingCaptureOnState) sendOpenSessions(context *BeaconSenderContext) *StatusResponse {

	var statusResponse *StatusResponse

	// TODO - Implement setLastOpenSessionBeaconSendTime

	for _, openSession := range context.getAllOpenAndConfiguredSessions() {
		context.logger.Debug("Found opened session! Sending? ", openSession.isDataSendingAllowed())

		if openSession.isDataSendingAllowed() {
			statusResponse = openSession.sendBeacon(context.httpClient)
		}
	}

	return statusResponse

}

func (b *beaconSendingCaptureOnState) getShutdownState() BeaconSendingState {
	return &beaconSendingFlushSessionsState{}
}

func (beaconSendingCaptureOnState) isTerminalState() bool {
	return false
}

func (beaconSendingCaptureOnState) ToString() string {
	return "beaconSendingCaptureOnState"
}

// ------------------------------------------------------------
// beaconSendingFlushSessionsState -> beaconSendingTerminalState

type beaconSendingFlushSessionsState struct{}

func (beaconSendingFlushSessionsState) execute(context *BeaconSenderContext) {

	// first get all sessions that do not have any multiplicity se
	for _, newSession := range context.getAllNewSessions() {
		currentConfiguration := newSession.getBeaconConfiguration()
		newSession.updateBeaconConfiguration(&BeaconConfiguration{
			multiplicity:        1,
			dataCollectionLevel: currentConfiguration.dataCollectionLevel,
			crashReportingLevel: currentConfiguration.crashReportingLevel,
		})

	}

	context.nextState = &beaconSendingTerminalState{}

}

func (beaconSendingFlushSessionsState) ToString() string {
	return "beaconSendingInitState"
}

func (beaconSendingFlushSessionsState) isTerminalState() bool {
	return false
}

func (b *beaconSendingFlushSessionsState) getShutdownState() BeaconSendingState {
	return &beaconSendingCaptureOnState{}
}

// ------------------------------------------------------------
// beaconSendingTerminalState ->

type beaconSendingTerminalState struct{}

func (beaconSendingTerminalState) execute(context *BeaconSenderContext) {
	context.sleep(1 * time.Second)

	// TODO - Only set this if StatusRequest is OK
	context.nextState = &beaconSendingCaptureOnState{}

	context.logger.Info("Executed state beaconSendingInitState")

}

func (beaconSendingTerminalState) ToString() string {
	return "beaconSendingInitState"
}

func (beaconSendingTerminalState) isTerminalState() bool {
	return false
}

func (b *beaconSendingTerminalState) getShutdownState() BeaconSendingState {
	return &beaconSendingCaptureOnState{}
}
