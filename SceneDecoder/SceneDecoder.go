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
func decodeChannel(channelBytes []byte) SavedChannel {
	var sc SavedChannel
	var channelTypeb = channelBytes[0:114]
	sc.Type = int(binary.LittleEndian.Uint16(channelTypeb))
	sc.FaderValue = GetFaderValue(channelBytes[121:123])
	channelNameBytes := bytes.IndexByte(channelBytes[0x9C:0xA4], 0)
	sc.DisplayName = string(channelBytes[0x9C : 0x9C+channelNameBytes])
	sc.GainValue = GetKnobValue(channelBytes[0x81])
	sc.Id = int(channelBytes[0xB7])
	sc.RawValue = hex.EncodeToString(channelBytes)
	sc.EQ = ChannelEQ{
		Enabled:  hex.EncodeToString(channelBytes[0x1A:0x1B]),
		RawBytes: hex.EncodeToString(channelBytes[0x00:0x1A])}
	sc.Compression = ChannelCompression{
		Enabled: hex.EncodeToString(channelBytes[0x29:0x2A])}
	sc.Gate = ChannelGate{
		Enabled: hex.EncodeToString(channelBytes[0x34:0x35])}
	sc.SendToMainFader = GetFaderValue(channelBytes[0x7e:0x80])

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
		var sc = decodeChannel(channelData1[offset : offset+channelSize])
		sc.Offset = offset
		sc.Index = i
		if i < 32 {
			sc.PhysicalName = "Channel " + strconv.Itoa(i+1)
		}
		Scene.Channels[i] = sc
	}
	Scene.Channels[32].PhysicalName = "ST1LR"
	Scene.Channels[33].PhysicalName = "ST2LR"
	Scene.Channels[34].PhysicalName = "ST3LR"

	Scene.Channels[35].PhysicalName = "Fx1RetLR"
	Scene.Channels[36].PhysicalName = "Fx2RetLR"
	Scene.Channels[37].PhysicalName = "Fx3RetLR"
	Scene.Channels[38].PhysicalName = "Fx4RetLR"

	//BEGIN GUESS
	Scene.Channels[47].PhysicalName = "GroupMix1-2"
	Scene.Channels[48].PhysicalName = "GroupMix3-4"
	Scene.Channels[49].PhysicalName = "GroupMix5-6"
	Scene.Channels[50].PhysicalName = "GroupMix7-8"
	//END GUESS

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
	Scene.Routes[0].Name = "Channel 1 to Mix 1"
	Scene.Routes[1].Name = "Channel 1 to Mix 2"
	Scene.Routes[2].Name = "Channel 1 to Mix 3"
	Scene.Routes[3].Name = "Channel 1 to Mix 4"
	Scene.Routes[4].Name = "Channel 1 to Mix 5-6"
	Scene.Routes[5].Name = "Channel 1 to Mix 7-8"
	Scene.Routes[6].Name = "Channel 1 to Mix 9-10"
	Scene.Routes[8].Name = "Channel 1 to GroupMix 1-2"
	Scene.Routes[9].Name = "Channel 1 to GroupMix 3-4"
	Scene.Routes[10].Name = "Channel 1 to GroupMix 5-6"
	Scene.Routes[11].Name = "Channel 1 to GroupMix 7-8"
	// BEGIN GUESS - 75% sure
	Scene.Routes[140].Name = "Channel 8 to Mix 1"
	Scene.Routes[300].Name = "Channel 16 to Mix 1"
	Scene.Routes[460].Name = "Channel 24 to Mix 1"
	Scene.Routes[620].Name = "Channel 32 to Mix 1"
	// END GUESS

	b, err := json.MarshalIndent(Scene, "", "  ")
	if err != nil {
		fmt.Println(err)
		return
	}
	_ = ioutil.WriteFile(fileName+".json", b[:], 0644)
}
