package handlers

import (
	"encoding/xml"
	"fmt"
	"strings"
	"time"

	"github.com/bclswl0827/mseedio"
)

type INFO struct{}

// Callback of "INFO <...>" command, implements handler interface
func (i *INFO) Callback(client *SeedLinkClient, provider SeedLinkProvider, consumer SeedLinkConsumer, args ...string) error {
	err := fmt.Errorf("arg error")
	if len(args) != 1 {
		return err
	}

	var (
		action    = strings.ToUpper(args[0])
		dataBytes []byte
	)
	switch action {
	case "ID":
		dataBytes, err = i.getID(provider, FLAG_INF)
	case "STATIONS":
		dataBytes, err = i.getStations(provider)
	case "CAPABILITIES":
		dataBytes, err = i.getCapabilities(provider)
	case "CONNECTIONS":
		dataBytes, err = i.getID(provider, FLAG_ERR)
	case "STREAMS":
		dataBytes, err = i.getStreams(provider)
	default:
		dataBytes, err = i.getID(provider, FLAG_ERR)
	}
	if err != nil {
		return err
	}

	_, err = client.Write(dataBytes)
	return err
}

// Fallback of "INFO <...>" command, implements handler interface
func (i *INFO) Fallback(client *SeedLinkClient, provider SeedLinkProvider, consumer SeedLinkConsumer, args ...string) {
	client.Write([]byte(RES_ERR))
}

// getID returns response of "INFO ID" command
func (i *INFO) getID(provider SeedLinkProvider, flag int) ([]byte, error) {
	result := seedLinkInfo{
		Software:     provider.GetSoftware(),
		Started:      formatInfoTime(provider.GetStartTime()),
		Organization: provider.GetOrganization(),
	}
	xmlData, err := xml.Marshal(result)
	if err != nil {
		return []byte(RES_ERR), err
	}
	// Set XML header and return response
	xmlBody := i.setXMLHeader(xmlData)
	currentTime := provider.GetCurrentTime()
	return i.setResponse(xmlBody, flag, currentTime)
}

// getStations returns response of "INFO STATIONS" command
func (i *INFO) getStations(provider SeedLinkProvider) ([]byte, error) {
	result := seedLinkInfo{
		Software:     provider.GetSoftware(),
		Started:      formatInfoTime(provider.GetStartTime()),
		Organization: provider.GetOrganization(),
		Stations:     stationInfo(provider.GetStations(), nil),
	}
	xmlData, err := xml.Marshal(result)
	if err != nil {
		return []byte(RES_ERR), err
	}
	// Set XML header and return response
	xmlBody := i.setXMLHeader(xmlData)
	currentTime := provider.GetCurrentTime()
	return i.setResponse(xmlBody, FLAG_INF, currentTime)
}

// getCapabilities returns response of "INFO CAPABILITIES" command
func (i *INFO) getCapabilities(provider SeedLinkProvider) ([]byte, error) {
	result := seedLinkInfo{
		Software:     provider.GetSoftware(),
		Started:      formatInfoTime(provider.GetStartTime()),
		Organization: provider.GetOrganization(),
		Capabilities: provider.GetCapabilities(),
	}
	xmlData, err := xml.Marshal(result)
	if err != nil {
		return []byte(RES_ERR), err
	}
	// Set XML header and return response
	xmlBody := i.setXMLHeader(xmlData)
	currentTime := provider.GetCurrentTime()
	return i.setResponse(xmlBody, FLAG_INF, currentTime)
}

// getStreams returns response of "INFO STREAMS" command
func (i *INFO) getStreams(provider SeedLinkProvider) ([]byte, error) {
	result := seedLinkInfo{
		Software:     provider.GetSoftware(),
		Started:      formatInfoTime(provider.GetStartTime()),
		Organization: provider.GetOrganization(),
		Stations:     stationInfo(provider.GetStations(), provider.GetStreams()),
	}
	xmlData, err := xml.Marshal(result)
	if err != nil {
		return []byte(RES_ERR), err
	}
	// Set XML header and return response
	xmlBody := i.setXMLHeader(xmlData)
	currentTime := provider.GetCurrentTime()
	return i.setResponse(xmlBody, FLAG_INF, currentTime)
}

