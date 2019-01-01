package main

type DSPPacket struct {
	ControlID   byte
	TargetValue byte
	ValueID     byte
	ClientID    byte
	Parameter1  byte
	Parameter2  byte
	Value       byte
}

func (sp DSPPacket) String() string {
	return "hey, this is a DSPPacket"
}
