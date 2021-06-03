package caching

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
