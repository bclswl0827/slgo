package slgo

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/bclswl0827/mseedio"
	"github.com/bclswl0827/slgo/handlers"
)

func (s *SeedLinkServer) Start(ctx context.Context, host string, port int, compress bool) error {
	if ctx == nil {
		return errors.New("slgo: nil context")
	}
	if s.Provider == nil {
		return errors.New("slgo: nil provider")
	}
	if s.Consumer == nil {
		return errors.New("slgo: nil consumer")
	}
	if err := ctx.Err(); err != nil {
		return nil
	}

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
	commands := builtinCommands(packetDataType)

	serveCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	listenerDone := make(chan struct{})
	go func() {
		select {
		case <-serveCtx.Done():
			_ = listener.Close()
		case <-listenerDone:
		}
	}()
	defer close(listenerDone)

	var clients sync.WaitGroup
	for {
		conn, acceptErr := listener.Accept()
		if acceptErr != nil {
			cancel()
			_ = listener.Close()
			clients.Wait()
			if ctx.Err() != nil || errors.Is(acceptErr, net.ErrClosed) {
				return nil
			}
			return fmt.Errorf("slgo: accept connection: %w", acceptErr)
		}

		client := &handlers.SeedLinkClient{Conn: conn}
		clients.Add(1)
		go func() {
			defer clients.Done()
			s.handleConnection(serveCtx, client, commands)
		}()
	}
}

func builtinCommands(packetDataType int) map[string]SeedLinkCommand {
	return map[string]SeedLinkCommand{
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
}
