package slgo

import (
	"bufio"
	"strings"

	"github.com/bclswl0827/slgo/handlers"
)

func (s *SeedLinkServer) handleConnection(client *handlers.SeedLinkClient, commands map[string]SeedLinkCommand) {
	s.hooks.OnConnection(client)
	defer s.hooks.OnClose(client)

	// Create a new reader
	reader := bufio.NewReader(client)
	defer client.Close()

	for {
		// Read client message
		clientMessage, err := reader.ReadString('\r')
		if err != nil {
			return
		} else {
			// Remove '\n' & '\r' from message and convert to uppercase
			trimmedMessage := strings.ReplaceAll(clientMessage, "\n", "")
			clientMessage = strings.ToUpper(strings.TrimSuffix(trimmedMessage, "\r"))
		}

		// Ignore empty message
		if len(clientMessage) == 0 {
			continue
		}

		// Disconnect if BYE received
		if clientMessage == "BYE" {
			return
		}

		// Exit from stream mode except INFO command
		if client.Streaming && clientMessage != "END" && !strings.Contains(clientMessage, "INFO ") {
			s.Consumer.Unsubscribe(client.RemoteAddr().String())
			client.Streaming = false
		}

		// Get command details from command list
		var (
			isCommandValid = true
			argumentList   = strings.Split(clientMessage, " ")
			mainArgument   = argumentList[0]
		)
		cmd, ok := commands[mainArgument]
		if !ok {
			isCommandValid = false
		}

		// Send error if command is invalid
		if !isCommandValid {
			client.Write([]byte(handlers.RES_ERR))
		} else {
			s.hooks.OnCommand(client, argumentList)
		}

		// Check for extra arguments
		if isCommandValid && cmd.HasArgs {
			if len(argumentList) == 0 {
				client.Write([]byte(handlers.RES_ERR))
			} else {
				args := argumentList[1:]
				for i := 0; i < len(args); i++ {
					if len(args[i]) == 0 {
						args = append(args[:i], args[i+1:]...)
					}
				}
				err = cmd.Handler.Callback(client, s.Provider, s.Consumer, args...)
				if err != nil {
					cmd.Handler.Fallback(client, s.Provider, s.Consumer, args...)
				}
			}
		} else if isCommandValid {
			err = cmd.Handler.Callback(client, s.Provider, s.Consumer)
			if err != nil {
				cmd.Handler.Fallback(client, s.Provider, s.Consumer)
			}
		}
	}
}
