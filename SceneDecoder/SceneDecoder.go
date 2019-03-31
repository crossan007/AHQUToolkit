package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

func main() {
	var files []string

	root := "Scenes"
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		files = append(files, path)
		return nil
	})
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		if filepath.Ext(file) == ".DAT" {
			log.Println("Converting " + file)
			convertScene(file)
		}

	}
}

func convertScene(fileName string) {
	var channelSize = 0xC0
	var channelsOffset = 0xA6
	var channelsCount = 60
	var Scene Scene

	c := make([]SavedChannel, channelsCount)
	Scene.Channels = c

	// var fileName = "CH1.dat"
	channelData1, _ := ioutil.ReadFile(fileName)

	Scene.Name = string(channelData1[12:23])
	for i := 0; i < channelsCount; i++ {
		var currentChannelOffset = channelsOffset + (i * channelSize)
		var channelTypeb = channelData1[currentChannelOffset : currentChannelOffset+2]
		var channelType = int(binary.LittleEndian.Uint16(channelTypeb))
		var faderValue = int(binary.LittleEndian.Uint16(channelData1[currentChannelOffset+10 : currentChannelOffset+13]))

		channelNameBytes := bytes.IndexByte(channelData1[currentChannelOffset+38:currentChannelOffset+46], 0)
		channelName := string(channelData1[currentChannelOffset+38 : currentChannelOffset+38+channelNameBytes])

		var channelID = int(channelData1[currentChannelOffset+65])

		Scene.Channels[i] = SavedChannel{
			Type:       channelType,
			Id:         channelID,
			FaderValue: faderValue,
			Name:       channelName}
	}
	_ = channelData1
	_ = channelsOffset
	_ = channelSize

	b, err := json.MarshalIndent(Scene, "", "  ")
	if err != nil {
		fmt.Println(err)
		return
	}
	_ = ioutil.WriteFile(fileName+".json", b[:], 0644)
}
