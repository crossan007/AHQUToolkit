package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io/ioutil"
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

type DSPChannel struct {
	// each DSP channel seems to be 0xc0 bytes long, begins with 0x802ECB, and ends with 0x0b0000
	// offset 0x6c into each block seems to be the main send level.
	Name          string
	MainSendLevel uint16
	Gain          uint16
	GateAttack    uint16
	GateRelease   uint16
	GateHold      uint16
	GateThreshold uint16
	GateDepth     uint16
}

func GetDSPDataSystemPacket() (sp SystemPacket) {
	var channelData [0x6520]byte
	preamble, _ := hex.DecodeString("B500FEFF035FD1110100000049676C657369610000000000000000000000000000000000000000000000000001A5A5A5EA7F02591500008087740B001082B0A4120000")
	copy(channelData[0:], preamble)
	for i := 0; i < 47; i++ {
		var bytes = GetDSPChannelBytes(thisMixer.DSPChannels[i])
		copy(channelData[len(preamble)+i*0xc0:], bytes)
	}
	channelData1, err := ioutil.ReadFile("DSPChannelPostamble.bin")
	check(err)
	copy(channelData[0x2c83:], channelData1)
	//log.Println("DSPData Bytes: " + hex.EncodeToString(channelData[:]))
	return SystemPacket{groupid: 0x06, data: channelData[:]}
}

func GetDSPChannelBytes(Channel DSPChannel) []byte {
	var DSPChannelBytes [0xc0]byte

	byteArray1, _ := hex.DecodeString("802ECB0B00000001003F55787800000B00008000800100")
	copy(DSPChannelBytes[0:], byteArray1)
	copy(DSPChannelBytes[0x17:], GetLittleEndianBytes(Channel.GateAttack))
	copy(DSPChannelBytes[0x19:], GetLittleEndianBytes(Channel.GateRelease))
	copy(DSPChannelBytes[0x21:], GetLittleEndianBytes(Channel.GateHold))
	copy(DSPChannelBytes[0x23:], GetLittleEndianBytes(Channel.GateThreshold))
	copy(DSPChannelBytes[0x25:], GetLittleEndianBytes(Channel.GateDepth))
	byteArray1b, _ := hex.DecodeString("000000800080008000800080008000800080008000800080008000800080008000800080008000800080008000800080008000800080008000800000000000000100A049010000000000")
	copy(DSPChannelBytes[0x27:], byteArray1b)
	copy(DSPChannelBytes[0x6b:], GetLittleEndianBytes(Channel.MainSendLevel))
	copy(DSPChannelBytes[0x6d:], GetLittleEndianBytes(Channel.Gain))
	byteArray2, _ := hex.DecodeString("008000000000000000000100008000000000FFFFFFFF009C0000")
	copy(DSPChannelBytes[0x6f:], byteArray2)
	copy(DSPChannelBytes[0x89:], []byte(Channel.Name))
	byteArray3, _ := hex.DecodeString("00000000987E6389607B005C00940000000000000000000000000000008097470B00008097650B0000802EA70B0000")
	copy(DSPChannelBytes[0x91:], byteArray3)
	return DSPChannelBytes[:]
}

func GetLittleEndianBytes(i uint16) []byte {
	bytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(bytes, uint16(i))
	return bytes
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
