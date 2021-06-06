package core

import (
	"bytes"
	"compress/gzip"
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo/configuration"
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo/protocol"
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo/utils"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

type RequestType int

const (
	STATUS      RequestType = 0
	BEACON      RequestType = 1
	NEW_SESSION RequestType = 2

	REQUEST_TYPE_MOBILE             = "type=m"
	QUERY_KEY_SERVER_ID             = "srvid"
	QUERY_KEY_APPLICATION           = "app"
	QUERY_KEY_VERSION               = "va"
	QUERY_KEY_PLATFORM_TYPE         = "pt"
	QUERY_KEY_AGENT_TECHNOLOGY_TYPE = "tt"
	QUERY_KEY_RESPONSE_TYPE         = "resp"
	QUERY_KEY_CONFIG_TIMESTAMP      = "cts"
	QUERY_KEY_NEW_SESSION           = "ns"

	QUERY_RESERVED_CHARACTERS = "_"

	MAX_SEND_RETRIES = 3
	RETRY_SLEEP_TIME = 200
	CONNECT_TIMEOUT  = 5000
	READ_TIMEOUT     = 30000
)

type HttpClient struct {
	monitorURL    string
	newSessionURL string
	serverID      int
	log           *log.Logger
	parser        protocol.ResponseParser

	transport *http.Transport

	requestTypes []string
}

func NewHttpClient(log *log.Logger, config *configuration.HttpClientConfiguration) HttpClient {
	return HttpClient{
		monitorURL:    buildMonitorURL(config.BaseURL, config.ApplicationID, config.ServerID),
		newSessionURL: buildNewSessionURL(config.BaseURL, config.ApplicationID, config.ServerID),
		serverID:      config.ServerID,
		log:           log,
		parser:        protocol.NewResponseParser(log),
		transport:     config.Transport,
		requestTypes:  []string{"Status", "Beacon", "NewSession"},
	}
}

func buildMonitorURL(baseUrl string, applicationID string, serverID int) string {

	var b strings.Builder

	b.WriteString(baseUrl)
	b.WriteRune('?')
	b.WriteString(REQUEST_TYPE_MOBILE)

	appendQueryParam(&b, QUERY_KEY_SERVER_ID, strconv.Itoa(serverID))
	appendQueryParam(&b, QUERY_KEY_APPLICATION, applicationID)
	appendQueryParam(&b, QUERY_KEY_VERSION, protocol.OPENKIT_VERSION)
	appendQueryParam(&b, QUERY_KEY_PLATFORM_TYPE, strconv.Itoa(protocol.PLATFORM_TYPE_OPENKIT))
	appendQueryParam(&b, QUERY_KEY_AGENT_TECHNOLOGY_TYPE, protocol.AGENT_TECHNOLOGY_TYPE)
	appendQueryParam(&b, QUERY_KEY_RESPONSE_TYPE, protocol.RESPONSE_TYPE)

	return b.String()
}

func buildNewSessionURL(baseUrl string, applicationID string, serverID int) string {
	var b strings.Builder
	b.WriteString(buildMonitorURL(baseUrl, applicationID, serverID))
	appendQueryParam(&b, QUERY_KEY_NEW_SESSION, "1")
	return b.String()
}

func appendQueryParam(b *strings.Builder, key string, value string) {
	b.WriteRune('&')
	b.WriteString(key)
	b.WriteRune('=')
	b.WriteString(utils.PercentEncode(value))
}

func (h *HttpClient) SendStatusRequest(ctx *BeaconSendingContext) protocol.StatusResponse {

	var b strings.Builder
	b.WriteString(h.monitorURL)
	h.appendAdditionalQueryParameters(&b, ctx)

	statusUrl := b.String()
	r, err := h.sendRequest(STATUS, statusUrl, "", nil, "GET")
	if err != nil {
		return protocol.NewStatusResponse(h.log, protocol.UndefinedResponseAttributes(), 999, nil)
	}
	return *r
}

func (h *HttpClient) SendNewSessionRequest(ctx *BeaconSendingContext) protocol.StatusResponse {

	var b strings.Builder
	b.WriteString(h.newSessionURL)
	h.appendAdditionalQueryParameters(&b, ctx)
	r, err := h.sendRequest(NEW_SESSION, b.String(), "", nil, "GET")
	if err != nil {
		return protocol.NewStatusResponse(h.log, protocol.UndefinedResponseAttributes(), -1, nil)
	}
	return *r

}

func (h *HttpClient) sendBeaconRequest(clientIPAddress string, data []byte, ctx *BeaconSendingContext) protocol.StatusResponse {

	var b strings.Builder
	b.WriteString(h.monitorURL)
	h.appendAdditionalQueryParameters(&b, ctx)
	r, err := h.sendRequest(BEACON, b.String(), clientIPAddress, data, "POST")
	if err != nil {
		return protocol.NewStatusResponse(h.log, protocol.UndefinedResponseAttributes(), -1, nil)
	}
	return *r

}

func (h *HttpClient) sendRequest(requestType RequestType, url string, clientIPAddress string, data []byte, method string) (*protocol.StatusResponse, error) {
	h.log.WithFields(log.Fields{"type": h.requestTypes[requestType], "url": url, "method": method}).Debug("sendRequest")

	client := http.Client{Transport: h.transport}

	var buf bytes.Buffer

	if data != nil {
		h.log.WithFields(log.Fields{"data": string(data)}).Debug("Beacon Body")
		g := gzip.NewWriter(&buf)

		if _, err := g.Write(data); err != nil {
			h.log.Error(err.Error())
			return nil, err
		}
		if err := g.Close(); err != nil {
			h.log.Error(err.Error())
			return nil, err
		}
	}

	request, err := http.NewRequest(method, url, &buf)
	if err != nil {
		h.log.Error(err.Error())
		return nil, err
	}

	if clientIPAddress != "" {
		request.Header.Add("X-Client-IP", clientIPAddress)
	}

	resp, err := client.Do(request)
	if err != nil {
		h.log.Error(err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		h.log.WithFields(log.Fields{"response": resp.Status}).Warning("Bad response from OpenKit")
	}

	var bodyString string
	if resp.StatusCode == http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		bodyString = string(bodyBytes)
	}

	h.log.WithFields(log.Fields{"response": bodyString, "code": resp.Status}).Debug("HttpClient handle response")
	responseAttributes := h.parser.ParseResponse(bodyString)
	statusResponse := protocol.NewStatusResponse(h.log, responseAttributes, resp.StatusCode, resp.Header)

	return &statusResponse, nil
}

func (h *HttpClient) appendAdditionalQueryParameters(builder *strings.Builder, ctx *BeaconSendingContext) {
	t := ctx.GetConfigurationTimestamp()

	if !t.IsZero() {
		appendQueryParam(builder, QUERY_KEY_CONFIG_TIMESTAMP, strconv.FormatInt(utils.TimeToMillis(t), 10))
	}
}
