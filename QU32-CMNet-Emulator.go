package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

var thisMixer Mixer

func main() {

	thisMixer = Mixer{
		Name: "MixingTool",
		Version: MixerVersion{
			BoardType:     3,
			FirmwareMajor: 1,
			FirmwareMinor: 95,
			FirmwarePatch: 4561}}

	for i := 0; i < 47; i++ {
		thisMixer.DSPChannels = append(thisMixer.DSPChannels, DSPChannel{Name: "Chn" + strconv.Itoa(i), Gain: 32768, MainSendLevel: 32768})
	}

	//b, err := json.Marshal(thisMixer)
	//log.Println(string(b))

	go HandleUDPMixerDiscoveryRequests(thisMixer)
	ln, err := net.Listen("tcp", ":51326")
	if err != nil {
		// handle error
		log.Println("Error creating listener")
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
			log.Println("Error accepting connection")
		}
		log.Println("Incoming connection from: " + conn.RemoteAddr().String())
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

	remoteControlClient := InitializeRemoteConnection(conn)
	go SendUDPHeartbeat(remoteControlClient)
	// read packets until end
	for {
		select {
		case sp := <-remoteControlClient.incomingSystemPackets:
			log.Println(sp)
		case dspp := <-remoteControlClient.incomingDSPPackets:
			log.Println(dspp)
		}
	}

}
func InitializeRemoteConnection(conn net.Conn) (remoteControlClient RemoteControlClient) {

	incomingSystemPackets := make(chan SystemPacket)
	incomingDSPPackets := make(chan DSPPacket)
	outgoingSystemPackets := make(chan SystemPacket)
	go ReceivePackets(conn, incomingSystemPackets, incomingDSPPackets)
	go SendPackets(conn, outgoingSystemPackets)

	remoteControlClient.incomingSystemPackets = incomingSystemPackets
	remoteControlClient.outgoingSystemPackets = outgoingSystemPackets
	remoteControlClient.incomingDSPPackets = incomingDSPPackets
	remoteControlClient.IPAddress = strings.Split(conn.RemoteAddr().String(), ":")[0]

	sp := <-incomingSystemPackets
	remoteControlClient.UDPHeartbeatPort = int(binary.LittleEndian.Uint16(sp.data))
	log.Println("Remote UDP Port: " + strconv.Itoa(remoteControlClient.UDPHeartbeatPort))

	sp2 := <-incomingSystemPackets
	var ClientType = int(binary.LittleEndian.Uint16(sp2.data))
	remoteControlClient.clientType = ClientType
	if ClientType == 256 {
		log.Println("QU-Pad connected")
	} else if ClientType == 0 {
		log.Println("QU-You connected")
	}
	// write the mixer handshake response
	outgoingSystemPackets <- GetUDPPortSystemPacket(49152)
	outgoingSystemPackets <- SystemPacket{groupid: 0x01, data: thisMixer.Version.ToBytes()}

	for i := 0; i < 1; i++ {
		sp := <-incomingSystemPackets
		log.Println(sp)
	}

	outgoingSystemPackets <- SystemPacket{groupid: 0x01, data: thisMixer.Version.ToBytes()}
	/* after the second time sending 03015.., wait for 10 system packets from the client;
	for i := 0; i < 10; i++ {
		sp := <-incomingSystemPackets
		log.Println(sp)
	}*/

	// after 10 packets received, send the channel data

	outgoingSystemPackets <- GetDSPDataSystemPacket()

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

	return remoteControlClient
}

func SendUDPHeartbeat(remoteControlClient RemoteControlClient) {
	//set up the UDP connection
	var UDPConnectionString = remoteControlClient.IPAddress + ":" + strconv.Itoa(remoteControlClient.UDPHeartbeatPort)
	log.Println("Setting up UDP connection to " + UDPConnectionString)
	UDPconn, err := net.Dial("udp", UDPConnectionString)
	if err != nil {
		fmt.Printf("Some error %v", err)
		return
	}
	// send the heartbeat on a regular interval with routine SendUDPHeartbeat
	ticker := time.NewTicker(100 * time.Millisecond)
	go func() {
		for t := range ticker.C {
			byteArray1, _ := hex.DecodeString("7f261200000000000000000000000000000000000000")
			_, err := UDPconn.Write(byteArray1)

			if err != nil {
				fmt.Printf("Couldn't send response %v", err)
			}

			byteArray2, _ := hex.DecodeString("7f2759000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")
			_, err2 := UDPconn.Write(byteArray2)
			if err2 != nil {
				fmt.Printf("Couldn't send response %v", err)
			}
			_ = t
		}
	}()

}
