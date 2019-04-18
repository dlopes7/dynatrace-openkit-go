package openkitgo

import (
	"bytes"
	"compress/gzip"
	"github.com/op/go-logging"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

var QUERY_RESERVED_CHARACTERS = []rune{'_'}

const (
	// request type constants
	REQUEST_TYPE_MOBILE                     = "type=m"
	REQUEST_TYPE_MOBILE_WITH_ARGS_SEPARATOR = "type=m&"

	// query parameter constants
	QUERY_KEY_SERVER_ID             = "srvid"
	QUERY_KEY_APPLICATION           = "app"
	QUERY_KEY_VERSION               = "va"
	QUERY_KEY_PLATFORM_TYPE         = "pt"
	QUERY_KEY_AGENT_TECHNOLOGY_TYPE = "tt"
	QUERY_KEY_NEW_SESSION           = "ns"

	// additional reserved characters for URL encoding

	// connection constants
	MAX_SEND_RETRIES = 3
	RETRY_SLEEP_TIME = 200 // retry sleep time in ms
	CONNECT_TIMEOUT  = 5000
	READ_TIMEOUT     = 30000

	REQUESTTYPE_STATUS       = "Status"
	REQUEST_TYPE_BEACON      = "Beacon"
	REQUEST_TYPE_NEW_SESSION = "NewSession"
)

type HttpClient struct {
	monitorURL    string
	newSessionURL string
	serverID      int
	logger        logging.Logger
}

func NewHttpClient(logger logging.Logger, configuration HTTPClientConfiguration) *HttpClient {
	httpClient := new(HttpClient)

	httpClient.logger = logger
	httpClient.serverID = configuration.serverID
	httpClient.monitorURL = buildMonitorURL(configuration.baseURL, configuration.applicationID, httpClient.serverID)
	httpClient.newSessionURL = buildNewSessionURL(configuration.baseURL, configuration.applicationID, httpClient.serverID)

	return httpClient

}

func buildMonitorURL(baseURL string, applicationID string, serverID int) string {
	var monitorURLBuilder strings.Builder

	monitorURLBuilder.WriteString(baseURL)
	monitorURLBuilder.WriteString("?")
	monitorURLBuilder.WriteString(REQUEST_TYPE_MOBILE)

	appendQueryParam(&monitorURLBuilder, QUERY_KEY_SERVER_ID, strconv.Itoa(serverID))
	appendQueryParam(&monitorURLBuilder, QUERY_KEY_APPLICATION, applicationID)
	appendQueryParam(&monitorURLBuilder, QUERY_KEY_VERSION, OPENKIT_VERSION)
	appendQueryParam(&monitorURLBuilder, QUERY_KEY_PLATFORM_TYPE, strconv.Itoa(PLATFORM_TYPE_OPENKIT))
	appendQueryParam(&monitorURLBuilder, QUERY_KEY_AGENT_TECHNOLOGY_TYPE, AGENT_TECHNOLOGY_TYPE)

	return monitorURLBuilder.String()
}

func buildNewSessionURL(baseURL string, applicationID string, serverID int) string {

	var monitorURLBuilder strings.Builder
	monitorURLBuilder.WriteString(buildMonitorURL(baseURL, applicationID, serverID))

	appendQueryParam(&monitorURLBuilder, QUERY_KEY_NEW_SESSION, "1")
	return monitorURLBuilder.String()
}

func (c *HttpClient) sendNewSessionRequest() *StatusResponse {
	c.logger.Debug("sendNewSessionRequest()")
	response, err := c.sendRequest(REQUEST_TYPE_NEW_SESSION, c.newSessionURL, nil, nil, "GET")
	if err != nil {
		c.logger.Errorf("Error getting response for sendNewSessionRequest: %s\n", err.Error())
		return nil
	}

	return response
}

func (c *HttpClient) sendRequest(requestType string, url string, clientIPAddress *string, data []byte, method string) (*StatusResponse, error) {
	c.logger.Debugf("sendRequest() - HTTP %s Request: %s", requestType, url)

	client := http.Client{}
	var buf bytes.Buffer

	if data != nil {
		g := gzip.NewWriter(&buf)

		if _, err := g.Write(data); err != nil {
			c.logger.Error(err.Error())
			return nil, err
		}
		if err := g.Close(); err != nil {
			c.logger.Error(err.Error())
			return nil, err
		}

	}

	request, err := http.NewRequest(method, url, &buf)
	if err != nil {
		c.logger.Error(err.Error())
		return nil, err
	}

	if clientIPAddress != nil {
		request.Header.Add("X-Client-IP", *clientIPAddress)
	}

	resp, err := client.Do(request)
	if err != nil {
		c.logger.Error(err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	var bodyString string
	if resp.StatusCode == http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		bodyString = string(bodyBytes)
	}

	return NewStatusResponse(c.logger, bodyString, resp.StatusCode, resp.Header), nil

}

func appendQueryParam(sb *strings.Builder, key string, value string) {
	sb.WriteString("&")
	sb.WriteString(key)
	sb.WriteString("=")

	sb.WriteString(encodeWithReservedChars(value, "UTF-8", QUERY_RESERVED_CHARACTERS))

}

type Response struct{}
