package core

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestExecuteCurrentState(t *testing.T) {
	ctx.executeCurrentState()
	assert.True(t, ctx.initOk)
}