// setXMLHeader sets XML header to body and return string
func (i *INFO) setXMLHeader(body []byte) []byte {
	header := []byte(`<?xml version="1.0" encoding="utf-8"?>`)
	return append(header, body...)
}

// setResponse assembles response in MiniSeed format
func (i *INFO) setResponse(body []byte, errFlag int, startTime time.Time) ([]byte, error) {
	// Convert body to int32 array
	bodyBuffer := []int32{}
	for _, v := range body {
		bodyBuffer = append(bodyBuffer, int32(v))
	}
	// Set channel code by error flag
	channelCode := "INF"
	if errFlag == FLAG_ERR {
		channelCode = "ERR"
	}
	// Initialize MiniSeed data
	var miniseed mseedio.MiniSeedData
	miniseed.Init(mseedio.ASCII, mseedio.MSBFIRST)
	// Split data into 512 bytes each
	bodyLength := len(bodyBuffer)
	dataLength := (512 - mseedio.FIXED_SECTION_LENGTH - mseedio.BLOCKETTE100X_SECTION_LENGTH)
	blockCount := (bodyLength + dataLength - 1) / dataLength
	if blockCount == 0 {
		blockCount = 1
	}
	// "SLINFO<space>*" or "SLINFO<space><space>" is signature
	// * indicates non-final block, <space> indicates final block
	// Append each block to MiniSeed data
	var resultBuffer []byte
	for blockIndex := 0; blockIndex < blockCount; blockIndex++ {
		startIndex := blockIndex * dataLength
		endIndex := (blockIndex + 1) * dataLength
		if endIndex > bodyLength {
			endIndex = bodyLength
		}
		blockHeader := []byte("SLINFO *")
		if blockIndex == blockCount-1 {
			blockHeader[7] = ' '
		}
		err := miniseed.Append(
			bodyBuffer[startIndex:endIndex],
			&mseedio.AppendOptions{
				SequenceNumber: fmt.Sprintf("%06d", (blockIndex%999999)+1),
				ChannelCode:    channelCode,
				StartTime:      startTime,
				StationCode:    "INFO ",
				LocationCode:   "  ",
				NetworkCode:    "SL",
				SampleRate:     0,
			},
		)
		if err != nil {
			return nil, err
		}
		// Encode MiniSeed data
		res, err := miniseed.Encode(mseedio.APPEND, mseedio.MSBFIRST)
		if err != nil {
			return nil, err
		}
		// Each block should be 512 bytes
		if len(res) < 512 {
			// Fill with 0x00 if length is less than 512
			res = append(res, make([]byte, 512-len(res))...)
		}
		resultBuffer = append(resultBuffer, blockHeader...)
		resultBuffer = append(resultBuffer, res...)
	}
	return resultBuffer, nil
}

type seedLinkInfo struct {
	XMLName      xml.Name              `xml:"seedlink"`
	Software     string                `xml:"software,attr"`
	Started      string                `xml:"started,attr"`
	Organization string                `xml:"organization,attr"`
	Capabilities []SeedLinkCapability  `xml:"capability,omitempty"`
	Stations     []seedLinkStationInfo `xml:"station,omitempty"`
}

type seedLinkStationInfo struct {
	SeedLinkStation
	StreamCheck string           `xml:"stream_check,attr,omitempty"`
	Streams     []SeedLinkStream `xml:"stream,omitempty"`
}

func stationInfo(stations []SeedLinkStation, streams []SeedLinkStream) []seedLinkStationInfo {
	result := make([]seedLinkStationInfo, 0, len(stations))
	includeStreams := streams != nil
	for _, station := range stations {
		item := seedLinkStationInfo{SeedLinkStation: station}
		if includeStreams {
			item.StreamCheck = "enabled"
			for _, stream := range streams {
				if stream.Station == station.Station {
					item.Streams = append(item.Streams, stream)
				}
			}
		}
		result = append(result, item)
	}
	return result
}

func formatInfoTime(value time.Time) string {
	return value.UTC().Format("2006-01-02 15:04:05")
}
