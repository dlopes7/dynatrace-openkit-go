package openkitgo

import (
	log "github.com/sirupsen/logrus"
	"math"
	"time"
)

type BeaconSendingState interface {
	execute(*BeaconSenderContext)
	isTerminalState() bool
	getShutdownState() BeaconSendingState
	String() string
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
		context.lastOpenSessionBeaconSendTime = time.Now()
		context.lastStatusCheckTime = time.Now()

		statusResponse = b.sendStatusRequest(context)

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

func (b *beaconSendingInitState) String() string {
	return "Init"
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

	// Send all new Sessions (Beacon Configure not set yet)
	newSessionsResponse := b.sendNewSessionRequests(context)
	if isTooManyRequestsResponse(newSessionsResponse) {
		context.nextState = &beaconSendingCaptureOffState{
			retryAfter: 10 * time.Minute, // TODO - This has to match the response header
		}
		return
	}

	finishedSessionsResponse := b.sendFinishedSessions(context)
	if isTooManyRequestsResponse(finishedSessionsResponse) {
		context.nextState = &beaconSendingCaptureOffState{
			retryAfter: 10 * time.Minute, // TODO - This has to match the response header
		}
		return
	}
	openSessionsResponse := b.sendOpenSessions(context)

	lastStatusResponse := newSessionsResponse
	if openSessionsResponse != nil {
		lastStatusResponse = openSessionsResponse
	}

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

		if statusResponse != nil {
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

	}

	return statusResponse

}

func (b *beaconSendingCaptureOnState) sendFinishedSessions(context *BeaconSenderContext) *StatusResponse {

	var statusResponse *StatusResponse

	for _, finishedSession := range context.getAllFinishedAndConfiguredSessions() {
		context.log.Debug("Found finished session! Sending!")

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

	now := time.Now()
	if now.Before(context.lastOpenSessionBeaconSendTime.Add(DEFAULT_SEND_INTERVAL)) {
		return nil
	}

	for _, openSession := range context.getAllOpenAndConfiguredSessions() {
		if openSession.isDataSendingAllowed() {
			context.log.WithFields(log.Fields{"session": openSession}).Debugf("Sending open session")
			statusResponse = openSession.sendBeacon(context.httpClient)
		}
	}

	context.lastOpenSessionBeaconSendTime = now

	return statusResponse

}

func (b *beaconSendingCaptureOnState) getShutdownState() BeaconSendingState {
	return &beaconSendingFlushSessionsState{}
}

func (beaconSendingCaptureOnState) isTerminalState() bool {
	return false
}

func (beaconSendingCaptureOnState) String() string {
	return "Capture On"
}

// ------------------------------------------------------------
// beaconSendingCaptureOffState -> beaconSendingCaptureOnState

type beaconSendingCaptureOffState struct {
	retryAfter time.Duration
}

func (beaconSendingCaptureOffState) execute(context *BeaconSenderContext) {
	// TODO - Implement Capture Off logic
	context.nextState = &beaconSendingCaptureOnState{}

}

func (beaconSendingCaptureOffState) String() string {
	return "Capture Off"
}

func (beaconSendingCaptureOffState) isTerminalState() bool {
	return false
}

func (b *beaconSendingCaptureOffState) getShutdownState() BeaconSendingState {
	return &beaconSendingFlushSessionsState{}
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

func (beaconSendingFlushSessionsState) String() string {
	return "Flush"
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

	// TODO - Implement beaconSendingTerminalState
	context.nextState = &beaconSendingCaptureOnState{}

	context.log.Debug("Executed state beaconSendingTerminalState")

}

func (beaconSendingTerminalState) String() string {
	return "Terminal"
}

func (beaconSendingTerminalState) isTerminalState() bool {
	return false
}

func (b *beaconSendingTerminalState) getShutdownState() BeaconSendingState {
	return &beaconSendingCaptureOnState{}
}
