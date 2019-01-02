package main

type RemoteControlClient struct {
	clientType            byte
	incomingSystemPackets chan SystemPacket
	outgoingSystemPackets chan SystemPacket
	incomingDSPPackets    chan DSPPacket
}
