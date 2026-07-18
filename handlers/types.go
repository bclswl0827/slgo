package handlers

import (
	"bytes"
	"encoding/xml"
	"io"
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
	writeMutex    sync.Mutex
	stateMutex    sync.RWMutex
	batch         bool
	onData        func(*SeedLinkClient, []byte)

	Streaming bool
	Network   string
	Station   string
	Location  string
	Channels  []SeedLinkChannel
	StartTime time.Time
	EndTime   time.Time
}

// Write serializes writes to a client connection. Streaming callbacks and
// command handlers may write concurrently, but SeedLink records must never be
// interleaved on the wire.
func (c *SeedLinkClient) Write(data []byte) (int, error) {
	if c == nil || c.Conn == nil {
		return 0, net.ErrClosed
	}

	c.stateMutex.RLock()
	suppress := c.batch && (bytes.Equal(data, []byte(RES_OK)) || bytes.Equal(data, []byte(RES_ERR)))
	hook := c.onData
	c.stateMutex.RUnlock()
	if suppress {
		return len(data), nil
	}

	c.writeMutex.Lock()
	written := 0
	for written < len(data) {
		n, err := c.Conn.Write(data[written:])
		written += n
		if err != nil {
			c.writeMutex.Unlock()
			if hook != nil && written > 0 {
				hook(c, data[:written])
			}
			return written, err
		}
		if n == 0 {
			c.writeMutex.Unlock()
			return written, io.ErrNoProgress
		}
	}
	c.writeMutex.Unlock()

	if hook != nil && written > 0 {
		hook(c, data[:written])
	}
	return written, nil
}

// SetDataHandler configures the callback invoked after bytes are written to
// the connection. It is primarily used by SeedLinkServer to implement OnData.
func (c *SeedLinkClient) SetDataHandler(handler func(*SeedLinkClient, []byte)) {
	c.stateMutex.Lock()
	c.onData = handler
	c.stateMutex.Unlock()
}

func (c *SeedLinkClient) enableBatchMode() {
	c.stateMutex.Lock()
	c.batch = true
	c.stateMutex.Unlock()
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
