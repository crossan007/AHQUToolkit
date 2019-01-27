package main

import (
	"encoding/binary"
	"encoding/hex"
	"io/ioutil"
	"log"
	"math"
	"net"
	"strconv"
	"strings"
	"time"
)

type RemoteControlClient struct {
	clientType               int // either 256 for QU-Pad or 0 for QU-You
	TCPConnection            net.Conn
	incomingSystemPackets    chan SystemPacket
	outgoingSystemPackets    chan SystemPacket
	incomingDSPPackets       chan DSPPacket
	incomingHeartbeatPackets chan HeartbeatPacket
	UDPHeartbeatPort         int
	IPAddress                string
}

type RemoteControlVUMeter struct {
	value []byte
}

func initializeRemoteConnection(conn net.Conn, thisMixer *Mixer) (remoteControlClient RemoteControlClient) {

	remoteControlClient.TCPConnection = conn
	remoteControlClient.IPAddress = strings.Split(conn.RemoteAddr().String(), ":")[0]
	setupPacketChannels(&remoteControlClient)

	log.Println("Waiting for incoming system packet")
	sp := <-remoteControlClient.incomingSystemPackets
	remoteControlClient.UDPHeartbeatPort = int(binary.LittleEndian.Uint16(sp.data))
	log.Println("Remote UDP Port: " + strconv.Itoa(remoteControlClient.UDPHeartbeatPort))

	sp2 := <-remoteControlClient.incomingSystemPackets
	var ClientType = int(binary.LittleEndian.Uint16(sp2.data))
	remoteControlClient.clientType = ClientType
	if ClientType == 256 {
		log.Println("QU-Pad connected")
	} else if ClientType == 0 {
		log.Println("QU-You connected")
	}
	// write the mixer handshake response
	remoteControlClient.outgoingSystemPackets <- GetUDPPortSystemPacket(49152)
	remoteControlClient.outgoingSystemPackets <- SystemPacket{groupid: 0x01, data: thisMixer.Version.ToBytes()}

	for i := 0; i < 1; i++ {
		sp := <-remoteControlClient.incomingSystemPackets
		log.Println(sp)
	}

	remoteControlClient.outgoingSystemPackets <- GetDSPDataSystemPacket()
	log.Println("Sent DSP Data")

	if ClientType == 0 {
		remoteControlClient.outgoingSystemPackets <- CreateSystemPacketFromHexString(0x02, "0401")
		log.Println("Sent QU-You init data")
		for i := 0; i < 1; i++ {
			sp := <-remoteControlClient.incomingSystemPackets
			log.Println(sp)
		}
		// app doesn't actually seem to care about receiving these
		//outgoingSystemPackets <- CreateSystemPacketFromHexString(0x07, "0029000025000101FFFF000025000101FFFF000025000101FFFF000025000101FFFF000025000101FFFF000025000101FFFF000025000101FFFFCC2E25000101FFFF945A25000101FFFF737225000101FFFF076625000101FFFF000025000101FFFF498325000101FFFF1F7825000101FFFF000025000101FFFF000025000101FFFF000025000101FFFF000025000101FFFFAB7D25000101FFFF000025000101FFFFD27D25000101FFFF000025000101FFFF8A7D25000101FFFF420625000101FFFF907D25000101FFFF426E25000101FFFF000025000101FFFF248225000101FFFF008A25000101FFFFB27A25000101FFFFDB8125000101FFFF457825000101FFFF000025000100FFFF000025000100FFFF000025000000FFFF000025000101FFFF000025000101FFFF000025000101FFFF000025000101FFFF000025000101FFFF000025000101FFFF000025000101FFFF000025000101FFFF000025000101FFFF000025000101FFFF000025000101FFFF000025000101FFFF000025000101FFFF000025000101FFFF000025000101FFFF000025000101FFFF000025000101FFFF000025000101FFFF000025000101FFFF000025000101FFFF000025000101FFFF000025000101FFFF000025000101FFFF000025000101FFFFFFFFFFFFFFFFFFFF")
		log.Println("Sent QU-You init data2")
		/*for i := 0; i < 1; i++ {
			sp := <-incomingSystemPackets
			log.Println(sp)
		}*/
		// app doesn't actually seem to care about receiving these
		//outgoingSystemPackets <- CreateSystemPacketFromHexString(0x07, "0529907C273F0300008097650B00D0821EC60B00C0792EDD140006000100A739607B000010000C8300800001282FB4708E7A005C00940000008000800080008000800080008000800080008000800080008000800080008000800080008000800080008000800080008000800080008000000000000001009653000000000000F97E009C008000000001000000000400008000000000FFFFFFFF00000000416D616E6461000000000000987E6389607B005C00940000000000FF0000000000000000")

		log.Println("Sent QU-You init data3")
		go SendQUYOUUDPHeartbeat(remoteControlClient, thisMixer)

	} else {
		remoteControlClient.outgoingSystemPackets <- SystemPacket{groupid: 0x01, data: thisMixer.Version.ToBytes()}
		// after the second time sending 03015.., wait for 10 system packets from the client;
		for i := 0; i < 10; i++ {
			sp := <-remoteControlClient.incomingSystemPackets
			log.Println(sp)
		}

		// after 10 packets received, send the channel data

		channelData2, err := ioutil.ReadFile("ChannelData2.bin")
		check(err)
		remoteControlClient.outgoingSystemPackets <- SystemPacket{groupid: 0x16, data: channelData2}

		remoteControlClient.outgoingSystemPackets <- CreateSystemPacketFromHexString(0x0b, "0000FF00")

		remoteControlClient.outgoingSystemPackets <- CreateSystemPacketFromHexString(0x0a, "00000000")

		remoteControlClient.outgoingSystemPackets <- CreateSystemPacketFromHexString(0x22, "0100")
		remoteControlClient.outgoingSystemPackets <- CreateSystemPacketFromHexString(0x21, "0100")
		remoteControlClient.outgoingSystemPackets <- CreateSystemPacketFromHexString(0x20, "0100")

		channelData3, err := ioutil.ReadFile("ChannelData3.bin")
		check(err)
		remoteControlClient.outgoingSystemPackets <- SystemPacket{groupid: 0x1a, data: channelData3} //groupid 1a

		channelData4, err := ioutil.ReadFile("ChannelData4.bin")
		check(err)
		remoteControlClient.outgoingSystemPackets <- SystemPacket{groupid: 0x1b, data: channelData4} //groupid 1b

		channelData5, err := ioutil.ReadFile("ChannelData5.bin")
		check(err)
		remoteControlClient.outgoingSystemPackets <- SystemPacket{groupid: 0x18, data: channelData5} //groupid 1b*/
		log.Println("Sent QU-Pad init data")
		//go SendUDPHeartbeat(remoteControlClient)

	}
	thisMixer.RemoteControlClients = append(thisMixer.RemoteControlClients, remoteControlClient)
	return remoteControlClient
}

