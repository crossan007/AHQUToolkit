package main

import (
	"encoding/hex"
)

type KnobValue struct {
	RawBytes     string
	IntValue     int
	PercentValue float32
}

func GetKnobValue(b byte) KnobValue {
	var intV = int(b)
	bs := make([]byte, 1)
	bs[0] = b
	return KnobValue{
		RawBytes:     hex.EncodeToString(bs),
		IntValue:     int(b),
		PercentValue: (float32(intV) / 32768) * 100}
}
