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

	channelData1, _ := ioutil.ReadFile(fileName)

	Scene.Name = string(channelData1[12:23])
	for i := 0; i < channelsCount; i++ {
		var sc SavedChannel
		sc.Offset = channelsOffset + (i * channelSize)
		var channelTypeb = channelData1[sc.Offset : sc.Offset+2]
		sc.Type = int(binary.LittleEndian.Uint16(channelTypeb))
		sc.FaderValue = int(binary.LittleEndian.Uint16(channelData1[sc.Offset+10 : sc.Offset+13]))

		channelNameBytes := bytes.IndexByte(channelData1[sc.Offset+38:sc.Offset+46], 0)
		sc.Name = string(channelData1[sc.Offset+38 : sc.Offset+38+channelNameBytes])

		sc.Id = int(channelData1[sc.Offset+65])

		Scene.Channels[i] = sc
	}

	b, err := json.MarshalIndent(Scene, "", "  ")
	if err != nil {
		fmt.Println(err)
		return
	}
	_ = ioutil.WriteFile(fileName+".json", b[:], 0644)
}