func setupPacketChannels(remoteControlClient *RemoteControlClient) {

	log.Println("Allocating channels for " + remoteControlClient.IPAddress)
	incomingSystemPackets := make(chan SystemPacket)
	incomingDSPPackets := make(chan DSPPacket)
	outgoingSystemPackets := make(chan SystemPacket)
	incomingHeartbeatPackets := make(chan HeartbeatPacket)

	remoteControlClient.incomingSystemPackets = incomingSystemPackets
	remoteControlClient.outgoingSystemPackets = outgoingSystemPackets
	remoteControlClient.incomingDSPPackets = incomingDSPPackets
	remoteControlClient.incomingHeartbeatPackets = incomingHeartbeatPackets
	log.Println("Allocated channels for " + remoteControlClient.IPAddress)

	log.Println("Configuring TCP channels " + remoteControlClient.IPAddress)
	go receiveTCPPackets(remoteControlClient)
	go sendTCPPackets(remoteControlClient)

	log.Println("Configuring UDP channels " + remoteControlClient.IPAddress)
	go ReceiveUDPPackets(remoteControlClient)
	go SendUDPPackets(remoteControlClient)

	log.Println("Configured all protocol channels " + remoteControlClient.IPAddress)
}

func receiveTCPPackets(remoteControlClient *RemoteControlClient) {

	for {

		var buf1 [1]byte
		n, err1 := remoteControlClient.TCPConnection.Read(buf1[0:])
		if err1 != nil {
			log.Println("Error reading connection buffer, read " + strconv.Itoa(n) + " bytes read")
		}
		if buf1[0] == 0x7f {
			var buf2 [3]byte
			_, err2 := remoteControlClient.TCPConnection.Read(buf2[0:])
			if err2 != nil {
				log.Println("Error reading packet group or data length")
			}
			var group = buf2[0]
			var len = int(buf2[1])
			buf3 := make([]byte, len)
			_, err3 := remoteControlClient.TCPConnection.Read(buf3[0:])
			if err3 != nil {
				log.Println("Error reading system packet data")
			}
			remoteControlClient.incomingSystemPackets <- SystemPacket{
				groupid: group,
				data:    buf3}
		} else if buf1[0] == 0xf7 {

			var DSPBytes [8]byte
			_, err2 := remoteControlClient.TCPConnection.Read(DSPBytes[0:])
			if err2 != nil {
				log.Println("Error reading packet group or data length")
			}
			log.Println("waiting for tcp input")
			remoteControlClient.incomingDSPPackets <- DSPPacket{
				ControlID:   DSPBytes[0],
				TargetGroup: DSPTargetGroup(DSPBytes[1]),
				ValueID:     DSPBytes[2],
				ClientID:    DSPBytes[3],
				Parameter1:  DSPBytes[4],
				Parameter2:  DSPBytes[5],
				Value:       DSPBytes[6],
			}
			log.Println("Expected header 0x07 for system packet; got: " + hex.EncodeToString(buf1[:]))
		}

	}

}

