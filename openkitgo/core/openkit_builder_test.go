package core

import (
	"github.com/dlopes7/dynatrace-openkit-go/openkitgo/configuration"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBuild(t *testing.T) {

	b := NewOpenKitBuilder("https://localhost", "", 1)
	b.Build()

	assert.Equal(t, configuration.DATA_USER_BEHAVIOR, b.(*OpenKitBuilder).dataCollectionLevel)

}
