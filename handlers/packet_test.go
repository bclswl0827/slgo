package handlers

import (
	"bytes"
	"encoding/binary"
	"encoding/xml"
	"errors"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/bclswl0827/mseedio"
)

func TestSendSeedLinkPacketFramingSequenceAndTiming(t *testing.T) {
	values := make([]int32, 101)
	for index := range values {
		values[index] = int32(index)
	}
	started := time.Date(2024, time.January, 2, 3, 4, 5, 0, time.UTC)

	next, packet, err := SendSeedLinkPacket("STA", "00", "NW", mseedio.INT32, 0xFFFFFF, SeedLinkDataPacket{
		Timestamp:  started.UnixMilli(),
		SampleRate: 30,
		Channel:    "BHZ",
		DataArr:    values,
	})
	if err != nil {
		t.Fatalf("SendSeedLinkPacket() error = %v", err)
	}
	if next != 0x1000001 {
		t.Fatalf("next sequence = %X, want 1000001", next)
	}
	if len(packet) != 2*(8+512) {
		t.Fatalf("packet length = %d, want 1040", len(packet))
	}
	if got := string(packet[:8]); got != "SLFFFFFF" {
		t.Fatalf("first header = %q", got)
	}
	if got := string(packet[520:528]); got != "SL000000" {
		t.Fatalf("wrapped header = %q", got)
	}

	var first, second mseedio.FixedSection
	if err := first.Parse(packet[8:56], mseedio.MSBFIRST); err != nil {
		t.Fatal(err)
	}
	if err := second.Parse(packet[528:576], mseedio.MSBFIRST); err != nil {
		t.Fatal(err)
	}
	if first.SamplesNumber != 100 || second.SamplesNumber != 1 {
		t.Fatalf("sample counts = %d, %d", first.SamplesNumber, second.SamplesNumber)
	}
	wantSecond := started.Add(100 * time.Second / 30).Truncate(100 * time.Microsecond)
	if !second.StartTime.Equal(wantSecond) {
		t.Fatalf("second start = %s, want %s", second.StartTime, wantSecond)
	}
}

func TestSendSeedLinkPacketRejectsInvalidData(t *testing.T) {
	tests := []SeedLinkDataPacket{
		{SampleRate: 0, Channel: "BHZ", DataArr: []int32{1}},
		{SampleRate: 1, Channel: "BHZ"},
		{SampleRate: 1, DataArr: []int32{1}},
	}
	for _, data := range tests {
		if _, _, err := SendSeedLinkPacket("STA", "00", "NW", mseedio.INT32, 0, data); err == nil {
			t.Fatalf("SendSeedLinkPacket(%+v) returned no error", data)
		}
	}
	if _, _, err := SendSeedLinkPacket("STA", "00", "NW", mseedio.INT32, -1, SeedLinkDataPacket{
		SampleRate: 1, Channel: "BHZ", DataArr: []int32{1},
	}); err == nil {
		t.Fatal("negative sequence returned no error")
	}
}

func TestParseSeedLinkTimeSupportsFractionalSeconds(t *testing.T) {
	parsed, err := parseSeedLinkTime("2024,01,02,03,04,05.125")
	if err != nil {
		t.Fatal(err)
	}
	want := time.Date(2024, 1, 2, 3, 4, 5, int(125*time.Millisecond), time.UTC)
	if !parsed.Equal(want) {
		t.Fatalf("parsed time = %s, want %s", parsed, want)
	}
	for _, invalid := range []string{"2024,02,30,00,00,00", "2024,01,01,00,00,60"} {
		if _, err := parseSeedLinkTime(invalid); err == nil {
			t.Fatalf("parseSeedLinkTime(%q) returned no error", invalid)
		}
	}
}

