package main

import (
	"fmt"
	"strconv"
)

type DSPPacket struct {
	ControlID   byte
	TargetGroup DSPTargetGroup
	ValueID     byte
	ClientID    byte
	Parameter1  byte
	Parameter2  byte
	Value       byte
}

func (sp DSPPacket) String() string {

	return fmt.Sprintf("DSP Packet: ControlID: %d; TargetGroup: %s; ValueID: %d, ClientID: %d, Parameter1: %d, Parameter2: %d, Value: %d",
		int(sp.ControlID),
		sp.TargetGroup,
		int(sp.ValueID),
		int(sp.ClientID),
		int(sp.Parameter1),
		int(sp.Parameter2),
		int(sp.Value))
}

type DSPTargetGroup byte

const (
	MIX           = 0x04
	PARAMETRIC_EQ = 0x05
	GAIN          = 0x06
	GATE          = 0x07
	COMPRESSION   = 0x08
)

func (s DSPTargetGroup) String() string {
	var name string
	switch s {
	case MIX:
		name = "Mix"
	case PARAMETRIC_EQ:
		name = "Parametric Eq"
	case GAIN:
		name = "Gain"
	case GATE:
		name = "Gate"
	case COMPRESSION:
		name = "Compression"
	}
	return name + ": " + strconv.Itoa(int(s))
}
