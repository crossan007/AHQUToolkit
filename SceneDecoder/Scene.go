package main

type Scene struct {
	Id        int
	Name      string
	Checksum1 string
	Checksum2 string
	Routes    []Route
	Channels  []SavedChannel
}
