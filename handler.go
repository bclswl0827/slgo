package slgo

import (
	"bufio"
	"context"
	"errors"
	"io"
	"strings"

	"github.com/bclswl0827/slgo/handlers"
)

var errCommandTooLong = errors.New("seedlink command too long")

func (s *SeedLinkServer) handleConnection(ctx context.Context, client *handlers.SeedLinkClient, commands map[string]SeedLinkCommand) {
	if s.hooks != nil {
		client.SetDataHandler(s.hooks.OnData)
		s.hooks.OnConnection(client)
		defer s.hooks.OnClose(client)
	}

	reader := bufio.NewReader(client)
	connectionDone := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			_ = client.Close()
		case <-connectionDone:
		}
	}()
	defer close(connectionDone)
	defer func() {
		if client.Streaming {
			_ = s.Consumer.Unsubscribe(client.RemoteAddr().String())
			client.Streaming = false
		}
		_ = client.Close()
	}()

	for {
		clientMessage, err := readCommand(reader)
		if err != nil {
			if errors.Is(err, errCommandTooLong) {
				_, _ = client.Write([]byte(handlers.RES_ERR))
			}
			if errors.Is(err, io.EOF) || ctx.Err() != nil {
				return
			}
			return
		}

		argumentList := strings.Fields(clientMessage)
		if len(argumentList) == 0 {
			continue
		}
		mainArgument := strings.ToUpper(argumentList[0])
		argumentList[0] = mainArgument

		// INFO and BYE are the only commands accepted while streaming.
		if mainArgument == "BYE" {
			return
		}
		if client.Streaming && mainArgument != "INFO" {
			_, _ = client.Write([]byte(handlers.RES_ERR))
			continue
		}

		cmd, ok := commands[mainArgument]
		if !ok {
			_, _ = client.Write([]byte(handlers.RES_ERR))
			continue
		}
		if s.hooks != nil {
			s.hooks.OnCommand(client, argumentList)
		}

		args := argumentList[1:]
		if err := cmd.Handler.Callback(client, s.Provider, s.Consumer, args...); err != nil {
			cmd.Handler.Fallback(client, s.Provider, s.Consumer, args...)
		}
	}
}

// readCommand accepts CR, LF, and CRLF terminators. A following terminator is
// intentionally treated as an empty command, as required by SeedLink.
func readCommand(reader *bufio.Reader) (string, error) {
	command := make([]byte, 0, 128)
	for {
		b, err := reader.ReadByte()
		if err != nil {
			return "", err
		}
		if b == '\r' || b == '\n' {
			return string(command), nil
		}
		if len(command) >= 4096 {
			return "", errCommandTooLong
		}
		command = append(command, b)
	}
}
