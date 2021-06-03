package protocol

import (
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"time"
)

const (
	RESPONSE_KEY_RETRY_AFTER = "retry-after"
	DEFAULT_RETRY_AFTER      = 10 * time.Minute
	RESPONSE_STATUS_ERROR    = "ERROR"
)

type StatusResponse struct {
	log                *log.Logger
	ResponseCode       int
	Headers            http.Header
	ResponseAttributes ResponseAttributes
}

func NewStatusResponse(log *log.Logger, attributes ResponseAttributes, responseCode int, headers http.Header) StatusResponse {
	return StatusResponse{
		log:                log,
		ResponseAttributes: attributes,
		Headers:            headers,
		ResponseCode:       responseCode,
	}
}

func (s *StatusResponse) GetRetryAfter() time.Duration {

	h := s.Headers.Get(RESPONSE_KEY_RETRY_AFTER)
	if h == "" {
		s.log.WithFields(log.Fields{"using": DEFAULT_RETRY_AFTER}).Warning("the retry after header is not available")
		return DEFAULT_RETRY_AFTER
	}

	n, err := strconv.Atoi(h)
	if err != nil {
		s.log.WithFields(log.Fields{"using": DEFAULT_RETRY_AFTER, "header": h}).Error("Could not parse the header to a number")
		return DEFAULT_RETRY_AFTER
	}
	return time.Duration(n) * time.Second

}
