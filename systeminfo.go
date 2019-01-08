package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	"net"
)

type SystemPacket struct {
	groupid byte
	data    []byte
}

type Mixer struct {
	Name        string
	Version     MixerVersion
	DSPChannels []DSPChannel
}

type MixerVersion struct {
	// Board Type can be 01 for QU-16, 02 for QU-24, or 03 for QU-32
	BoardType     uint16
	FirmwareMajor uint16
	FirmwareMinor uint16
	FirmwarePatch uint16
}

func (sp SystemPacket) String() string {
	if sp.groupid == 0 {
		var port = int(binary.LittleEndian.Uint16(sp.data))
		return fmt.Sprintf("Received remote control UDP listening port: %d", port)
	} else if sp.groupid == 4 {
		return "Received heartbeat packet"
	} else {
		return fmt.Sprintf("System Packet:   GroupID: %d; length: %d; data:%s", sp.groupid, len(sp.data), sp.data)
	}
}

func (mv MixerVersion) ToBytes() []byte {
	//mixer version [(1)board - 0x01=16; 0x02=24; 0x03=32][(1)major 0x01 ][(2)minor 0x5f00 ][(2)patch 0xd111]
	outbuf := make([]byte, 12)
	response := make([]byte, 2)
	binary.LittleEndian.PutUint16(response, mv.BoardType)
	outbuf[0] = response[0]
	binary.LittleEndian.PutUint16(response, mv.FirmwareMajor)
	outbuf[1] = response[0]
	binary.LittleEndian.PutUint16(response, mv.FirmwareMinor)
	outbuf[2] = response[0]
	binary.LittleEndian.PutUint16(response, mv.FirmwarePatch)
	copy(outbuf[4:], response)
	return outbuf
}

func GetUDPPortSystemPacket(portNumber int) (sp SystemPacket) {

	UDPPort := uint16(portNumber)
	response := make([]byte, 2)
	binary.LittleEndian.PutUint16(response, UDPPort)
	//response, _ := hex.DecodeString("00c0") // this tells the remote client the mixer's UDP listening port: 49152
	packet := SystemPacket{
		groupid: 0x00,
		data:    response}
	return packet
}

func CreateSystemPacketFromHexString(groupId byte, data string) (sp SystemPacket) {
	byteArray, e := hex.DecodeString(data)
	check(e)
	return SystemPacket{groupid: groupId, data: byteArray}
}

func WriteSystemPackets(sp []SystemPacket, conn net.Conn) {
	var outbuf []byte
	for _, onepacket := range sp {
		packetbytes := SystemPacketToByteArray(onepacket)
		outbuf = append(outbuf, packetbytes...)
	}
	_, err := conn.Write(outbuf)
	if err != nil {
		log.Println("Error writing to connection")
	}
	//log.Println("Wrote System packet: " + hex.EncodeToString(outbuf))
}
func WriteSystemPacket(sp SystemPacket, conn net.Conn) {
	outbuf := SystemPacketToByteArray(sp)
	_, err := conn.Write(outbuf)
	if err != nil {
		log.Println("Error writing to connection")
	}
	//log.Println("Wrote System packet: " + hex.EncodeToString(outbuf))
}

func SystemPacketToByteArray(sp SystemPacket) (bytes []byte) {
	var packetByteCount = 4 + len(sp.data)
	outbuf := make([]byte, packetByteCount)
	// set the packet header and groupid
	outbuf[0] = 0x7f
	outbuf[1] = sp.groupid
	//set the data length in Little Endian encoding.
	lengthBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(lengthBytes, uint16(len(sp.data)))
	copy(outbuf[2:], lengthBytes)
	// copy the data to the output buffer.
	copy(outbuf[4:], sp.data)
	return outbuf
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
