package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"strings"
)

type Mixer struct {
	Name                     string
	Version                  MixerVersion
	DSPChannels              []DSPChannel
	RemoteControlClients     []RemoteControlClient
	outgoingHeartbeatPackets chan HeartbeatPacket
}

type MixerVersion struct {
	// Board Type can be 01 for QU-16, 02 for QU-24, or 03 for QU-32
	BoardType     uint16
	FirmwareMajor uint16
	FirmwareMinor uint16
	FirmwarePatch uint16
}

func (mv MixerVersion) ToBytes() []byte {
	//mixer version [(1)board - 0x01=16; 0x02=24; 0x03=32][(1)major 0x01 ][(2)minor 0x5f00 ][(2)patch 0xd111]
	outbuf := make([]byte, 12)
	response := make([]byte, 2)
	binary.LittleEndian.PutUint16(response, mv.BoardType)
	outbuf[0] = response[0]
	binary.LittleEndian.PutUint16(response, mv.FirmwareMajor)
	outbuf[1] = response[0]
	binary.LittleEndian.PutUint16(response, mv.FirmwareMinor)
	outbuf[2] = response[0]
	binary.LittleEndian.PutUint16(response, mv.FirmwarePatch)
	copy(outbuf[4:], response)
	return outbuf
}

func HandleUDPMixerDiscoveryRequests(mixer Mixer) {
	c, err := net.ListenPacket("udp", ":51320")
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()
	for {

		b := make([]byte, 7)
		_, peer, err := c.ReadFrom(b)
		if err != nil {
			log.Fatal(err)
		}
		//log.Println("Discovery request from: " + peer.String())
		var nameBytes = []byte(mixer.Name)
		if _, err := c.WriteTo(nameBytes, peer); err != nil {
			log.Fatal(err)
		}
	}
}

func ListenForHeartbeats(thisMixer *Mixer) {
	c, err := net.ListenPacket("udp", ":49152")
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()
	for {
		b := make([]byte, 4)
		_, peer, err := c.ReadFrom(b)
		if err != nil {
			log.Fatal(err)
		}
		var ipaddress = strings.Split(peer.String(), ":")[0]
		for i := range thisMixer.RemoteControlClients {
			if thisMixer.RemoteControlClients[i].IPAddress == ipaddress {
				thisMixer.RemoteControlClients[i].incomingHeartbeatPackets <- HeartbeatPacket{data: b}
			}
		}

		// send three of this packet
		/*byteArray1, _ := hex.DecodeString("7f250000")
		_, err := UDPconn.Write(byteArray1)

		if err != nil {
			fmt.Printf("Couldn't send response %v", err)
		}*/

	}
}

func sendHeartbeats(thisMixer *Mixer) {
	//set up the UDP connection
	LocalAddr, _ := net.ResolveUDPAddr("udp", "0.0.0.0:49152")
	outgoingHeartbeatPackets := make(chan HeartbeatPacket)
	thisMixer.outgoingHeartbeatPackets = outgoingHeartbeatPackets

	for {
		hbp := <-thisMixer.outgoingHeartbeatPackets
		RemoteEP := net.UDPAddr{IP: net.ParseIP(hbp.remoteControlClient.IPAddress), Port: hbp.remoteControlClient.UDPHeartbeatPort}
		//log.Println("Setting up UDP connection to " + RemoteEP.String())
		UDPconn, err := net.DialUDP("udp", LocalAddr, &RemoteEP)
		if err != nil {
			log.Println("Some error %v", err)
			return
		}
		_, err1 := UDPconn.Write(hbp.data)

		if err1 != nil {
			fmt.Println("Couldn't send response %v", err)
		}
		UDPconn.Close()
	}
}
