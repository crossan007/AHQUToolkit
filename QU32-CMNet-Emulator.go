package main

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"strconv"
)

type SystemPacket struct {
	groupid int
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
		handleClient(conn)
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
	WriteSystemPacket(GetUDPPortSystemPacket(41952), conn)
	response1, _ := hex.DecodeString("03015f01d111000000000000")
	WriteSystemPacket(SystemPacket{groupid: 01, data: response1}, conn)
	for i := 0; i < 1; i++ {
		sp1, err1 := ReadSystemPacket(conn)
		if err1 != nil {
			fmt.Println("Error reading system packet: " + err1.Error())
			return
		} else {
			fmt.Println(sp1)
		}
	}
	response2, _ := hex.DecodeString("13000000ffffffffffff9f0f0000000000000000000000000000000000e003c0ffffff7f")
	WriteSystemPacket(SystemPacket{groupid: 04, data: response2}, conn)

	response3, _ := hex.DecodeString("15000000fac1230604000028d1041c000000000000000000000000000000000000000000")
	WriteSystemPacket(SystemPacket{groupid: 04, data: response3}, conn)

	response4, _ := hex.DecodeString("02007f0402000f00")
	WriteSystemPacket(SystemPacket{groupid: 04, data: response4}, conn)

	response5, _ := hex.DecodeString("07007f04020006007f0404001b640000") //87
	WriteSystemPacket(SystemPacket{groupid: 04, data: response5}, conn)

	response5b, _ := hex.DecodeString("1a640000")
	WriteSystemPacket(SystemPacket{groupid: 04, data: response5b}, conn)

	response6, _ := hex.DecodeString("19640000")
	WriteSystemPacket(SystemPacket{groupid: 04, data: response6}, conn)

	response7, _ := hex.DecodeString("1100")
	WriteSystemPacket(SystemPacket{groupid: 04, data: response7}, conn)

	response8, _ := hex.DecodeString("12007f0402001000")
	WriteSystemPacket(SystemPacket{groupid: 04, data: response8}, conn)

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
		groupid: 0,
		data:    response}
	return packet
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
	var group = int(buf2[0])
	var len = int(buf2[1])
	fmt.Println("groupid: " + strconv.Itoa(group))
	fmt.Println("len: " + strconv.Itoa(len))

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
	fmt.Println("Wrote System packet: " + hex.EncodeToString(outbuf))
}
func WriteSystemPacket(sp SystemPacket, conn net.Conn) {
	outbuf := SystemPacketToByteArray(sp)
	_, err := conn.Write(outbuf)
	if err != nil {
		fmt.Println("Error writing to connection")
	}
	fmt.Println("Wrote System packet: " + hex.EncodeToString(outbuf))
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
