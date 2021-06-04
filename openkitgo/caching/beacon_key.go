package caching

import "fmt"

type BeaconKey struct {
	BeaconId    uint32
	BeaconSeqNo uint32
}

func NewBeaconKey(beaconId uint32, beaconSeqNo uint32) BeaconKey {
	return BeaconKey{
		BeaconId:    beaconId,
		BeaconSeqNo: beaconSeqNo,
	}
}

func (k *BeaconKey) String() string {
	return fmt.Sprintf("BeaconKey(sn=%d, seq=%d)", k.BeaconId, k.BeaconSeqNo)
}
