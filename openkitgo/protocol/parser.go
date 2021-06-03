package protocol

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"strconv"
	"strings"
	"time"
)

const (
	KEY_VALUE_RESPONSE_TYPE_MOBILE                = "type=m"
	KEY_VALUE_RESPONSE_TYPE_MOBILE_WITH_SEPARATOR = "type=m&"

	RESPONSE_KEY_MAX_BEACON_SIZE_IN_KB      = "bl"
	RESPONSE_KEY_SEND_INTERVAL_IN_SEC       = "si"
	RESPONSE_KEY_CAPTURE                    = "cp"
	RESPONSE_KEY_REPORT_CRASHES             = "cr"
	RESPONSE_KEY_REPORT_ERRORS              = "er"
	RESPONSE_KEY_TRAFFIC_CONTROL_PERCENTAGE = "tc"
	RESPONSE_KEY_SERVER_ID                  = "id"
	RESPONSE_KEY_MULTIPLICITY               = "mp"
)

type JsonResponse struct {
	MobileAgentConfig MobileAgentConfig `json:"mobileAgentConfig"`
	AppConfig         AppConfig         `json:"appConfig"`
	DynamicConfig     DynamicConfig     `json:"dynamicConfig"`
	Timestamp         int64             `json:"timestamp"`
}
type MobileAgentConfig struct {
	MaxBeaconSizeKb        int  `json:"maxBeaconSizeKb"`
	SelfMonitoring         bool `json:"selfmonitoring"`
	MaxSessionDurationMins int  `json:"maxSessionDurationMins"`
	MaxEventsPerSession    int  `json:"maxEventsPerSession"`
	SessionTimeoutSec      int  `json:"sessionTimeoutSec"`
	SendIntervalSec        int  `json:"sendIntervalSec"`
	VisitStoreVersion      int  `json:"visitStoreVersion"`
	MaxCachedCrashesCount  int  `json:"maxCachedCrashesCount"`
}
type ReplayConfig struct {
	Capture                     bool `json:"capture"`
	ImageRetentionTimeInMinutes int  `json:"imageRetentionTimeInMinutes"`
}
type AppConfig struct {
	Capture                  int          `json:"capture"`
	CaptureLifecycle         int          `json:"captureLifecycle"`
	ReportCrashes            int          `json:"reportCrashes"`
	ReportErrors             int          `json:"reportErrors"`
	TrafficControlPercentage int          `json:"trafficControlPercentage"`
	ReplayConfig             ReplayConfig `json:"replayConfig"`
	ApplicationID            string       `json:"applicationId"`
}
type DynamicConfig struct {
	ServerID     int    `json:"serverId"`
	SwitchServer bool   `json:"switchServer"`
	Multiplicity int    `json:"multiplicity"`
	Status       string `json:"status"`
}

type ResponseParser struct {
	log *log.Logger
}

func NewResponseParser(log *log.Logger) ResponseParser {
	return ResponseParser{
		log: log,
	}
}

func (p *ResponseParser) ParseResponse(response string) ResponseAttributes {
	if response == KEY_VALUE_RESPONSE_TYPE_MOBILE || strings.HasPrefix(response, KEY_VALUE_RESPONSE_TYPE_MOBILE_WITH_SEPARATOR) {
		return p.parseKeyValuePair(response)
	}
	return p.parseJson(response)

}

