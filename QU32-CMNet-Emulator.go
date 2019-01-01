package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net"
	"strconv"
	"time"
)

var thisMixerVersion MixerVersion

func main() {

	thisMixerVersion = MixerVersion{
		BoardType:     3,
		FirmwareMajor: 1,
		FirmwareMinor: 95,
		FirmwarePatch: 4561}

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
		* send a system packet with the mxer type and firmware version

	*/

	incomingSystemPackets := make(chan SystemPacket)
	outgoingSystemPackets := make(chan SystemPacket)
	go ReceivePackets(conn, incomingSystemPackets)
	go SendPackets(conn, outgoingSystemPackets)
	InitializeRemoteConnection(incomingSystemPackets, outgoingSystemPackets)
	// read packets until end
	for {
		sp := <-incomingSystemPackets
		fmt.Println(sp)
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
func InitializeRemoteConnection(incomingSystemPackets <-chan SystemPacket, outgoingSystemPackets chan<- SystemPacket) {
	var ClientUDPListeningPort = 0

	for i := 0; i < 2; i++ {
		//read two system packets from the remote
		sp := <-incomingSystemPackets
		if sp.groupid == 0 {
			ClientUDPListeningPort = int(binary.LittleEndian.Uint16(sp.data))
		}

		fmt.Println(sp)
	}
	fmt.Println(strconv.Itoa(ClientUDPListeningPort))
	// write the mixer handshake response
	outgoingSystemPackets <- GetUDPPortSystemPacket(49152)
	outgoingSystemPackets <- SystemPacket{groupid: 0x01, data: thisMixerVersion.ToBytes()}

	for i := 0; i < 1; i++ {
		sp := <-incomingSystemPackets
		fmt.Println(sp)
	}

	outgoingSystemPackets <- SystemPacket{groupid: 0x01, data: thisMixerVersion.ToBytes()}
	// after the second time sending 03015.., wait for 10 system packets from the client;
	for i := 0; i < 10; i++ {
		sp := <-incomingSystemPackets
		fmt.Println(sp)
	}

	// after 10 packets received, send the channel data

	channelData1, err := ioutil.ReadFile("ChannelData1.bin")
	check(err)
	outgoingSystemPackets <- SystemPacket{groupid: 0x06, data: channelData1}

	channelData2, err := ioutil.ReadFile("ChannelData2.bin")
	check(err)
	outgoingSystemPackets <- SystemPacket{groupid: 0x16, data: channelData2}

	outgoingSystemPackets <- CreateSystemPacketFromHexString(0x0b, "0000FF00")

	outgoingSystemPackets <- CreateSystemPacketFromHexString(0x0a, "00000000")

	outgoingSystemPackets <- CreateSystemPacketFromHexString(0x22, "0100")
	outgoingSystemPackets <- CreateSystemPacketFromHexString(0x21, "0100")
	outgoingSystemPackets <- CreateSystemPacketFromHexString(0x20, "0100")

	channelData3, err := ioutil.ReadFile("ChannelData3.bin")
	check(err)
	outgoingSystemPackets <- SystemPacket{groupid: 0x1a, data: channelData3} //groupid 1a

	channelData4, err := ioutil.ReadFile("ChannelData4.bin")
	check(err)
	outgoingSystemPackets <- SystemPacket{groupid: 0x1b, data: channelData4} //groupid 1b

	channelData5, err := ioutil.ReadFile("ChannelData5.bin")
	check(err)
	outgoingSystemPackets <- SystemPacket{groupid: 0x18, data: channelData5} //groupid 1b

	/* set up the UDP connection
	var UDPConnectionString = strings.Split(conn.RemoteAddr().String(), ":")[0] + ":" + strconv.Itoa(ClientUDPListeningPort)
	fmt.Println("Setting up UDP connection to " + UDPConnectionString)
	UDPconn, err := net.Dial("udp", UDPConnectionString)
	if err != nil {
		fmt.Printf("Some error %v", err)
		return
	}
	// send the heartbeat on a regular interval with routine SendUDPHeartbeat
	go SendUDPHeartbeat(UDPconn)
	*/
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
