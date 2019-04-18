package openkitgo

import (
	"encoding/hex"
	"github.com/op/go-logging"
	"strconv"
	"strings"
)

const (
	RESPONSE_KEY_CAPTURE         = "cp"
	RESPONSE_KEY_SEND_INTERVAL   = "si"
	RESPONSE_KEY_MONITOR_NAME    = "bn"
	RESPONSE_KEY_SERVER_ID       = "id"
	RESPONSE_KEY_MAX_BEACON_SIZE = "bl"
	RESPONSE_KEY_CAPTURE_ERRORS  = "er"
	RESPONSE_KEY_CAPTURE_CRASHES = "cr"
	RESPONSE_KEY_MULTIPLICITY    = "mp"
)

type StatusResponse struct {
	logger       logging.Logger
	responseCode int
	headers      map[string][]string

	capture        bool
	sendInterval   int
	monitorName    string
	serverID       int
	maxBeaconSize  int
	captureErrors  int
	captureCrashes int
	multiplicity   int
}

type KeyValuePair struct {
	key   string
	value string
}

func NewStatusResponse(logger logging.Logger, response string, responseCode int, headers map[string][]string) *StatusResponse {

	s := new(StatusResponse)
	s.logger = logger
	s.responseCode = responseCode
	s.headers = headers

	s.parseResponse(response)
	return s
}

func (s *StatusResponse) parseResponse(response string) {

	for _, kv := range s.parseResponseKeyValuePair(response) {

		if RESPONSE_KEY_CAPTURE == kv.key {
			s.capture = kv.value == "1"
		} else if RESPONSE_KEY_SEND_INTERVAL == kv.key {
			value, _ := strconv.Atoi(kv.value)
			s.sendInterval = value * 1000

		} else if RESPONSE_KEY_MONITOR_NAME == kv.key {
			s.monitorName = kv.value

		} else if RESPONSE_KEY_SERVER_ID == kv.key {
			value, _ := strconv.Atoi(kv.value)
			s.serverID = value

		} else if RESPONSE_KEY_MAX_BEACON_SIZE == kv.key {
			value, _ := strconv.Atoi(kv.value)
			s.maxBeaconSize = value

		} else if RESPONSE_KEY_CAPTURE_ERRORS == kv.key {
			value, _ := strconv.Atoi(kv.value)
			s.captureErrors = value

		} else if RESPONSE_KEY_CAPTURE_CRASHES == kv.key {
			value, _ := strconv.Atoi(kv.value)
			s.captureCrashes = value

		} else if RESPONSE_KEY_MULTIPLICITY == kv.key {
			value, _ := strconv.Atoi(kv.value)
			s.multiplicity = value

		}

	}

}

func (s *StatusResponse) parseResponseKeyValuePair(response string) []*KeyValuePair {

	result := make([]*KeyValuePair, 0)
	tokens := strings.Split(response, "&")

	for _, token := range tokens {
		keyValueSeparatorIndex := strings.Index(token, "=")
		if keyValueSeparatorIndex != -1 {
			keyValue := strings.Split(token, "=")
			result = append(result, &KeyValuePair{
				key:   keyValue[0],
				value: keyValue[1],
			})
		}
	}

	return result
}

func encodeWithReservedChars(input string, encoding string, additionalReservedChars []rune) string {
	src := []byte(input)

	dst := make([]byte, hex.EncodedLen(len(src)))
	hex.Encode(dst, src)

	return string(dst)

}
