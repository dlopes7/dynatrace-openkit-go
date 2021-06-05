package core

type StateCaptureOff struct {
}

func NewStateCaptureOff() *StateCaptureOff {
	return &StateCaptureOff{}
}

func (s StateCaptureOff) terminal() bool {
	panic("implement me")
}

func (s StateCaptureOff) execute(ctx *BeaconSendingContext) {
	panic("implement me")
}

func (s StateCaptureOff) getShutdownState() BeaconState {
	panic("implement me")
}
