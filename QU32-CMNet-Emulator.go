package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"strconv"
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
	go ListenForHeartbeats(&thisMixer)
	go sendHeartbeats(&thisMixer)
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

	remoteControlClient := initializeRemoteConnection(conn, &thisMixer)
	log.Println("Console now has: " + strconv.Itoa(len(thisMixer.RemoteControlClients)) + " connected")

	// read packets until end
	for {
		select {
		case sp := <-remoteControlClient.incomingSystemPackets:
			_ = sp
			//log.Println(sp)
		case dspp := <-remoteControlClient.incomingDSPPackets:
			_ = dspp
			//log.Println(dspp)
		}
	}

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
