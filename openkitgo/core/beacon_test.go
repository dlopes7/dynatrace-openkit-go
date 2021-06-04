package core

import (
	"testing"
	"time"
)

func TestAddAction(t *testing.T) {

	session := NewSession(logger, nil, beacon)
	action := NewAction(logger, session, "Action", beacon, time.Now())
	beacon.AddActionAt(action, time.Now().Add(10*time.Minute))

}
