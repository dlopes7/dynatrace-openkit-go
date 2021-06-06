package core

import (
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo/configuration"
	log "github.com/sirupsen/logrus"
	"time"
)

const (
	THREAD_NAME      = "BeaconSender"
	SHUTDOWN_TIMEOUT = 10 * time.Second
)

type BeaconSender struct {
	log     *log.Logger
	context *BeaconSendingContext

	// Used to communicate with the sender goroutine
	channel *chan bool
}

func NewBeaconSender(log *log.Logger, httpClientConfig *configuration.HttpClientConfiguration) *BeaconSender {

	return &BeaconSender{
		log:     log,
		context: NewBeaconSendingContext(log, httpClientConfig),
	}
}

// BeaconSenderRoutine contains the goroutine that runs until a shutdown is requested
func BeaconSenderRoutine(log *log.Logger, ctx *BeaconSendingContext) *chan bool {

	stop := make(chan bool)

	go func() {
		log.Debug("BeaconSenderRoutine.start()")
		for !ctx.IsInTerminalState() {
			ctx.executeCurrentState()
		}
		log.Debug("BeaconSenderRoutine.stop()")
	}()

	return &stop
}

func (s *BeaconSender) Initialize() {
	s.channel = BeaconSenderRoutine(s.log, s.context)
}

func (s *BeaconSender) WaitForInit() bool {
	return s.context.WaitForInit()
}

func (s *BeaconSender) WaitForInitTimeout(duration time.Duration) bool {
	return s.context.WaitForInitTimeout(duration)
}

func (s *BeaconSender) IsInitialized() bool {
	return s.context.IsInitialized()
}

func (s *BeaconSender) Shutdown() {
	s.context.requestShutDown()
	*s.channel <- true
}

func (s *BeaconSender) GetLastServerConfiguration() *configuration.ServerConfiguration {
	return s.context.GetLastServerConfiguration()
}

func (s *BeaconSender) GetCurrentServerId() int {
	return s.context.GetCurrentServerId()
}

func (s *BeaconSender) AddSession(session *Session) {
	s.log.WithFields(log.Fields{"session": session.String()}).Debug("BeaconSender.AddSession()")
	s.context.AddSession(session)
}
