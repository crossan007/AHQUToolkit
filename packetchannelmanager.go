package main

import (
	"fmt"
	"net"
)

func ReceivePackets(conn net.Conn, systempackets chan<- SystemPacket) {

	for {
		var buf1 [1]byte

		_, err1 := conn.Read(buf1[0:])
		if err1 != nil {
			//return sp, errors.New("Error reading connection buffer, read " + strconv.Itoa(n) + " bytes read")
		}
		if buf1[0] == 0x7f {
			var buf2 [3]byte
			_, err2 := conn.Read(buf2[0:])
			if err2 != nil {
				//return sp, errors.New("Error reading packet group or data length")
			}
			var group = buf2[0]
			var len = int(buf2[1])
			buf3 := make([]byte, len)
			_, err3 := conn.Read(buf3[0:])
			if err3 != nil {
				//return sp, errors.New("Error reading system packet data")
			}

			systempackets <- SystemPacket{
				groupid: group,
				data:    buf3}
		} else if buf1[0] == 0xf7 {
			fmt.Println("Received DSP Packet")
			//return sp, errors.New("Expected header 0x07 for system packet; got: " + hex.EncodeToString(buf1[:]))
		}
	}

}

func SendPackets(conn net.Conn, systempackets <-chan SystemPacket) {
	for {
		sp := <-systempackets
		WriteSystemPacket(sp, conn)
	}
}
