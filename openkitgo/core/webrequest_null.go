package core

import (
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo"
	"time"
)

type NullWebRequestTracer struct{}

func NewNullWebRequestTracer() openkitgo.WebRequestTracer {
	return &NullWebRequestTracer{}
}

func (n *NullWebRequestTracer) GetTag() string {
	return ""
}

func (n *NullWebRequestTracer) SetBytesSent(bytesSent int) openkitgo.WebRequestTracer {
	return n
}

func (n *NullWebRequestTracer) SetBytesReceived(bytesReceived int) openkitgo.WebRequestTracer {
	return n
}

func (n *NullWebRequestTracer) Start() openkitgo.WebRequestTracer {
	return n
}

func (n *NullWebRequestTracer) StartAt(timestamp time.Time) openkitgo.WebRequestTracer {
	return n
}

func (n *NullWebRequestTracer) Stop(responseCode int) {}

func (n *NullWebRequestTracer) StopAt(responseCode int, timestamp time.Time) {}
