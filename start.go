package slgo

import (
	"context"
	"fmt"
	"net"

	"github.com/bclswl0827/mseedio"
	"github.com/bclswl0827/slgo/handlers"
)

func (s *SeedLinkServer) Start(ctx context.Context, host string, port int, compress bool) error {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return err
	}
	defer listener.Close()

	// Set packet data type to INT32 or STEIM2
	packetDataType := mseedio.INT32
	if compress {
		packetDataType = mseedio.STEIM2
	}

	// Builtin implementation of command handlers
	commands := map[string]SeedLinkCommand{
		"END":          {HasArgs: false, Handler: &handlers.END{DataType: packetDataType}},
		"BATCH":        {HasArgs: false, Handler: &handlers.BATCH{}},
		"DATA":         {HasArgs: true, Handler: &handlers.DATA{}},
		"TIME":         {HasArgs: true, Handler: &handlers.TIME{}},
		"INFO":         {HasArgs: true, Handler: &handlers.INFO{}},
		"HELLO":        {HasArgs: false, Handler: &handlers.HELLO{}},
		"SELECT":       {HasArgs: true, Handler: &handlers.SELECT{}},
		"STATION":      {HasArgs: true, Handler: &handlers.STATION{}},
		"CAPABILITIES": {HasArgs: true, Handler: &handlers.CAPABILITIES{}},
	}

	done := make(chan struct{})

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				select {
				case <-ctx.Done():
					close(done)
					return
				default:
					continue
				}
			}

			client := handlers.SeedLinkClient{Conn: conn}
			go s.handleConnection(ctx, &client, commands)
		}
	}()

	<-ctx.Done()
	listener.Close()
	<-done

	return nil
}
