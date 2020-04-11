package openkitgo

import (
	log "github.com/sirupsen/logrus"
	"net/url"
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
	log          log.Logger
	responseCode int
	headers      map[string][]string

	capture        bool
	sendInterval   int
	monitorName    string
	serverID       int
	maxBeaconSize  int
	captureErrors  bool
	captureCrashes bool
	multiplicity   int
}

type KeyValuePair struct {
	key   string
	value string
}

func NewStatusResponse(log log.Logger, response string, responseCode int, headers map[string][]string) *StatusResponse {

	s := new(StatusResponse)
	s.log = log
	s.responseCode = responseCode
	s.headers = headers

	s.capture = true
	s.captureCrashes = true
	s.captureErrors = true
	s.sendInterval = -1
	s.serverID = -1
	s.maxBeaconSize = -1
	s.multiplicity = 1

	log.Debugf("NewStatusResponse: %s", response)

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
			s.maxBeaconSize = value * 1024

		} else if RESPONSE_KEY_CAPTURE_ERRORS == kv.key {
			value, _ := strconv.Atoi(kv.value)
			s.captureErrors = value != 0

		} else if RESPONSE_KEY_CAPTURE_CRASHES == kv.key {
			value, _ := strconv.Atoi(kv.value)
			s.captureCrashes = value != 0

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

func encodeWithReservedChars(input string, encoding string, additionalReservedChars []string) string {
	return url.PathEscape(input)
}

func isSuccessfulResponse(response *StatusResponse) bool {
	return response != nil && response.responseCode < 400

}

func isTooManyRequestsResponse(response *StatusResponse) bool {
	return response != nil && response.responseCode == 429

}

// vv=3&va=7.0.0000&ap=2d18d003-3c76-47b1-9649-463da552e41e&an=My%20OpenKit%20application&vn=1.0.0.0&pt=1&tt=okjava&vi=42&sn=1740741601&ip=8.8.8.8&os=Windows%2010&mf=MyCompany&md=MyModelID&dl=2&cl=2&tx=1555643572502&tv=1555643558485&mp=1&et=18&it=1&pa=0&s0=1&t0=0&et=60&na=jane.doe%40example.com&it=1&pa=0&s0=2&t0=3006&et=19&it=1&pa=0&s0=7&t0=14014&et=1&na=childAction&it=1&ca=2&pa=1&s0=4&t0=7010&s1=5&t1=3001&et=1&na=rootActionName&it=1&ca=1&pa=0&s0=3&t0=4010&s1=6&t1=7003
// vv=3&va=7.0.0000&ap=2d18d003-3c76-47b1-9649-463da552e41e&an=David%20Helper&vn=1.000&pt=1&tt=okjava&vi=19&sn=1427131847&ip=192.168.15.102&os=arch&mf=david%20inc&md=dellzao&dl=2&cl=2&tx=1555643229523&tv=1555643229057&mp=1&pa=0&s0=1&t0=0
