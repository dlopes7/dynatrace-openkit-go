package core

type BeaconState interface {
	terminal() bool
	execute(ctx *BeaconSendingContext)
	getShutdownState() BeaconState
}
