package main

type RemoteControlClient struct {
	clientType            int // either 256 for QU-Pad or 0 for QU-You
	incomingSystemPackets chan SystemPacket
	outgoingSystemPackets chan SystemPacket
	incomingDSPPackets    chan DSPPacket
	UDPHeartbeatPort      int
	IPAddress             string
}
