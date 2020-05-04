package openkitgo

import (
	log "github.com/sirupsen/logrus"
	"time"
)

/*
type WebRequestTracer interface {
	GetTag() string
	SetResponseCode(int) WebRequestTracer
	setBytesSent(int) WebRequestTracer
	setBytesReceived(int) WebRequestTracer
	start() WebRequestTracer
	stop() WebRequestTracer
}
*/

type WebRequestTracer struct {
	log             *log.Logger
	parent          OpenKitComposite
	url             string
	beacon          *Beacon
	parentActionID  int
	ResponseCode    int
	BytesSent       int
	BytesReceived   int
	startTime       time.Time
	endTime         time.Time
	startSequenceNo int
	endSequenceNo   int
	tag             string
}

func NewWebRequestTracer(log *log.Logger, parent OpenKitComposite, url string, beacon *Beacon) *WebRequestTracer {
	w := new(WebRequestTracer)
	w.log = log
	w.parent = parent
	w.url = url
	w.beacon = beacon

	w.parentActionID = parent.getActionID()
	w.startSequenceNo = beacon.CreateSequenceNumber()

	w.tag = beacon.createTag(w.parentActionID, w.startSequenceNo)
	w.startTime = time.Now()

	return w
}

func NewWebRequestTracerAt(log *log.Logger, parent OpenKitComposite, url string, beacon *Beacon, timestamp time.Time) *WebRequestTracer {
	w := new(WebRequestTracer)
	w.log = log
	w.parent = parent
	w.url = url
	w.beacon = beacon

	w.parentActionID = parent.getActionID()
	w.startSequenceNo = beacon.CreateSequenceNumber()

	w.tag = beacon.createTag(w.parentActionID, w.startSequenceNo)
	w.startTime = timestamp

	return w
}

func (w *WebRequestTracer) Start() *WebRequestTracer {
	return w
}

func (w *WebRequestTracer) Stop(responseCode int) {
	w.ResponseCode = responseCode
	w.endSequenceNo = w.beacon.CreateSequenceNumber()
	w.endTime = time.Now()

	w.beacon.addWebRequest(w.parentActionID, w)
	// w.parent.onChildClosed(w)
	w.parent = nil

}

func (w *WebRequestTracer) StopAt(responseCode int, timestamp time.Time) {
	w.ResponseCode = responseCode
	w.endSequenceNo = w.beacon.CreateSequenceNumber()
	w.endTime = timestamp

	w.beacon.addWebRequest(w.parentActionID, w)
	// w.parent.onChildClosed(w)
	w.parent = nil

}

func (w *WebRequestTracer) close() {
	w.Stop(w.ResponseCode)
}

func (w *WebRequestTracer) closeAt(timestamp time.Time) {
	w.StopAt(w.ResponseCode, timestamp)
}
