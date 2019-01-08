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
	var UDPConnectionString = remoteControlClient.IPAddress + ":" + strconv.Itoa(remoteControlClient.UDPHeartbeatPort)
	log.Println("Setting up UDP connection to " + UDPConnectionString)
	UDPconn, err := net.Dial("udp", UDPConnectionString)
	if err != nil {
		fmt.Printf("Some error %v", err)
		return
	}
	// send the heartbeat on a regular interval with routine SendUDPHeartbeat
	ticker := time.NewTicker(500 * time.Millisecond)
	var counter = 1
	go func() {
		for t := range ticker.C {
			if counter%4 != 0 {
				// send three of this packet
				byteArray1, _ := hex.DecodeString("7f261200000000000000000000000000000000000000")
				_, err := UDPconn.Write(byteArray1)

				if err != nil {
					fmt.Printf("Couldn't send response %v", err)
				}
			} else {
				// and every fourth packet should be these two
				byteArray2, _ := hex.DecodeString("7F233003BA43C229CE374736BA43C32C0112176CFD7F0D805B475B475B475B475B473B44D13D176C0D800D80EA4E3E4D3E4D3E4D3E4D974BBA43176C0D800D80BF495B475B475B47BF495B47D13D176C0D800D803E4D974BA451A451BF49974B974BC9720D800D80D1575A575A575A57DD56DD56A451176C0D800D80DD565A575A575A575A573355E14F176C0D800D800112011201120112011201120112166C0D800D800112011201120112011201120112166C0D800D800112D13DBA43BA430112D13DD13D176CFB7F0D800112011201120112011201120112166C0D800D800112011201120112011201120112166C0D800D800112011201120112011201120112166C0D800D80EA4EEA4ECA4BCA4BCA4BCA4B77480E80FB7F0D800112011201120112011201120112176C0D800D800112011201120112011201120112166C0D800D800112011201120112011201120112166C06800D800112011201120112011201120112176C0D800D8039664C69AB68AB683966AB6801120E80FB7F0D80F9340112B13AB13AF93401120112176CFC7F0D80E14FE14FE14FE14FE14FE14F0112166C0D800D80D13DD13D5B475B47D13DD13D0112176C06800D80E853C950D157DD56A4517152CA55176C04800D803E4D974BA451A451974B974BCA55176CF67F0D800112BA43BA43BA430112BA430112166C0D800D800112BA43BA43BA430112BA430112166C0D800D80801E801E801E801E801E801E0112166C0D800D805C1B152115211521801E801E0112166C0D800D800112011201120112011201120112166C0D800D800112011201120112011201120112166C0D800D80011201120112011200920092011201000D800100011201120112011200920092011201000D8001002F67B2649A649A6495649564AB61166CFA7F0D800112011201120112011201120112166C0D800D800112011201120112011201120112176C0D800D800112011201120112011201120112176C0D800D800112011201120112011201120112176CFB7F0D80F9340112D13DD13D6D2C01120112166C00800D800112011201120112011201120112166C0D800D8001120112D13DD13DD92301120112176C0D800D80")
				_, err2 := UDPconn.Write(byteArray2)
				if err2 != nil {
					fmt.Printf("Couldn't send response %v", err)
				}
				_ = t

				byteArray3, _ := hex.DecodeString("7F240C030112B565DE6402640264AA58E863E8630D8001000112CB65236623662366C950CB6523660780010001129A6867685668566843679A685668078001000112B3673E673E6730672C6230673E670780010001123E6C1F6C296C1F6C16703E6C3E6C078001000112666C486C526C486C4570666C3E6C078001000112D862786178617861156178615A61078001000112D8625A615A615A6115615A615A61078001000112805FC66015610E628362805F1360FB7F01000112B25FC66015612C62BD62B25F1360FB7F0100011271527152C950C950C950E14FE14F07800100011271527152C950C950C950E14FE14F07800100011201120112011201120112011201120D800D80011201120112011201120112011201120D800D8001123E4D3E4D3E4D3E4D974B3E4D3E4D07800D8001123E4D3E4D3E4D3E4D974B3E4D3E4D07800D8001128A5C025B025B025B5A57025B025B0D800D8001128A5C025B025B025B5A57025B025B0D800D800112A86BCB66CB66A266BD62A86BA2660D800D800112A86BCB66CB66A266BD62A86BA2660D800D800112F56239663966876BF96E0E625965FB7F01000112F56239663966A86B166F2C625965FB7F01000112F466F466F466E265F96EF865F865078001000112E066E066E0660E661D6F0E66F86507800100C950C950C950DD564260C950C950C950C950C950C950011201127E636363C950011201120112011201120112CE37CE37B13AB73DD13DD13DD13D3B443B445B47D13DD13DBA43BA43BA43BA439A40D13DD13D01120112011201120112BF12011201120112011201120112F934F934B13AE23AB13AB13AD13D3B443B443B44D13DD13DBA43BA43BA43BA43BA43D13DD13D011201120112011201129B15BF49BF49BF49D13DD13D4D4AB24DD13DD13D01120112011201120112011201120112011201120112011201120112134E555C0112011201120112011201120112011201120112011201120112011201120112010001000112011201120112011201120112011201120112011201120112011201120112E34D925C01120112011201120112011201120112011201120112")
				_, err3 := UDPconn.Write(byteArray3)
				if err3 != nil {
					fmt.Printf("Couldn't send response %v", err)
				}
				_ = t
			}
			counter++
		}
	}()
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
