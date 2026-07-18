package slgo

import (
	"bufio"
	"context"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/bclswl0827/mseedio"
	"github.com/bclswl0827/slgo/handlers"
)

func TestHandleConnectionParsesCommandsWithoutCorruptingArguments(t *testing.T) {
	serverConn, peer := net.Pipe()
	defer peer.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hooks := &recordingHooks{}
	server := New(serverProvider{}, &serverConsumer{}, hooks)
	done := make(chan struct{})
	go func() {
		server.handleConnection(ctx, &handlers.SeedLinkClient{Conn: serverConn}, builtinCommands(mseedio.INT32))
		close(done)
	}()
	reader := bufio.NewReader(peer)

	writeCommand(t, peer, "STATION\r\n")
	readReply(t, reader, handlers.RES_ERR)
	writeCommand(t, peer, "station Sta nW\n")
	readReply(t, reader, handlers.RES_OK)
	writeCommand(t, peer, "select 00eHz.d\r")
	readReply(t, reader, handlers.RES_OK)

	hooks.Lock()
	client := hooks.client
	commands := append([][]string(nil), hooks.commands...)
	hooks.Unlock()
	if client == nil || client.Station != "Sta" || client.Network != "nW" {
		t.Fatalf("station arguments were altered: %+v", client)
	}
	if len(client.Channels) != 1 || client.Channels[0].ChannelName != "eHz" || client.Channels[0].ChannelType != "d" {
		t.Fatalf("selector arguments were altered: %+v", client.Channels)
	}
	if got := commands[len(commands)-1]; got[0] != "SELECT" || got[1] != "00eHz.d" {
		t.Fatalf("hook command = %#v", got)
	}

	writeCommand(t, peer, "hello\r\n")
	if line, err := reader.ReadString('\n'); err != nil || line == "" {
		t.Fatalf("HELLO release line = %q, %v", line, err)
	}
	if line, err := reader.ReadString('\n'); err != nil || line != "org\r\n" {
		t.Fatalf("HELLO organization line = %q, %v", line, err)
	}

	writeCommand(t, peer, "BYE\r\n")
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("connection did not close after BYE")
	}
	if hooks.dataWrites == 0 {
		t.Fatal("OnData was not invoked")
	}
}

func TestBatchSuppressesNegotiationReplies(t *testing.T) {
	serverConn, peer := net.Pipe()
	defer peer.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	server := New(serverProvider{}, &serverConsumer{}, nil)
	done := make(chan struct{})
	go func() {
		server.handleConnection(ctx, &handlers.SeedLinkClient{Conn: serverConn}, builtinCommands(mseedio.INT32))
		close(done)
	}()
	reader := bufio.NewReader(peer)

	writeCommand(t, peer, "BATCH\r\n")
	readReply(t, reader, handlers.RES_OK)
	writeCommand(t, peer, "STATION STA NW\r\n")
	if err := peer.SetReadDeadline(time.Now().Add(50 * time.Millisecond)); err != nil {
		t.Fatal(err)
	}
	buffer := make([]byte, 1)
	if _, err := peer.Read(buffer); err == nil {
		t.Fatal("received a reply after BATCH")
	} else if netErr, ok := err.(net.Error); !ok || !netErr.Timeout() {
		t.Fatalf("read error = %v, want timeout", err)
	}
	_ = peer.SetReadDeadline(time.Time{})
	writeCommand(t, peer, "BYE\n")
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("connection did not close")
	}
}

func TestCancellationClosesStreamingClientAndUnsubscribes(t *testing.T) {
	serverConn, peer := net.Pipe()
	defer peer.Close()
	ctx, cancel := context.WithCancel(context.Background())
	consumer := &serverConsumer{subscribed: make(chan struct{}), unsubscribed: make(chan struct{})}
	server := New(serverProvider{}, consumer, nil)
	done := make(chan struct{})
	go func() {
		server.handleConnection(ctx, &handlers.SeedLinkClient{Conn: serverConn}, builtinCommands(mseedio.INT32))
		close(done)
	}()

	writeCommand(t, peer, "END\r\n")
	select {
	case <-consumer.subscribed:
	case <-time.After(time.Second):
		t.Fatal("END did not subscribe")
	}
	cancel()
	select {
	case <-consumer.unsubscribed:
	case <-time.After(time.Second):
		t.Fatal("cancellation did not unsubscribe")
	}
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("cancellation did not stop connection handler")
	}
}

func writeCommand(t *testing.T, conn net.Conn, command string) {
	t.Helper()
	if _, err := conn.Write([]byte(command)); err != nil {
		t.Fatal(err)
	}
}

func readReply(t *testing.T, reader *bufio.Reader, want string) {
	t.Helper()
	got, err := reader.ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("reply = %q, want %q", got, want)
	}
}

type serverProvider struct{}

func (serverProvider) GetSoftware() string                            { return "test" }
func (serverProvider) GetStartTime() time.Time                        { return time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC) }
func (serverProvider) GetCurrentTime() time.Time                      { return time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC) }
func (serverProvider) GetOrganization() string                        { return "org" }
func (serverProvider) GetStations() []handlers.SeedLinkStation        { return nil }
func (serverProvider) GetStreams() []handlers.SeedLinkStream          { return nil }
func (serverProvider) GetCapabilities() []handlers.SeedLinkCapability { return nil }
func (serverProvider) QueryHistory(time.Time, time.Time, []handlers.SeedLinkChannel) ([]handlers.SeedLinkDataPacket, error) {
	return nil, nil
}

type serverConsumer struct {
	once         sync.Once
	subscribed   chan struct{}
	unsubscribed chan struct{}
}

func (c *serverConsumer) Subscribe(string, []handlers.SeedLinkChannel, func(handlers.SeedLinkDataPacket)) error {
	if c.subscribed != nil {
		c.once.Do(func() { close(c.subscribed) })
	}
	return nil
}
func (c *serverConsumer) Unsubscribe(string) error {
	if c.unsubscribed != nil {
		select {
		case <-c.unsubscribed:
		default:
			close(c.unsubscribed)
		}
	}
	return nil
}

type recordingHooks struct {
	sync.Mutex
	client     *handlers.SeedLinkClient
	commands   [][]string
	dataWrites int
}

func (h *recordingHooks) OnConnection(client *handlers.SeedLinkClient) {
	h.Lock()
	h.client = client
	h.Unlock()
}
func (h *recordingHooks) OnData(_ *handlers.SeedLinkClient, _ []byte) {
	h.Lock()
	h.dataWrites++
	h.Unlock()
}
func (*recordingHooks) OnClose(*handlers.SeedLinkClient) {}
func (h *recordingHooks) OnCommand(_ *handlers.SeedLinkClient, command []string) {
	h.Lock()
	h.commands = append(h.commands, append([]string(nil), command...))
	h.Unlock()
}
