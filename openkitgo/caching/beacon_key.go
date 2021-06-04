package caching

import "fmt"

type BeaconKey struct {
	beaconId    uint32
	beaconSeqNo uint32
}

func NewBeaconKey(beaconId uint32, beaconSeqNo uint32) BeaconKey {
	return BeaconKey{
		beaconId:    beaconId,
		beaconSeqNo: beaconSeqNo,
	}
}

func (k *BeaconKey) String() string {
	return fmt.Sprintf("BeaconKey(sn=%d, seq=%d)", k.beaconId, k.beaconSeqNo)
}
