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

func (sp SystemPacket) String() string {
	if sp.groupid == 0 {
		var port = int(binary.LittleEndian.Uint16(sp.data))
		return fmt.Sprintf("Received remote control UDP listening port: %d", port)
	} else if sp.groupid == 4 {
		return "Received heartbeat packet"
	} else {
		return fmt.Sprintf("System Packet:   GroupID: %d; length: %d; data:%s", sp.groupid, len(sp.data), hex.EncodeToString(sp.data))
	}
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
