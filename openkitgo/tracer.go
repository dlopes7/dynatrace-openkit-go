package openkitgo

type WebRequestTracer interface {
	GetTag() string
	SetResponseCode(int) WebRequestTracer
	setBytesSent(int) WebRequestTracer
	setBytesReceived(int) WebRequestTracer
	start() WebRequestTracer
	stop() WebRequestTracer
}
