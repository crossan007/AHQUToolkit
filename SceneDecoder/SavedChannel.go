package main

type SavedChannel struct {
	Offset      int
	Type        int
	Id          int
	Name        string
	GainValue   KnobValue
	FaderValue  FaderValue
	RawValue    string
	EQ          ChannelEQ
	Gate        ChannelGate
	Compression ChannelCompression
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
