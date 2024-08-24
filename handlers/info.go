package handlers

import (
	"fmt"
	"time"

	"github.com/bclswl0827/mseedio"
	"github.com/clbanning/anyxml"
)

type INFO struct{}

// Callback of "INFO <...>" command, implements handler interface
func (i *INFO) Callback(client *SeedLinkClient, provider SeedLinkProvider, consumer SeedLinkConsumer, args ...string) error {
	err := fmt.Errorf("arg error")
	if len(args) < 1 {
		return err
	}

	var (
		action    = args[0]
		dataBytes []byte
	)
	switch action {
	case "ID":
		dataBytes, err = i.getID(provider, FLAG_INF)
	case "STATIONS":
		dataBytes, err = i.getStations(provider)
	case "CAPABILITIES", "CONNECTIONS":
		dataBytes, err = i.getCapabilities(provider)
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
	result := map[string]string{
		"-software":     provider.GetSoftware(),
		"-started":      provider.GetStartTime().Format("2006-01-02 15:04:01"),
		"-organization": provider.GetOrganization(),
	}
	xmlData, err := anyxml.Xml(result, "seedlink")
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
	result := map[string]any{
		"-software":     provider.GetSoftware(),
		"-started":      provider.GetStartTime().Format("2006-01-02 15:04:01"),
		"-organization": provider.GetOrganization(),
		"station":       provider.GetStations(),
	}
	xmlData, err := anyxml.Xml(result, "seedlink")
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
	result := map[string]any{
		"-software":     provider.GetSoftware(),
		"-started":      provider.GetStartTime().Format("2006-01-02 15:04:01"),
		"-organization": provider.GetOrganization(),
		"capability":    provider.GetCapabilities(),
	}
	xmlData, err := anyxml.Xml(result, "seedlink")
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
	type respModel struct {
		SeedLinkStation
		Streams     []SeedLinkStream `xml:"stream"`
		StreamCheck string           `xml:"stream_check,attr"`
	}
	result := map[any]any{
		"-software":     provider.GetSoftware(),
		"-started":      provider.GetStartTime().Format("2006-01-02 15:04:01"),
		"-organization": provider.GetOrganization(),
	}
	var resp []respModel
	for _, v := range provider.GetStations() {
		// Match stream by station name
		var availableStreams []SeedLinkStream
		for _, s := range provider.GetStreams() {
			if s.Station == v.Station {
				availableStreams = append(availableStreams, s)
			}
		}
		resp = append(resp, respModel{
			SeedLinkStation: v,
			Streams:         availableStreams,
			StreamCheck:     "enabled",
		})
	}
	result["station"] = resp
	xmlData, err := anyxml.Xml(result, "seedlink")
	if err != nil {
		fmt.Println(err)
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
	fullLength := bodyLength + mseedio.FIXED_SECTION_LENGTH + mseedio.BLOCKETTE100X_SECTION_LENGTH
	blockCount := fullLength / 512
	// "SLINFO<space>*" or "SLINFO<space><space>" is signature
	// * indicates non-final block, <space> indicates final block
	blockHeader := []byte{'S', 'L', 'I', 'N', 'F', 'O', ' ', '*'}
	// Append each block to MiniSeed data
	var resultBuffer []byte
	for i := 0; i <= blockCount; i++ {
		startIndex := i * dataLength
		endIndex := (i + 1) * dataLength
		if i == blockCount {
			// Set final block flag
			blockHeader[7] = ' '
			endIndex = bodyLength
		}
		err := miniseed.Append(
			bodyBuffer[startIndex:endIndex],
			&mseedio.AppendOptions{
				SequenceNumber: fmt.Sprintf("%06d", i+1),
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
		resultBuffer = append(resultBuffer, append(blockHeader, res...)...)
	}
	return resultBuffer, nil
}
