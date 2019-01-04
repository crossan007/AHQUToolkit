package main

import (
	"encoding/hex"
	"log"
	"net"
)

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
		log.Println("Bytes: " + hex.EncodeToString(b))
		var nameBytes = []byte(mixer.name)
		if _, err := c.WriteTo(nameBytes, peer); err != nil {
			log.Fatal(err)
		}
	}
}
