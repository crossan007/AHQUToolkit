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

	sp, err := ReadSystemPacket(conn)
	if err != nil {
		fmt.Println("Error reading system packet")
		return
	}
	var port = int(binary.LittleEndian.Uint16(sp.data))
	fmt.Println("Received System Packet.  GroupID: " + strconv.Itoa(sp.groupid) + "; length: " + strconv.Itoa(sp.length) + "; port:" + strconv.Itoa(port))

	var response = []byte{0x7f, 0x00, 0x02, 0x00, 0x00, 0xc0, 0x7f, 0x01, 0x0c, 0x00, 0x03, 0x01, 0x5f, 0x01, 0xd1, 0x11, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00} // this seems to be the mixer's response after the app sends a system packet
	_, err2 := conn.Write(response)
	if err2 != nil {
		fmt.Println("Error writing to connection")
	}
	fmt.Println("Wrote mixer response")

}

func ReadSystemPacket(conn net.Conn) (sp SystemPacket, err error) {
	var buf1 [2]byte

	_, err1 := conn.Read(buf1[0:])
	if err1 != nil {
		fmt.Println("Error reading connection buffer")
	}
	if buf1[0] != 0x7f {
		return sp, errors.New("Expected header 0x07 for system packet; got: " + hex.EncodeToString(buf1[:]))
	}

	var buf2 [2]byte
	_, err2 := conn.Read(buf2[0:])
	if err2 != nil {
		fmt.Println("Error reading connection buffer")
	}
	var len = int(buf2[0])
	fmt.Println("length: " + strconv.Itoa(len))

	buf3 := make([]byte, len)
	_, err3 := conn.Read(buf3[0:])
	if err3 != nil {
		fmt.Println("Error reading system packet data")
	}

	return SystemPacket{
		length: len,
		data:   buf3}, nil
}
