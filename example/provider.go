package main

import (
	"errors"
	"time"

	"github.com/bclswl0827/slgo/handlers"
	"github.com/samber/lo"
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
			EndTime:   p.GetCurrentTime().Format("2006-01-02 15:04:01"),
			SeedName:  "EHZ",
			Location:  "00",
			Type:      "D",
			Station:   "SHAKE", // Should match the station name in GetStations
		},
		{
			BeginTime: p.GetStartTime().Format("2006-01-02 15:04:01"),
			EndTime:   p.GetCurrentTime().Format("2006-01-02 15:04:01"),
			SeedName:  "EHE",
			Location:  "00",
			Type:      "D",
			Station:   "SHAKE", // Should match the station name in GetStations
		},
		{
			BeginTime: p.GetStartTime().Format("2006-01-02 15:04:01"),
			EndTime:   p.GetCurrentTime().Format("2006-01-02 15:04:01"),
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

func (p *provider) QueryHistory(startTime, endTime time.Time, channels []handlers.SeedLinkChannel) ([]handlers.SeedLinkDataPacket, error) {
	channels = lo.Filter(channels, func(item handlers.SeedLinkChannel, _ int) bool {
		return lo.Contains([]string{"EHN", "EHE", "EHZ"}, item.ChannelName)
	})
	if len(channels) == 0 {
		return nil, errors.New("no available channel!")
	}

	dataPackets := make([]handlers.SeedLinkDataPacket, 0)

	// Generate random data packets for each channel, every second
	startTimestamp, endTimestamp := startTime.UnixMilli(), endTime.UnixMilli()
	for i := startTimestamp; i < endTimestamp; i += 1000 {
		for _, channel := range channels {
			dataPacket := handlers.SeedLinkDataPacket{
				Timestamp:  i,
				SampleRate: SAMPLE_RATE,
				Channel:    channel.ChannelName,
				DataArr:    generateSineWave(SAMPLE_RATE),
			}
			dataPackets = append(dataPackets, dataPacket)
		}
	}

	return dataPackets, nil
}
