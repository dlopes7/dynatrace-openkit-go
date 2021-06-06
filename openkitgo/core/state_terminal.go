package core

type StateTerminal struct{}

func (s *StateTerminal) terminal() bool {
	return true
}

func (s *StateTerminal) execute(ctx *BeaconSendingContext) {
	ctx.requestShutDown()
}

func (s *StateTerminal) getShutdownState() BeaconState {
	return s
}
func (s *StateTerminal) String() string {
	return "StateTerminal"
}
