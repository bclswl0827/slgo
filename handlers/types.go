package handlers

import (
	"encoding/xml"
	"net"
	"sync"
	"time"
)

// The Seedlink protocol specifies that the maximum length of a data packet is 512 bytes.
// which can accommodate approximately 100 samples (int32, 4 bytes each).
// Samples more than 100 will be split into chunks of 100 samples per packet.
const CHUNK_SIZE = 100

// SeedLink handshake constant flags
const RELEASE = "SeedLink v3.1 AnyShake Edition (Basic implementation in Go, repository: https://github.com/bclswl0827/slgo) :: SLPROTO:3.1 CAP EXTREPLY NSWILDCARD BATCH WS:13"

// SeedLink error flags
const (
	FLAG_INF = iota
	FLAG_ERR
)

// SeedLink response data
const (
	RES_OK  = "OK\r\n"
	RES_ERR = "ERROR\r\n"
)

type SeedLinkChannel struct {
	ChannelName string
	ChannelType string
}

type SeedLinkClient struct {
	net.Conn

	sequenceMutex sync.Mutex
	sequence      int64

	Streaming bool
	Network   string
	Station   string
	Location  string
	Channels  []SeedLinkChannel
	StartTime time.Time
	EndTime   time.Time
}

// SeedLink event hooks interface
type SeedLinkHooks interface {
	OnConnection(client *SeedLinkClient)
	OnData(client *SeedLinkClient, data []byte)
	OnClose(client *SeedLinkClient)
	OnCommand(client *SeedLinkClient, command []string)
}

// Station field model of INFO STATIONS command
type SeedLinkStation struct {
	XMLName       xml.Name `xml:"station"`
	BeginSequence string   `xml:"begin_seq,attr"`
	EndSequence   string   `xml:"end_seq,attr"`
	Station       string   `xml:"name,attr"`
	Network       string   `xml:"network,attr"`
	Description   string   `xml:"description,attr"`
}

// Stream field model of INFO STREAMS command
type SeedLinkStream struct {
	XMLName   xml.Name `xml:"stream"`
	BeginTime string   `xml:"begin_time,attr"`
	EndTime   string   `xml:"end_time,attr"`
	SeedName  string   `xml:"seedname,attr"`
	Location  string   `xml:"location,attr"`
	Type      string   `xml:"type,attr"`
	// Exclusive attribute to match station
	Station string `xml:"station,attr"`
}

// Capability field model of INFO CAPABILITY command
type SeedLinkCapability struct {
	XMLName xml.Name `xml:"capability"`
	Name    string   `xml:"name,attr"`
}

// SeedLink data packet model
type SeedLinkDataPacket struct {
	Timestamp  int64
	SampleRate int
	Channel    string
	DataArr    []int32
}

// Provider interface for SeedLink server to get information
type SeedLinkProvider interface {
	GetSoftware() string
	GetStartTime() time.Time
	GetCurrentTime() time.Time
	GetOrganization() string
	GetStations() []SeedLinkStation
	GetStreams() []SeedLinkStream
	GetCapabilities() []SeedLinkCapability
	QueryHistory(startTime, endTime time.Time, channels []SeedLinkChannel) ([]SeedLinkDataPacket, error)
}

// Consumer interface for SeedLink server to stream data
type SeedLinkConsumer interface {
	Subscribe(clientId string, channels []SeedLinkChannel, eventHandler func(SeedLinkDataPacket)) error
	Unsubscribe(clientId string) error
}