func TestInfoResponseDoesNotAddEmptyRecord(t *testing.T) {
	body := bytes.Repeat([]byte{'x'}, 512-mseedio.FIXED_SECTION_LENGTH-mseedio.BLOCKETTE100X_SECTION_LENGTH)
	response, err := (&INFO{}).setResponse(body, FLAG_INF, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	if len(response) != 520 {
		t.Fatalf("response length = %d, want one framed record", len(response))
	}
	if got := string(response[:8]); got != "SLINFO  " {
		t.Fatalf("INFO header = %q", got)
	}
}

func TestInfoStreamsProducesStandardXML(t *testing.T) {
	provider := testProvider{}
	response, err := (&INFO{}).getStreams(provider)
	if err != nil {
		t.Fatal(err)
	}
	body := unpackInfoBody(t, response)

	var document struct {
		XMLName      xml.Name `xml:"seedlink"`
		Software     string   `xml:"software,attr"`
		Started      string   `xml:"started,attr"`
		Organization string   `xml:"organization,attr"`
		Stations     []struct {
			Name        string `xml:"name,attr"`
			StreamCheck string `xml:"stream_check,attr"`
			Streams     []struct {
				SeedName string `xml:"seedname,attr"`
			} `xml:"stream"`
		} `xml:"station"`
	}
	if err := xml.Unmarshal(body, &document); err != nil {
		t.Fatalf("invalid INFO XML %q: %v", body, err)
	}
	if document.XMLName.Local != "seedlink" || document.Started != "2024-01-02 03:04:05" {
		t.Fatalf("unexpected root: %+v", document)
	}
	if len(document.Stations) != 1 || document.Stations[0].StreamCheck != "enabled" ||
		len(document.Stations[0].Streams) != 1 || document.Stations[0].Streams[0].SeedName != "BHZ" {
		t.Fatalf("unexpected station XML: %+v", document.Stations)
	}
}

func TestEndSendsHistoryBeforeQueuedLiveData(t *testing.T) {
	conn := &memoryConn{}
	client := &SeedLinkClient{
		Conn:      conn,
		Station:   "STA",
		Network:   "NW",
		Location:  "00",
		StartTime: time.Date(2024, 1, 2, 3, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2024, 1, 2, 3, 1, 0, 0, time.UTC),
	}
	consumer := &testConsumer{onSubscribe: func(handler func(SeedLinkDataPacket)) {
		handler(SeedLinkDataPacket{Timestamp: client.EndTime.UnixMilli(), SampleRate: 1, Channel: "BHZ", DataArr: []int32{2}})
	}}

	if err := (&END{DataType: mseedio.INT32}).Callback(client, testProvider{}, consumer); err != nil {
		t.Fatal(err)
	}
	output := conn.Bytes()
	if len(output) != 1040 {
		t.Fatalf("stream length = %d, want 1040", len(output))
	}
	first := int32(binary.BigEndian.Uint32(output[72:76]))
	second := int32(binary.BigEndian.Uint32(output[592:596]))
	if first != 1 || second != 2 {
		t.Fatalf("sample order = %d, %d; want history then live", first, second)
	}
}

func unpackInfoBody(t *testing.T, response []byte) []byte {
	t.Helper()
	var body []byte
	for offset := 0; offset < len(response); offset += 520 {
		var fixed mseedio.FixedSection
		if err := fixed.Parse(response[offset+8:offset+56], mseedio.MSBFIRST); err != nil {
			t.Fatal(err)
		}
		start := offset + 8 + 64
		body = append(body, response[start:start+int(fixed.SamplesNumber)]...)
	}
	return body
}

type testProvider struct{}

func (testProvider) GetSoftware() string       { return "test-server" }
func (testProvider) GetStartTime() time.Time   { return time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC) }
func (testProvider) GetCurrentTime() time.Time { return time.Date(2024, 1, 2, 4, 0, 0, 0, time.UTC) }
func (testProvider) GetOrganization() string   { return "test-org" }
func (testProvider) GetCapabilities() []SeedLinkCapability {
	return []SeedLinkCapability{{Name: "info:id"}}
}
func (testProvider) GetStations() []SeedLinkStation {
	return []SeedLinkStation{{Station: "STA", Network: "NW", BeginSequence: "0", EndSequence: "1"}}
}
func (testProvider) GetStreams() []SeedLinkStream {
	return []SeedLinkStream{{Station: "STA", Location: "00", SeedName: "BHZ", Type: "D"}}
}
func (testProvider) QueryHistory(time.Time, time.Time, []SeedLinkChannel) ([]SeedLinkDataPacket, error) {
	return []SeedLinkDataPacket{{Timestamp: time.Date(2024, 1, 2, 3, 0, 0, 0, time.UTC).UnixMilli(), SampleRate: 1, Channel: "BHZ", DataArr: []int32{1}}}, nil
}

type testConsumer struct {
	onSubscribe func(func(SeedLinkDataPacket))
}

func (c *testConsumer) Subscribe(_ string, _ []SeedLinkChannel, handler func(SeedLinkDataPacket)) error {
	if c.onSubscribe != nil {
		c.onSubscribe(handler)
	}
	return nil
}
func (*testConsumer) Unsubscribe(string) error { return nil }

type memoryConn struct {
	sync.Mutex
	bytes.Buffer
}

func (c *memoryConn) Read([]byte) (int, error) { return 0, errors.New("not implemented") }
func (c *memoryConn) Write(data []byte) (int, error) {
	c.Lock()
	defer c.Unlock()
	return c.Buffer.Write(data)
}
func (c *memoryConn) Close() error                     { return nil }
func (c *memoryConn) LocalAddr() net.Addr              { return testAddr("local") }
func (c *memoryConn) RemoteAddr() net.Addr             { return testAddr("remote") }
func (c *memoryConn) SetDeadline(time.Time) error      { return nil }
func (c *memoryConn) SetReadDeadline(time.Time) error  { return nil }
func (c *memoryConn) SetWriteDeadline(time.Time) error { return nil }
func (c *memoryConn) Bytes() []byte {
	c.Lock()
	defer c.Unlock()
	return append([]byte(nil), c.Buffer.Bytes()...)
}

type testAddr string

func (a testAddr) Network() string { return "test" }
func (a testAddr) String() string  { return string(a) }
