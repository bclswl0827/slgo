package main

import (
	"time"

	"github.com/bclswl0827/slgo/handlers"
)

var currentTime time.Time

func init() {
	currentTime = time.Now().UTC()
}

type provider struct{}

func (p *provider) GetSoftware() string {
	return "slgo"
}

func (p *provider) GetStartTime() time.Time {
	return currentTime
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
			Station:   "SHAKE",
		},
		{
			BeginTime: p.GetStartTime().Format("2006-01-02 15:04:01"),
			EndTime:   "9999-12-31 23:59:59",
			SeedName:  "EHE",
			Location:  "00",
			Type:      "D",
			Station:   "SHAKE",
		},
		{
			BeginTime: p.GetStartTime().Format("2006-01-02 15:04:01"),
			EndTime:   "9999-12-31 23:59:59",
			SeedName:  "EHN",
			Location:  "00",
			Type:      "D",
			Station:   "SHAKE",
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
	return []handlers.SeedLinkDataPacket{}, nil
}
