package configuration

import "net/http"

type HttpClientConfiguration struct {
	BaseURL       string
	ServerID      int
	ApplicationID string
	Transport     *http.Transport
}
