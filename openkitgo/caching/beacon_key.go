package caching

import "fmt"

type BeaconKey struct {
	BeaconId    int32
	BeaconSeqNo int32
}

func NewBeaconKey(beaconId int32, beaconSeqNo int32) BeaconKey {
	return BeaconKey{
		BeaconId:    beaconId,
		BeaconSeqNo: beaconSeqNo,
	}
}

func (k *BeaconKey) String() string {
	return fmt.Sprintf("BeaconKey(sn=%d, seq=%d)", k.BeaconId, k.BeaconSeqNo)
}
