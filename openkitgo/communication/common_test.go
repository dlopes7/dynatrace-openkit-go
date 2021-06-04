package communication

import (
	"crypto/tls"
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo/configuration"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"testing"
)

var httpClient HttpClient
var ctx *BeaconSendingContext

func TestMain(m *testing.M) {
	l := log.New()
	l.SetLevel(log.DebugLevel)
	c := configuration.HttpClientConfiguration{
		BaseURL:       "https://localhost:9999/mbeacon/e/eaa50379",
		ServerID:      1,
		ApplicationID: "98972aef-02ac-4ecb-be1e-a6698af2de60",
		Transport:     &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
	}
	ctx = NewBeaconSendingContext(l, c)
	httpClient = NewHttpClient(l, c)
	os.Exit(m.Run())
}