func (p *ResponseParser) parseKeyValuePair(response string) ResponseAttributes {
	keyValuePairs := make(map[string]string)

	pairs := strings.Split(response, "&")
	for _, pair := range pairs {
		tuple := strings.Split(pair, "=")
		if len(tuple) == 2 {
			keyValuePairs[tuple[0]] = tuple[1]
		}
	}

	r := DefaultResponseAttributes()

	if val, ok := keyValuePairs[RESPONSE_KEY_MAX_BEACON_SIZE_IN_KB]; ok {
		i, err := strconv.Atoi(val)
		if err != nil {
			r.MaxBeaconSizeInBytes = i * 1024
		}
	}

	if val, ok := keyValuePairs[RESPONSE_KEY_SEND_INTERVAL_IN_SEC]; ok {
		i, err := strconv.Atoi(val)
		if err != nil {
			r.SendInterval = time.Duration(i) * time.Second
		}
	}

	if val, ok := keyValuePairs[RESPONSE_KEY_CAPTURE]; ok {
		r.Capture = val == "1"
	}

	if val, ok := keyValuePairs[RESPONSE_KEY_REPORT_CRASHES]; ok {
		r.CaptureCrashes = val != "0"
	}

	if val, ok := keyValuePairs[RESPONSE_KEY_REPORT_ERRORS]; ok {
		r.CaptureErrors = val != "0"
	}

	if val, ok := keyValuePairs[RESPONSE_KEY_TRAFFIC_CONTROL_PERCENTAGE]; ok {
		i, err := strconv.Atoi(val)
		if err != nil {
			r.TrafficControlPercentage = i
		}
	}

	if val, ok := keyValuePairs[RESPONSE_KEY_SERVER_ID]; ok {
		i, err := strconv.Atoi(val)
		if err != nil {
			r.ServerID = i
		}
	}

	if val, ok := keyValuePairs[RESPONSE_KEY_MULTIPLICITY]; ok {
		i, err := strconv.Atoi(val)
		if err != nil {
			r.Multiplicity = i
		}
	}
	return r
}

func (p *ResponseParser) parseJson(response string) ResponseAttributes {
	r := DefaultResponseAttributes()
	jsonResponse := JsonResponse{}

	if err := json.Unmarshal([]byte(response), &jsonResponse); err == nil {

		// Agent config
		agentConfig := jsonResponse.MobileAgentConfig
		if agentConfig.MaxBeaconSizeKb != 0 {
			r.MaxBeaconSizeInBytes = agentConfig.MaxBeaconSizeKb
		}
		if agentConfig.MaxSessionDurationMins != 0 {
			r.MaxSessionDuration = time.Duration(agentConfig.MaxSessionDurationMins) * time.Minute
		}
		if agentConfig.MaxEventsPerSession != 0 {
			r.MaxEventsPerSession = agentConfig.MaxEventsPerSession
		}
		if agentConfig.SendIntervalSec != 0 {
			r.SendInterval = time.Duration(agentConfig.SendIntervalSec) * time.Second
		}
		if agentConfig.SessionTimeoutSec != 0 {
			r.SessionTimeout = time.Duration(agentConfig.SessionTimeoutSec) * time.Second
		}
		if agentConfig.VisitStoreVersion != 0 {
			r.VisitStoreVersion = agentConfig.VisitStoreVersion
		}

		// Application
		appConfig := jsonResponse.AppConfig
		if appConfig.ApplicationID != "" {
			r.ApplicationID = appConfig.ApplicationID
		}
		r.Capture = appConfig.Capture == 1
		r.CaptureCrashes = appConfig.ReportCrashes != 0
		r.CaptureErrors = appConfig.ReportErrors != 0
		r.TrafficControlPercentage = appConfig.TrafficControlPercentage

		// Dynamic
		dynConfig := jsonResponse.DynamicConfig
		if dynConfig.Status != "" {
			r.Status = dynConfig.Status
		}
		if dynConfig.ServerID != 0 {
			r.ServerID = dynConfig.ServerID
		}
		if dynConfig.Multiplicity != 0 {
			r.Multiplicity = dynConfig.Multiplicity
		}

		// Root
		if jsonResponse.Timestamp != 0 {
			r.Timestamp = time.Unix(jsonResponse.Timestamp/1000, 0)
		}
	} else {
		p.log.WithFields(log.Fields{"error": err.Error()}).Error("could not parse the JSON response")
	}

	return r

}
