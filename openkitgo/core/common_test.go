package core

import (
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo/caching"
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo/configuration"
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo/providers"
	log "github.com/sirupsen/logrus"
	"os"
	"testing"
	"time"
)

var beacon *Beacon
var logger *log.Logger

func TestMain(m *testing.M) {

	logger = log.New()
	logger.SetLevel(log.DebugLevel)

	o := configuration.NewOpenKitConfiguration(nil)
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
