package handlers

import (
	"encoding/xml"
	"net"
	"time"
)

// The Seedlink protocol specifies that the maximum length of a data packet is 512 bytes.
// which can accommodate approximately 100 samples (int32, 4 bytes each).
// Samples more than 100 will be sent in chunks of 100.
const CHUNK_SIZE = 100

// SeedLink handshake constant flags
const RELEASE = "SeedLink v3.1 AnyShake Edition (Basic implementation in Go, repository: https://github.com/bclswl0827/slgo) :: SLPROTO:3.1 CAP EXTREPLY NSWILDCARD BATCH WS:13 :: Constructing Realtime Seismic Network Ambitiously."

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

type SeedLinkClient struct {
	net.Conn
	Streaming bool
	Sequence  int64
	Network   string
	Station   string
	Location  string
	Channels  []string
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
	QueryHistory(startTime, endTime time.Time, channels []string) ([]SeedLinkDataPacket, error)
}

// Consumer interface for SeedLink server to stream data
type SeedLinkConsumer interface {
	Subscribe(clientId string, eventHandler func(SeedLinkDataPacket)) error
	Unsubscribe(clientId string) error
}
