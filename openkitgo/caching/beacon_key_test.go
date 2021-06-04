package caching

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewBeaconKey(t *testing.T) {
	key := uint32(1)
	seq := uint32(2)
	k := NewBeaconKey(key, seq)
	assert.Equal(t, k.BeaconId, key)
	assert.Equal(t, k.BeaconSeqNo, seq)

}
