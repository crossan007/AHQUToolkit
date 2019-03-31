package main

import (
	"encoding/binary"
)

type RemoteControlVUMeter struct {
	Volume float32
	value  []byte
}

func toBytes(VUMeter RemoteControlVUMeter) []byte {

	VUBytes := make([]byte, 20)
	Volume := (.375 * 65535) + (.200*65535)*VUMeter.Volume
	VolumeBytes := make([]byte, 4)
	binary.LittleEndian.PutUint16(VolumeBytes, uint16(Volume))
	//hexBytes, _ := hex.DecodeString("0000000000000000000000000000000000000000")
	copy(VUBytes[4:], VolumeBytes)
	return VUBytes
}
