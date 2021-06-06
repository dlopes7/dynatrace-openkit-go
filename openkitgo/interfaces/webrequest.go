package interfaces

import "time"

type WebRequestTracer interface {
	GetTag() string

	SetBytesSent(bytesSent int) WebRequestTracer
	SetBytesReceived(bytesReceived int) WebRequestTracer

	Start() WebRequestTracer
	StartAt(timestamp time.Time) WebRequestTracer

	Stop(responseCode int)
	StopAt(responseCode int, timestamp time.Time)
}
