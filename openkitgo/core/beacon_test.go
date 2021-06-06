package core

import (
	"testing"
	"time"
)

func TestAddAction(t *testing.T) {

	session := NewSession(logger, nil, beacon, time.Now())
	action := NewAction(logger, session, nil, "Action", beacon, time.Now())
	beacon.AddActionAt(action, time.Now().Add(10*time.Minute))

}
