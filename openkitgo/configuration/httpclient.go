package configuration

import "net/http"

type HttpClient struct {
	BaseURL       string
	ServerID      int
	ApplicationID string
	Transport     *http.Transport
	// TODO httpRequestInterceptor
	// TODO httpResponseInterceptor
}
