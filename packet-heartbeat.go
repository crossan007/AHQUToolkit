package main

type HeartbeatPacket struct {
	data                []byte
	remoteControlClient RemoteControlClient
}
