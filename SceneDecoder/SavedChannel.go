package main

type SavedChannel struct {
	Index           int
	Offset          int
	Id              int
	Type            int
	PhysicalName    string
	DisplayName     string
	GainValue       KnobValue
	FaderValue      FaderValue
	RawValue        string
	EQ              ChannelEQ
	Gate            ChannelGate
	Compression     ChannelCompression
	SendToMainFader FaderValue
}

type ChannelEQ struct {
	Enabled  string
	RawBytes string
}

type ChannelGate struct {
	Enabled  string
	RawBytes string
}

type ChannelCompression struct {
	Enabled  string
	RawBytes string
}
