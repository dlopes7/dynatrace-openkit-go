package communication

type BeaconState interface {
	terminal() bool
	execute(ctx *BeaconSendingContext)
	onInterrupted(ctx *BeaconSendingContext)
	getShutdownState() BeaconState
}
