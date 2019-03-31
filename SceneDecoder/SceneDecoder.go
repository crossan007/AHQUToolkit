package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
)

func main() {

	var channelSize = 0xC0
	var channelsOffset = 0xA6
	var channelsCount = 60
	channels := make([]SavedChannel, channelsCount)
	var fileName = "SCENE020.DAT"
	// var fileName = "CH1.dat"
	channelData1, _ := ioutil.ReadFile(fileName)

	for i := 0; i < channelsCount; i++ {
		var currentChannelOffset = channelsOffset + (i * channelSize)
		log.Println("Channel: " + strconv.Itoa(i) + " offset: " + strconv.Itoa(currentChannelOffset))

		var faderValue = int(binary.LittleEndian.Uint16(channelData1[currentChannelOffset+10 : currentChannelOffset+13]))

		channelNameBytes := bytes.IndexByte(channelData1[currentChannelOffset+38:currentChannelOffset+46], 0)
		channelName := string(channelData1[currentChannelOffset+38 : currentChannelOffset+38+channelNameBytes])

		channels[i] = SavedChannel{
			FaderValue: faderValue,
			Name:       channelName}
	}
	_ = channelData1
	_ = channelsOffset
	_ = channelSize

	b, err := json.MarshalIndent(channels, "", "  ")
	if err != nil {
		fmt.Println(err)
		return
	}
	_ = ioutil.WriteFile(fileName+".json", b[:], 0644)
}
