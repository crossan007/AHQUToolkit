package main

type Route struct {
	Offset     int
	Id         int
	Name       string
	RawValue   string
	Enabled    bool
	PreOrPost  string
	FaderValue FaderValue
}
