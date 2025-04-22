package handlers

import (
	"strings"
)

type SELECT struct{}

// Callback of "SELECT <...>" command, implements handler interface
func (*SELECT) Callback(client *SeedLinkClient, provider SeedLinkProvider, consumer SeedLinkConsumer, args ...string) error {
	if len(args) == 0 {
		_, err := client.Write([]byte(RES_ERR))
		return err
	}

	locationAndChannel := args[0]
	if len(locationAndChannel) < 5 {
		_, err := client.Write([]byte(RES_ERR))
		return err
	}
	if strings.Contains(locationAndChannel, ".") {
		locationAndChannel = strings.Split(locationAndChannel, ".")[0]
	}

	if len(locationAndChannel) == 3 {
		client.Location = "00"
		client.Channels = append(client.Channels, locationAndChannel)
	} else if len(locationAndChannel) == 5 {
		client.Location = locationAndChannel[:2]
		client.Channels = append(client.Channels, locationAndChannel[2:5])
	}

	_, err := client.Write([]byte(RES_OK))
	return err
}

// Fallback of "SELECT <...>" command, implements handler interface
func (*SELECT) Fallback(client *SeedLinkClient, provider SeedLinkProvider, consumer SeedLinkConsumer, args ...string) {
	client.Close()
}
