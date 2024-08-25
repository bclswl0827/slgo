package main

import (
	"time"

	"github.com/bclswl0827/slgo/handlers"
)

type provider struct {
	startTime time.Time
}

func (p *provider) GetSoftware() string {
	return "slgo"
}

func (p *provider) GetStartTime() time.Time {
	return p.startTime
}

func (p *provider) GetCurrentTime() time.Time {
	return time.Now().UTC()
}

func (p *provider) GetOrganization() string {
	return "anyshake.org"
}

func (p *provider) GetStations() []handlers.SeedLinkStation {
	return []handlers.SeedLinkStation{
		{
			BeginSequence: "000000",
			EndSequence:   "FFFFFF",
			Station:       "SHAKE",
			Network:       "AS",
			Description:   "Sample station",
		},
	}
}

func (p *provider) GetStreams() []handlers.SeedLinkStream {
	return []handlers.SeedLinkStream{
		{
			BeginTime: p.GetStartTime().Format("2006-01-02 15:04:01"),
			EndTime:   "9999-12-31 23:59:59",
			SeedName:  "EHZ",
			Location:  "00",
			Type:      "D",
			Station:   "SHAKE", // Should match the station name in GetStations
		},
		{
			BeginTime: p.GetStartTime().Format("2006-01-02 15:04:01"),
			EndTime:   "9999-12-31 23:59:59",
			SeedName:  "EHE",
			Location:  "00",
			Type:      "D",
			Station:   "SHAKE", // Should match the station name in GetStations
		},
		{
			BeginTime: p.GetStartTime().Format("2006-01-02 15:04:01"),
			EndTime:   "9999-12-31 23:59:59",
			SeedName:  "EHN",
			Location:  "00",
			Type:      "D",
			Station:   "SHAKE", // Should match the station name in GetStations
		},
	}
}

func (p *provider) GetCapabilities() []handlers.SeedLinkCapability {
	return []handlers.SeedLinkCapability{
		{Name: "info:all"}, {Name: "info:gaps"}, {Name: "info:streams"},
		{Name: "dialup"}, {Name: "info:id"}, {Name: "multistation"},
		{Name: "window-extraction"}, {Name: "info:connections"},
		{Name: "info:capabilities"}, {Name: "info:stations"},
	}
}

func (p *provider) QueryHistory(startTime, endTime time.Time, channels []string) ([]handlers.SeedLinkDataPacket, error) {
	var dataPackets []handlers.SeedLinkDataPacket

	// Generate random data packets for each channel, every second
	startTimestamp, endTimestamp := startTime.UnixMilli(), endTime.UnixMilli()
	for i := startTimestamp; i < endTimestamp; i += 1000 {
		for _, channel := range channels {
			dataPacket := handlers.SeedLinkDataPacket{
				Timestamp:  i,
				SampleRate: SAMPLE_RATE,
				Channel:    channel,
				DataArr:    generateRandomArray(SAMPLE_RATE, -32768, 32768),
			}
			dataPackets = append(dataPackets, dataPacket)
		}
	}

	return dataPackets, nil
}
