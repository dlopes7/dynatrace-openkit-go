package core

import (
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo/interfaces"
	"time"
)

type NullWebRequestTracer struct{}

func NewNullWebRequestTracer() interfaces.WebRequestTracer {
	return &NullWebRequestTracer{}
}

func (n *NullWebRequestTracer) GetTag() string {
	return ""
}

func (n *NullWebRequestTracer) SetBytesSent(bytesSent int) interfaces.WebRequestTracer {
	return n
}

func (n *NullWebRequestTracer) SetBytesReceived(bytesReceived int) interfaces.WebRequestTracer {
	return n
}

func (n *NullWebRequestTracer) Start() interfaces.WebRequestTracer {
	return n
}

func (n *NullWebRequestTracer) StartAt(timestamp time.Time) interfaces.WebRequestTracer {
	return n
}

func (n *NullWebRequestTracer) Stop(responseCode int) {}

func (n *NullWebRequestTracer) StopAt(responseCode int, timestamp time.Time) {}
