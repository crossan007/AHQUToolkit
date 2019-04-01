package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
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
	log.Println("Done converting all files")
}

func hash(b []byte) uint32 {
	h := fnv.New32()
	h.Write(b)
	return h.Sum32()
}
func decodeChannel(channelBytes []byte, offset int) SavedChannel {
	var sc SavedChannel
	sc.Offset = offset
	var channelTypeb = channelBytes[0:114]
	sc.Type = int(binary.LittleEndian.Uint16(channelTypeb))
	sc.FaderValue = GetFaderValue(channelBytes[121:123])
	channelNameBytes := bytes.IndexByte(channelBytes[0x9C:0xA4], 0)
	sc.Name = string(channelBytes[0x9C : 0x9C+channelNameBytes])
	sc.GainValue = GetKnobValue(channelBytes[0x81])
	sc.Id = int(channelBytes[177])
	sc.RawValue = hex.EncodeToString(channelBytes)
	sc.EQ = ChannelEQ{
		Enabled:  hex.EncodeToString(channelBytes[0x1A:0x1B]),
		RawBytes: hex.EncodeToString(channelBytes[0x00:0x1A])}
	sc.Compression = ChannelCompression{
		Enabled: hex.EncodeToString(channelBytes[0x29:0x2A])}
	sc.Gate = ChannelGate{
		Enabled: hex.EncodeToString(channelBytes[0x34:0x35])}

	return sc
}

func convertScene(fileName string) {
	var channelSize = 0xC0
	var channelsOffset = 0x30
	var channelsCount = 60
	var Scene Scene

	c := make([]SavedChannel, channelsCount)
	Scene.Channels = c

	channelData1, _ := ioutil.ReadFile(fileName)

	Scene.Name = string(channelData1[12:23])
	Scene.Id = int(channelData1[3])

	// Take a few gesses at this being a hashing algo?
	Scene.Checksum1 = strconv.FormatUint(uint64(hash(channelData1[3:0x6518])), 16)
	Scene.Checksum1 = strconv.FormatUint(3570586489, 16)
	Scene.Checksum2 = hex.EncodeToString(channelData1[0x651C:0x6520])
	// not really working

	for i := 0; i < channelsCount; i++ {
		var offset = channelsOffset + (i * channelSize)
		Scene.Channels[i] = decodeChannel(channelData1[offset:offset+channelSize], offset)
	}

	var routesCount = 1180
	var routesOffsetBegin = 0x2E60
	var routeLength = 8
	// Routes with non-zero sends can be found with regex in the resultant JSON object  /RawValue": "[^0][0-9a-f]{15}"/
	r := make([]Route, routesCount)
	Scene.Routes = r
	for i := 0; i < routesCount; i++ {
		var r Route
		r.Id = i
		r.Offset = routesOffsetBegin + i*routeLength
		r.RawValue = hex.EncodeToString(channelData1[r.Offset : r.Offset+8])
		r.FaderValue = GetFaderValue(channelData1[r.Offset : r.Offset+3])
		r.Enabled = channelData1[r.Offset+5] == 0x01
		if channelData1[r.Offset+4] == 0x01 {
			r.PreOrPost = "Pre"
		} else {
			r.PreOrPost = "Post"
		}
		Scene.Routes[i] = r
	}

	b, err := json.MarshalIndent(Scene, "", "  ")
	if err != nil {
		fmt.Println(err)
		return
	}
	_ = ioutil.WriteFile(fileName+".json", b[:], 0644)
}
