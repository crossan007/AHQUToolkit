package main

import (
	"encoding/binary"
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
	var buf [512]byte

	n, err := conn.Read(buf[0:])
	if err != nil {
		fmt.Println("Error reading connection buffer")
	}
	if buf[0] == 0x7f {
		sp := SystemPacket{
			groupid: int(buf[1]),
			length:  int(buf[2]),
			data:    buf[4:n]}

		var port = int(binary.LittleEndian.Uint16(sp.data))
		fmt.Println("Received System Packet.  GroupID: " + strconv.Itoa(sp.groupid) + "; length: " + strconv.Itoa(sp.length) + "; port:" + strconv.Itoa(port))

		var response = []byte{0x7f, 0x00, 0x02, 0x00, 0x00, 0xc0, 0x7f, 0x01, 0x0c, 0x00, 0x03, 0x01, 0x5f, 0x01, 0xd1, 0x11, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00} // this seems to be the mixer's response after the app sends a system packet
		_, err2 := conn.Write(response)
		if err2 != nil {
			fmt.Println("Error writing to connection")
		}
		fmt.Println("Wrote mixer response")

	}

}
