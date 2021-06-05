package core

import (
	"crypto/tls"
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo/caching"
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo/configuration"
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo/providers"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"testing"
	"time"
)

var beacon *Beacon
var logger *log.Logger
var httpClient HttpClient
var ctx *BeaconSendingContext

func TestMain(m *testing.M) {

	logger = log.New()
	logger.SetLevel(log.DebugLevel)

	httpClientConfig := &configuration.HttpClientConfiguration{
		BaseURL:       "https://localhost:9999/mbeacon/e/eaa50379",
		ServerID:      1,
		ApplicationID: "98972aef-02ac-4ecb-be1e-a6698af2de60",
		Transport:     &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
	}
	ctx = NewBeaconSendingContext(logger, httpClientConfig)
	httpClient = NewHttpClient(logger, httpClientConfig)

	o := &configuration.OpenKitConfiguration{}
	p := &configuration.PrivacyConfiguration{
		DataCollectionLevel: configuration.DATA_USER_BEHAVIOR,
		CrashReportingLevel: configuration.CRASH_OPT_IN_CRASHES,
	}
	s := configuration.DefaultServerConfiguration()
	s.Capture = true
	s.TrafficControlPercentage = 100
	s.Multiplicity = 1
	c := configuration.NewBeaconConfiguration(o, p, 1)
	c.ServerConfiguration = s

	beacon = NewBeacon(logger,
		caching.NewBeaconCache(logger),
		providers.NewSessionIDProvider(),
		NewSessionProxy(),
		c,
		time.Now(),
		1,
		"")

	os.Exit(m.Run())
}
