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

func convertScene(fileName string) {
	var channelSize = 0xC0
	var channelsOffset = 0x45
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
		var sc SavedChannel
		sc.Offset = channelsOffset + (i * channelSize)
		var channelTypeb = channelData1[sc.Offset : sc.Offset+99]
		sc.Type = int(binary.LittleEndian.Uint16(channelTypeb))
		sc.FaderValue = GetFaderValue(channelData1[sc.Offset+106 : sc.Offset+108])
		channelNameBytes := bytes.IndexByte(channelData1[sc.Offset+135:sc.Offset+143], 0)
		sc.Name = string(channelData1[sc.Offset+135 : sc.Offset+135+channelNameBytes])
		sc.GainValue = GetKnobValue(channelData1[sc.Offset+108])
		sc.Id = int(channelData1[sc.Offset+162])
		sc.RawValue = hex.EncodeToString(channelData1[sc.Offset : sc.Offset+channelSize])
		sc.Compression = ChannelCompression{
			Enabled: hex.EncodeToString(channelData1[sc.Offset+20 : sc.Offset+21])}
		sc.Gate = ChannelGate{
			Enabled: hex.EncodeToString(channelData1[sc.Offset+31 : sc.Offset+32])}
		sc.EQ = ChannelEQ{
			Enabled: hex.EncodeToString(channelData1[sc.Offset+5 : sc.Offset+6])}

		Scene.Channels[i] = sc
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
