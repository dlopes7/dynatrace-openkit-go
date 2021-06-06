package caching

import (
	log "github.com/sirupsen/logrus"
	"os"
	"testing"
)

var logger *log.Logger

func TestMain(m *testing.M) {
	logger = log.New()
	logger.SetLevel(log.DebugLevel)
	os.Exit(m.Run())
}