func sendTCPPackets(remoteControlClient *RemoteControlClient) {
	for {
		sp := <-remoteControlClient.outgoingSystemPackets
		WriteSystemPacket(sp, remoteControlClient.TCPConnection)
	}
}

func ReceiveUDPPackets(remoteControlClient *RemoteControlClient) {
	for {
		hbp := <-remoteControlClient.incomingHeartbeatPackets
		log.Println("Heartbeat packet received: " + hex.EncodeToString(hbp.data))
	}
}

func SendUDPPackets(remoteControlClient *RemoteControlClient) {

}

func SendQUYOUUDPHeartbeat(remoteControlClient RemoteControlClient, thisMixer *Mixer) {
	// send the heartbeat on a regular interval with routine SendUDPHeartbeat
	ticker := time.NewTicker(10 * time.Millisecond)
	//ticker2 := time.NewTicker(1000 * time.Millisecond)
	var counter = 3
	go func() {
		for t := range ticker.C {
			_ = t
			if counter%5 != 0 {
				// send three of this packet
				byteArray1, _ := hex.DecodeString("7f261200000000000000000000000000000000000000")
				thisMixer.outgoingHeartbeatPackets <- HeartbeatPacket{data: byteArray1, remoteControlClient: remoteControlClient}
			} else {
				VUBytes := make([]byte, 0x324)
				VUs := make([]RemoteControlVUMeter, 40)
				var ba2string = "7f232003"
				for i := 0; i < 40; i++ {
					//randbytes := make([]byte, 20)
					//rand.Read(randbytes)

					Volume := (.375 * 65535) + (.200*65535)*math.Abs(math.Sin(math.Pi*float64(counter)/1000))
					VolumeResponse := make([]byte, 4)
					binary.LittleEndian.PutUint16(VolumeResponse, uint16(Volume))
					hexBytes, _ := hex.DecodeString("0000000000000000000000000000000000000000")
					copy(hexBytes[4:], VolumeResponse)
					VUs[i].value = hexBytes
				}

				ba2string = "7f232003"
				// and every fourth packet should be these two
				byteArray2, _ := hex.DecodeString(ba2string)
				copy(VUBytes[:], byteArray2)
				for i, VU := range VUs {
					copy(VUBytes[4+(20*i):], VU.value)
				}

				thisMixer.outgoingHeartbeatPackets <- HeartbeatPacket{data: VUBytes, remoteControlClient: remoteControlClient}

				byteArray3, _ := hex.DecodeString("7f240c030112b565de6402640264aa58e863e8630d8001000112cb65236623662366c950cb6523660780010001129a6867685668566843679a685668078001000112b3673e673e6730672c6230673e670780010001123e6c1f6c296c1f6c16703e6c3e6c078001000112666c486c526c486c4570666c3e6c078001000112d862786178617861156178615a61078001000112d8625a615a615a6115615a615a61078001000112805fc66015610e628362805f1360fb7f01000112b25fc66015612c62bd62b25f1360fb7f0100011271527152c950c950c950e14fe14f07800100011271527152c950c950c950e14fe14f07800100011201120112011201120112011201120d800d80011201120112011201120112011201120d800d8001123e4d3e4d3e4d3e4d974b3e4d3e4d07800d8001123e4d3e4d3e4d3e4d974b3e4d3e4d07800d8001128a5c025b025b025b5a57025b025b0d800d8001128a5c025b025b025b5a57025b025b0d800d800112a86bcb66cb66a266bd62a86ba2660d800d800112a86bcb66cb66a266bd62a86ba2660d800d800112f56239663966876bf96e0e625965fb7f01000112f56239663966a86b166f2c625965fb7f01000112f466f466f466e265f96ef865f865078001000112e066e066e0660e661d6f0e66f86507800100c950c950c950dd564260c950c950c950c950c950c950011201127e636363c950011201120112011201120112ce37ce37b13ab73dd13dd13dd13d3b443b445b47d13dd13dba43ba43ba43ba439a40d13dd13d01120112011201120112bf12011201120112011201120112f934f934b13ae23ab13ab13ad13d3b443b443b44d13dd13dba43ba43ba43ba43ba43d13dd13d011201120112011201129b15bf49bf49bf49d13dd13d4d4ab24dd13dd13d01120112011201120112011201120112011201120112011201120112134e555c0112011201120112011201120112011201120112011201120112011201120112010001000112011201120112011201120112011201120112011201120112011201120112e34d925c01120112011201120112011201120112011201120112")
				thisMixer.outgoingHeartbeatPackets <- HeartbeatPacket{data: byteArray3, remoteControlClient: remoteControlClient}
			}
			counter++

		}
	}()

}
