package main

import (
	"encoding/hex"
)

type EQBand struct {
	RawBytes        string
	BandWidth       FaderValue
	CenterFrequency FaderValue
	Gain            FaderValue
}

func GetEQBand(bytes []byte) EQBand {
	return EQBand{
		RawBytes:        hex.EncodeToString(bytes),
		BandWidth:       GetFaderValue(bytes[0x04:0x06]),
		CenterFrequency: GetFaderValue(bytes[0x02:0x04]),
		Gain:            GetFaderValue(bytes[0x00:0x02])}
}
