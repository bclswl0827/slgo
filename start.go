package slgo

import (
	"fmt"
	"net"

	"github.com/bclswl0827/slgo/handlers"
)

func (s *SeedLinkServer) Start(host string, port int) error {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return err
	}
	defer listener.Close()

	// Builtin implementation of command handlers
	commands := map[string]SeedLinkCommand{
		"END":          {HasArgs: false, Handler: &handlers.END{}},
		"DATA":         {HasArgs: true, Handler: &handlers.DATA{}},
		"TIME":         {HasArgs: true, Handler: &handlers.TIME{}},
		"INFO":         {HasArgs: true, Handler: &handlers.INFO{}},
		"HELLO":        {HasArgs: false, Handler: &handlers.HELLO{}},
		"SELECT":       {HasArgs: true, Handler: &handlers.SELECT{}},
		"STATION":      {HasArgs: true, Handler: &handlers.STATION{}},
		"CAPABILITIES": {HasArgs: true, Handler: &handlers.CAPABILITIES{}},
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}

		client := handlers.SeedLinkClient{Conn: conn}
		go s.handleConnection(&client, commands)
	}
}
