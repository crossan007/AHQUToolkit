package main

import (
	"encoding/binary"
	"encoding/hex"
)

type FaderValue struct {
	RawBytes     string
	IntValue     int
	PercentValue float32
}

func GetFaderValue(bytes []byte) FaderValue {
	var intV = int(binary.LittleEndian.Uint16(bytes))
	return FaderValue{
		RawBytes:     hex.EncodeToString(bytes),
		IntValue:     int(binary.LittleEndian.Uint16(bytes)),
		PercentValue: (float32(intV) / 32768) * 100}
}
