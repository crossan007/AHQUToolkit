package main

type Route struct {
	Offset     int
	Id         int
	RawValue   string
	Enabled    bool
	PreOrPost  string
	FaderValue FaderValue
}
