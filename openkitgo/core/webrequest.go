package core

import (
	"fmt"
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo/interfaces"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

const (
	UNKNOWN_URL = "<unknown>"
)

type WebRequestTracer struct {
	log             *log.Logger
	parent          OpenKitComposite
	mutex           sync.RWMutex
	tag             string
	beacon          *Beacon
	parentActionID  int
	url             string
	bytesReceived   int
	bytesSent       int
	startTime       time.Time
	endTime         time.Time
	startSequenceNo int32
	endSequenceNo   int32
	responseCode    int
}

func NewWebRequestTracer(log *log.Logger, parent OpenKitComposite, url string, beacon *Beacon, timestamp time.Time) *WebRequestTracer {

	t := &WebRequestTracer{
		log:             log,
		parent:          parent,
		beacon:          beacon,
		url:             url,
		startTime:       timestamp,
		bytesReceived:   -1,
		bytesSent:       -1,
		endSequenceNo:   -1,
		startSequenceNo: beacon.CreateSequenceNumber(),
		parentActionID:  parent.getActionID(),
	}
	t.tag = beacon.CreateTag(t.parentActionID, int(t.startSequenceNo))

	return t
}

func (w *WebRequestTracer) GetTag() string {
	w.log.WithFields(log.Fields{"tag": w.tag}).Debug("WebRequestTracer.GetTag()")
	return w.tag
}

func (w *WebRequestTracer) SetBytesSent(bytesSent int) interfaces.WebRequestTracer {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	if !w.isStopped() {
		w.bytesSent = bytesSent
	}
	return w
}

func (w *WebRequestTracer) SetBytesReceived(bytesReceived int) interfaces.WebRequestTracer {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	if !w.isStopped() {
		w.bytesReceived = bytesReceived
	}
	return w
}

func (w *WebRequestTracer) Start() interfaces.WebRequestTracer {
	return w.StartAt(time.Now())
}

func (w *WebRequestTracer) StartAt(timestamp time.Time) interfaces.WebRequestTracer {
	w.log.WithFields(log.Fields{"timestamp": timestamp}).Debug("WebRequestTracer.Start()")

	w.mutex.Lock()
	defer w.mutex.Unlock()
	if !w.isStopped() {
		w.startTime = timestamp
	}

	return w
}

func (w *WebRequestTracer) Stop(responseCode int) {
	w.StopAt(responseCode, time.Now())
}

func (w *WebRequestTracer) StopAt(responseCode int, timestamp time.Time) {
	w.log.WithFields(log.Fields{"timestamp": timestamp, "responseCode": responseCode}).Debug("WebRequestTracer.Stop()")
	w.doStop(responseCode, false, timestamp)
}

func (w *WebRequestTracer) doStop(responseCode int, discardData bool, timestamp time.Time) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	if w.isStopped() {
		return
	}
	w.responseCode = responseCode
	w.endSequenceNo = w.beacon.CreateSequenceNumber()
	w.endTime = timestamp

	if !discardData {
		w.beacon.addWebRequest(w.parentActionID, w)
	}

	w.parent.onChildClosed(w)
	w.parent = nil
}

func (w *WebRequestTracer) isStopped() bool {
	return !w.endTime.IsZero()
}

func (w *WebRequestTracer) close() {
	w.closeAt(time.Now())
}

func (w *WebRequestTracer) closeAt(timestamp time.Time) {
	w.StopAt(w.responseCode, timestamp)
}

func (w *WebRequestTracer) String() string {
	return fmt.Sprintf("WebRequestTracer(sn=%d, id=%d, url=%s)", w.beacon.GetSessionNumber(), w.parentActionID, w.url)
}
