package main

type SavedChannel struct {
	Offset     int
	Type       int
	Id         int
	Name       string
	FaderValue FaderValue
	RawValue   string
}
