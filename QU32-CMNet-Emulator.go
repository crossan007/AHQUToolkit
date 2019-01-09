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

	//outgoingSystemPackets <- SystemPacket{groupid: 0x01, data: thisMixer.Version.ToBytes()}
	/* after the second time sending 03015.., wait for 10 system packets from the client;
	for i := 0; i < 10; i++ {
		sp := <-incomingSystemPackets
		log.Println(sp)
	}*/

	// after 10 packets received, send the channel data

	outgoingSystemPackets <- GetDSPDataSystemPacket()
	log.Println("Sent DSP Data")

	if ClientType == 0 {
		outgoingSystemPackets <- CreateSystemPacketFromHexString(0x02, "0401")
		log.Println("Sent QU-You init data")
		for i := 0; i < 1; i++ {
			sp := <-incomingSystemPackets
			log.Println(sp)
		}
		outgoingSystemPackets <- CreateSystemPacketFromHexString(0x07, "0029000025000101FFFF000025000101FFFF000025000101FFFF000025000101FFFF000025000101FFFF000025000101FFFF000025000101FFFFCC2E25000101FFFF945A25000101FFFF737225000101FFFF076625000101FFFF000025000101FFFF498325000101FFFF1F7825000101FFFF000025000101FFFF000025000101FFFF000025000101FFFF000025000101FFFFAB7D25000101FFFF000025000101FFFFD27D25000101FFFF000025000101FFFF8A7D25000101FFFF420625000101FFFF907D25000101FFFF426E25000101FFFF000025000101FFFF248225000101FFFF008A25000101FFFFB27A25000101FFFFDB8125000101FFFF457825000101FFFF000025000100FFFF000025000100FFFF000025000000FFFF000025000101FFFF000025000101FFFF000025000101FFFF000025000101FFFF000025000101FFFF000025000101FFFF000025000101FFFF000025000101FFFF000025000101FFFF000025000101FFFF000025000101FFFF000025000101FFFF000025000101FFFF000025000101FFFF000025000101FFFF000025000101FFFF000025000101FFFF000025000101FFFF000025000101FFFF000025000101FFFF000025000101FFFF000025000101FFFF000025000101FFFF000025000101FFFFFFFFFFFFFFFFFFFF")
		log.Println("Sent QU-You init data2")
		/*for i := 0; i < 1; i++ {
			sp := <-incomingSystemPackets
			log.Println(sp)
		}*/

		outgoingSystemPackets <- CreateSystemPacketFromHexString(0x07, "0529907C273F0300008097650B00D0821EC60B00C0792EDD140006000100A739607B000010000C8300800001282FB4708E7A005C00940000008000800080008000800080008000800080008000800080008000800080008000800080008000800080008000800080008000800080008000000000000001009653000000000000F97E009C008000000001000000000400008000000000FFFFFFFF00000000416D616E6461000000000000987E6389607B005C00940000000000FF0000000000000000")

		log.Println("Sent QU-You init data3")
		go SendQUYOUUDPHeartbeat(remoteControlClient)

	} else {

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
		outgoingSystemPackets <- SystemPacket{groupid: 0x18, data: channelData5} //groupid 1b*/
		log.Println("Sent QU-Pad init data")
		go SendUDPHeartbeat(remoteControlClient)
	}

	return remoteControlClient
}

