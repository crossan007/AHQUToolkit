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
		PrintSystemPacket(sp)
	}
	// write the mixer handshake response
	WriteMixerHandshakeResponse1(conn)
	WriteMixerHandshakeResponse2(conn)
	for {
		sp1, err1 := ReadSystemPacket(conn)
		if err1 != nil {
			fmt.Println("Error reading system packet: " + err1.Error())
			return
		} else {
			PrintSystemPacket(sp1)
		}

	}

}
func WriteMixerHandshakeResponse1(conn net.Conn) {
	response, err := hex.DecodeString("00c0")
	if err != nil {
	}
	sp := SystemPacket{
		groupid: 0,
		data:    response,
		length:  len(response)}
	WriteSystemPacket(sp, conn)
}
func WriteMixerHandshakeResponse2(conn net.Conn) {
	response, err := hex.DecodeString("03015f01d111000000000000")
	if err != nil {
	}
	sp := SystemPacket{
		groupid: 01,
		data:    response,
		length:  len(response)}
	WriteSystemPacket(sp, conn)
}

func PrintSystemPacket(sp SystemPacket) {
	if sp.groupid == 0 {
		var port = int(binary.LittleEndian.Uint16(sp.data))
		fmt.Println("Received System Packet.  GroupID: " + strconv.Itoa(sp.groupid) + "; length: " + strconv.Itoa(sp.length) + "; port:" + strconv.Itoa(port))
	} else if sp.groupid == 4 {
		fmt.Println("Received heartbeat packet")
	} else {
		fmt.Println("Received System Packet.  GroupID: " + strconv.Itoa(sp.groupid) + "; length: " + strconv.Itoa(sp.length) + "; data:" + hex.EncodeToString(sp.data))
	}

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

func WriteSystemPacket(sp SystemPacket, conn net.Conn) {
	var len = 4 + len(sp.data)
	outbuf := make([]byte, len)
	outbuf[0] = 0x7f
	outbuf[1] = byte(sp.groupid)
	outbuf[2] = byte(sp.length)
	outbuf[3] = 0
	copy(outbuf[4:], sp.data)
	_, err := conn.Write(outbuf)
	if err != nil {
		fmt.Println("Error writing to connection")
	}
	fmt.Println("Wrote System packet: " + hex.EncodeToString(outbuf))
}
