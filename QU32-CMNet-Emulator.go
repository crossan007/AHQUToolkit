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
	length  int
	data    []byte
}

func (sp *SystemPacket) String() string {
	if sp.groupid == 0 {
		var port = int(binary.LittleEndian.Uint16(sp.data))
		return fmt.Sprintf("Received remote control UDP listening port: %d", port)
	} else if sp.groupid == 4 {
		return "Received heartbeat packet"
	} else {
		return fmt.Sprintf("Received System Packet.  GroupID: %d; length: %d; data:%s", sp.groupid, sp.length, sp.data)
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
	WriteMixerHandshakeResponse(conn)
	for {
		sp1, err1 := ReadSystemPacket(conn)
		if err1 != nil {
			fmt.Println("Error reading system packet: " + err1.Error())
			return
		} else {
			fmt.Println(sp1)
		}

	}

}

func WriteMixerHandshakeResponse(conn net.Conn) {
	var packets [2]SystemPacket
	response, _ := hex.DecodeString("00c0") // this tells the remote client the mixer's UDP listening port: 49152

	packets[0] = SystemPacket{
		groupid: 0,
		data:    response,
		length:  len(response)}

	response2, _ := hex.DecodeString("03015f01d111000000000000")

	packets[1] = SystemPacket{
		groupid: 01,
		data:    response2,
		length:  len(response2)}
	WriteSystemPacket(packets[:], conn)
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
		length:  len,
		data:    buf3}, nil
}

func WriteSystemPacket(sp []SystemPacket, conn net.Conn) {
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

func SystemPacketToByteArray(sp SystemPacket) (bytes []byte) {
	var len = 4 + len(sp.data)
	outbuf := make([]byte, len)
	outbuf[0] = 0x7f
	outbuf[1] = byte(sp.groupid)
	outbuf[2] = byte(sp.length)
	outbuf[3] = 0
	copy(outbuf[4:], sp.data)
	return outbuf
}