func SendQUYOUUDPHeartbeat(remoteControlClient RemoteControlClient) {
	//set up the UDP connection

	LocalAddr, err := net.ResolveUDPAddr("udp", "0.0.0.0:49152")
	RemoteEP := net.UDPAddr{IP: net.ParseIP(remoteControlClient.IPAddress), Port: remoteControlClient.UDPHeartbeatPort}
	log.Println("Setting up UDP connection to " + RemoteEP.String())
	UDPconn, err := net.DialUDP("udp", LocalAddr, &RemoteEP)
	if err != nil {
		fmt.Printf("Some error %v", err)
		return
	}
	// send the heartbeat on a regular interval with routine SendUDPHeartbeat
	ticker := time.NewTicker(100 * time.Millisecond)
	//ticker2 := time.NewTicker(1000 * time.Millisecond)
	var counter = 3
	go func() {
		for t := range ticker.C {
			if counter%5 != 0 {
				// send three of this packet
				byteArray1, _ := hex.DecodeString("7f261200000000000000000000000000000000000000")
				_, err := UDPconn.Write(byteArray1)

				if err != nil {
					fmt.Printf("Couldn't send response %v", err)
				}
			} else {
				// and every fourth packet should be these two
				byteArray2, _ := hex.DecodeString("7f232003ba43c229ce374736ba43c32c0112176cfd7f0d805b475b475b475b475b473b44d13d176c0d800d80ea4e3e4d3e4d3e4d3e4d974bba43176c0d800d80bf495b475b475b47bf495b47d13d176c0d800d803e4d974ba451a451bf49974b974bc9720d800d80d1575a575a575a57dd56dd56a451176c0d800d80dd565a575a575a575a573355e14f176c0d800d800112011201120112011201120112166c0d800d800112011201120112011201120112166c0d800d800112d13dba43ba430112d13dd13d176cfb7f0d800112011201120112011201120112166c0d800d800112011201120112011201120112166c0d800d800112011201120112011201120112166c0d800d80ea4eea4eca4bca4bca4bca4b77480e80fb7f0d800112011201120112011201120112176c0d800d800112011201120112011201120112166c0d800d800112011201120112011201120112166c06800d800112011201120112011201120112176c0d800d8039664c69ab68ab683966ab6801120e80fb7f0d80f9340112b13ab13af93401120112176cfc7f0d80e14fe14fe14fe14fe14fe14f0112166c0d800d80d13dd13d5b475b47d13dd13d0112176c06800d80e853c950d157dd56a4517152ca55176c04800d803e4d974ba451a451974b974bca55176cf67f0d800112ba43ba43ba430112ba430112166c0d800d800112ba43ba43ba430112ba430112166c0d800d80801e801e801e801e801e801e0112166c0d800d805c1b152115211521801e801e0112166c0d800d800112011201120112011201120112166c0d800d800112011201120112011201120112166c0d800d80011201120112011200920092011201000d800100011201120112011200920092011201000d8001002f67b2649a649a6495649564ab61166cfa7f0d800112011201120112011201120112166c0d800d800112011201120112011201120112176c0d800d800112011201120112011201120112176c0d800d800112011201120112011201120112176cfb7f0d80f9340112d13dd13d6d2c01120112166c00800d800112011201120112011201120112166c0d800d8001120112d13dd13dd92301120112176c0d800d80")
				_, err2 := UDPconn.Write(byteArray2)
				if err2 != nil {
					fmt.Printf("Couldn't send response %v", err)
				}
				_ = t

				byteArray3, _ := hex.DecodeString("7f240c030112b565de6402640264aa58e863e8630d8001000112cb65236623662366c950cb6523660780010001129a6867685668566843679a685668078001000112b3673e673e6730672c6230673e670780010001123e6c1f6c296c1f6c16703e6c3e6c078001000112666c486c526c486c4570666c3e6c078001000112d862786178617861156178615a61078001000112d8625a615a615a6115615a615a61078001000112805fc66015610e628362805f1360fb7f01000112b25fc66015612c62bd62b25f1360fb7f0100011271527152c950c950c950e14fe14f07800100011271527152c950c950c950e14fe14f07800100011201120112011201120112011201120d800d80011201120112011201120112011201120d800d8001123e4d3e4d3e4d3e4d974b3e4d3e4d07800d8001123e4d3e4d3e4d3e4d974b3e4d3e4d07800d8001128a5c025b025b025b5a57025b025b0d800d8001128a5c025b025b025b5a57025b025b0d800d800112a86bcb66cb66a266bd62a86ba2660d800d800112a86bcb66cb66a266bd62a86ba2660d800d800112f56239663966876bf96e0e625965fb7f01000112f56239663966a86b166f2c625965fb7f01000112f466f466f466e265f96ef865f865078001000112e066e066e0660e661d6f0e66f86507800100c950c950c950dd564260c950c950c950c950c950c950011201127e636363c950011201120112011201120112ce37ce37b13ab73dd13dd13dd13d3b443b445b47d13dd13dba43ba43ba43ba439a40d13dd13d01120112011201120112bf12011201120112011201120112f934f934b13ae23ab13ab13ad13d3b443b443b44d13dd13dba43ba43ba43ba43ba43d13dd13d011201120112011201129b15bf49bf49bf49d13dd13d4d4ab24dd13dd13d01120112011201120112011201120112011201120112011201120112134e555c0112011201120112011201120112011201120112011201120112011201120112010001000112011201120112011201120112011201120112011201120112011201120112e34d925c01120112011201120112011201120112011201120112")
				_, err3 := UDPconn.Write(byteArray3)
				if err3 != nil {
					fmt.Printf("Couldn't send response %v", err)
				}
				_ = t
			}
			counter++
		}
	}()

	go func() {
		c, err := net.ListenPacket("udp", ":49152")
		if err != nil {
			log.Fatal(err)
		}
		defer c.Close()
		for {
			WaitQUYouHeartBeatResponse(c)
			// send three of this packet
			byteArray1, _ := hex.DecodeString("7f250000")
			_, err := UDPconn.Write(byteArray1)

			if err != nil {
				fmt.Printf("Couldn't send response %v", err)
			}

		}
	}()
}

func WaitQUYouHeartBeatResponse(c net.PacketConn) {

	b := make([]byte, 4)
	_, peer, err := c.ReadFrom(b)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Heartbeat Echo from " + peer.String())

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
