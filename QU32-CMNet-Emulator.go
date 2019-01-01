package main

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"strconv"
	"strings"
	"time"
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
	var ClientUDPListeningPort = 0

	for i := 0; i < 2; i++ {
		//read two system packets from the remote
		sp, err := ReadSystemPacket(conn)
		if err != nil {
			fmt.Println("Error reading system packet: " + err.Error())
			return
		}
		if sp.groupid == 0 {
			ClientUDPListeningPort = int(binary.LittleEndian.Uint16(sp.data))
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

	// after the second time sending 03015.., wait for 10 system packets from the client;
	for i := 0; i < 10; i++ {
		sp1, err1 := ReadSystemPacket(conn)
		if err1 != nil {
			fmt.Println("Error reading system packet: " + err1.Error())
			return
		} else {
			fmt.Println(sp1)
		}
	}

	// after 10 packets received, send the channel data

	channelData1, err := ioutil.ReadFile("ChannelData1.bin")
	check(err)
	WriteSystemPacket(SystemPacket{groupid: 0x06, data: channelData1}, conn)

	channelData2, err := ioutil.ReadFile("ChannelData2.bin")
	check(err)
	WriteSystemPacket(SystemPacket{groupid: 0x16, data: channelData2}, conn)

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

	// set up the UDP connection
	var UDPConnectionString = strings.Split(conn.RemoteAddr().String(), ":")[0] + ":" + strconv.Itoa(ClientUDPListeningPort)
	fmt.Println("Setting up UDP connection to " + UDPConnectionString)
	UDPconn, err := net.Dial("udp", UDPConnectionString)
	if err != nil {
		fmt.Printf("Some error %v", err)
		return
	}
	// send the heartbeat on a regular interval with routine SendUDPHeartbeat
	go SendUDPHeartbeat(UDPconn)

	// read packets until end
	for {
		sp1, err1 := ReadSystemPacket(conn)
		if err1 != nil {
			fmt.Println("Error reading system packet: " + err1.Error())
		} else {
			fmt.Println(sp1)
		}
	}

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

func SendUDPHeartbeat(conn net.Conn) {

	ticker := time.NewTicker(100 * time.Millisecond)
	go func() {
		for t := range ticker.C {
			byteArray1, _ := hex.DecodeString("7f261200000000000000000000000000000000000000")
			_, err := conn.Write(byteArray1)

			if err != nil {
				fmt.Printf("Couldn't send response %v", err)
			}

			byteArray2, _ := hex.DecodeString("7f2759000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")
			_, err2 := conn.Write(byteArray2)
			if err2 != nil {
				fmt.Printf("Couldn't send response %v", err)
			}
			_ = t
		}
	}()

}
