package main

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"strconv"
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
		return fmt.Sprintf("Received System Packet.  GroupID: %d; length: %d; data:%s", sp.groupid, len(sp.data), sp.data)
	}
}

func main() {
	ln, err := net.Listen("tcp", ":51326")
	if err != nil {
		// handle error
		fmt.Println("Error creating listener")
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
			fmt.Println("Error accepting connection")
		}
		fmt.Println("Incoming connection from: " + conn.RemoteAddr().String())
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	/*
		When a remote (QU-You; QU-Pad; MixingStationPro) connects, it will:
		* Send a system packet containing the UDP port on which it's listening
		* Send a heartbeat packet

		Then, the mixer's response will be
		* send a system packet with the UDP Port number (49152)
		* send a system packet with some data; not sure what it means: 03015f01d111000000000000

	*/
	for i := 0; i < 2; i++ {
		//read two system packets from the remote
		sp, err := ReadSystemPacket(conn)
		if err != nil {
			fmt.Println("Error reading system packet: " + err.Error())
			return
		}
		fmt.Println(sp)
	}
	// write the mixer handshake response
	WriteSystemPacket(GetUDPPortSystemPacket(49152), conn)
	WriteSystemPacket(CreateSystemPacketFromHexString(0x01, "03015f01d111000000000000"), conn)

	for i := 0; i < 1; i++ {
		sp1, err1 := ReadSystemPacket(conn)
		if err1 != nil {
			fmt.Println("Error reading system packet: " + err1.Error())
			return
		} else {
			fmt.Println(sp1)
		}
	}

	WriteSystemPacket(CreateSystemPacketFromHexString(0x01, "03015f01d111000000000000"), conn)

	channelData1, err := ioutil.ReadFile("ChannelData1.bin")
	check(err)
	WriteSystemPacket(SystemPacket{groupid: 0x06, data: channelData1}, conn)

	channelData2, err := ioutil.ReadFile("ChannelData2.bin")
	check(err)
	WriteSystemPacket(SystemPacket{groupid: 0x22, data: channelData2}, conn)

	WriteSystemPacket(CreateSystemPacketFromHexString(0x0b, "0000FF00"), conn)

	WriteSystemPacket(CreateSystemPacketFromHexString(0x0a, "00000000"), conn)

	WriteSystemPacket(CreateSystemPacketFromHexString(0x22, "0100"), conn)
	WriteSystemPacket(CreateSystemPacketFromHexString(0x21, "0100"), conn)
	WriteSystemPacket(CreateSystemPacketFromHexString(0x20, "0100"), conn)

	channelData3, err := ioutil.ReadFile("ChannelData3.bin")
	check(err)
	WriteSystemPacket(SystemPacket{groupid: 0x1a, data: channelData3}, conn) //groupid 1a

	channelData4, err := ioutil.ReadFile("ChannelData4.bin")
	check(err)
	WriteSystemPacket(SystemPacket{groupid: 0x1b, data: channelData4}, conn) //groupid 1b

	channelData5, err := ioutil.ReadFile("ChannelData5.bin")
	check(err)
	WriteSystemPacket(SystemPacket{groupid: 0x18, data: channelData5}, conn) //groupid 1b

	/* I think I was emulating the wrong side of the conversation with this block:
	WriteSystemPacket(CreateSystemPacketFromHexString(4, "13000000ffffffffffff9f0f0000000000000000000000000000000000e003c0ffffff7f"), conn)
	WriteSystemPacket(CreateSystemPacketFromHexString(4, "15000000fac1230604000028d1041c000000000000000000000000000000000000000000"), conn)
	WriteSystemPacket(CreateSystemPacketFromHexString(4, "0700"), conn)
	WriteSystemPacket(CreateSystemPacketFromHexString(4, "0600"), conn)
	WriteSystemPacket(CreateSystemPacketFromHexString(4, "1b640000"), conn)
	WriteSystemPacket(CreateSystemPacketFromHexString(4, "1a640000"), conn)
	WriteSystemPacket(CreateSystemPacketFromHexString(4, "19640000"), conn)
	WriteSystemPacket(CreateSystemPacketFromHexString(4, "1100"), conn)
	WriteSystemPacket(CreateSystemPacketFromHexString(4, "1200"), conn)
	WriteSystemPacket(CreateSystemPacketFromHexString(4, "1000"), conn)*/

	for i := 0; i < 10; i++ {
		sp1, err1 := ReadSystemPacket(conn)
		if err1 != nil {
			fmt.Println("Error reading system packet: " + err1.Error())
			return
		} else {
			fmt.Println(sp1)
		}
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
	byteArray, _ := hex.DecodeString(data)
	return SystemPacket{groupid: groupId, data: byteArray}
}

func ReadSystemPacket(conn net.Conn) (sp SystemPacket, err error) {
	var buf1 [1]byte

	n, err1 := conn.Read(buf1[0:])
	if err1 != nil {
		return sp, errors.New("Error reading connection buffer, read " + strconv.Itoa(n) + " bytes read")
	}
	if buf1[0] != 0x7f {
		return sp, errors.New("Expected header 0x07 for system packet; got: " + hex.EncodeToString(buf1[:]))
	}

	var buf2 [3]byte
	_, err2 := conn.Read(buf2[0:])
	if err2 != nil {
		return sp, errors.New("Error reading packet group or data length")
	}
	var group = buf2[0]
	var len = int(buf2[1])
	buf3 := make([]byte, len)
	_, err3 := conn.Read(buf3[0:])
	if err3 != nil {
		return sp, errors.New("Error reading system packet data")
	}

	return SystemPacket{
		groupid: group,
		data:    buf3}, nil
}

func WriteSystemPackets(sp []SystemPacket, conn net.Conn) {
	var outbuf []byte
	for _, onepacket := range sp {
		packetbytes := SystemPacketToByteArray(onepacket)
		outbuf = append(outbuf, packetbytes...)
	}
	_, err := conn.Write(outbuf)
	if err != nil {
		fmt.Println("Error writing to connection")
	}
	//fmt.Println("Wrote System packet: " + hex.EncodeToString(outbuf))
}
func WriteSystemPacket(sp SystemPacket, conn net.Conn) {
	outbuf := SystemPacketToByteArray(sp)
	_, err := conn.Write(outbuf)
	if err != nil {
		fmt.Println("Error writing to connection")
	}
	//fmt.Println("Wrote System packet: " + hex.EncodeToString(outbuf))
}

func SystemPacketToByteArray(sp SystemPacket) (bytes []byte) {
	var packetByteCount = 4 + len(sp.data)
	outbuf := make([]byte, packetByteCount)
	outbuf[0] = 0x7f
	outbuf[1] = byte(sp.groupid)
	outbuf[2] = byte(len(sp.data))
	outbuf[3] = 0
	copy(outbuf[4:], sp.data)
	return outbuf
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
