package handlers

import (
	"strings"
)

type SELECT struct{}

// Callback of "SELECT <...>" command, implements handler interface
func (*SELECT) Callback(client *SeedLinkClient, provider SeedLinkProvider, consumer SeedLinkConsumer, args ...string) error {
	if len(args) != 1 {
		_, err := client.Write([]byte(RES_ERR))
		return err
	}

	var channel SeedLinkChannel
	locationAndChannel := args[0]
	if selector, subtype, found := strings.Cut(locationAndChannel, "."); found {
		if selector == "" || subtype == "" || strings.Contains(subtype, ".") {
			_, err := client.Write([]byte(RES_ERR))
			return err
		}
		locationAndChannel = selector
		channel.ChannelType = subtype
	}

	if !validCode(locationAndChannel, "?*_-!") {
		_, err := client.Write([]byte(RES_ERR))
		return err
	}
	switch len(locationAndChannel) {
	case 3:
		client.Location = "00"
		channel.ChannelName = locationAndChannel
	case 5:
		client.Location = locationAndChannel[:2]
		channel.ChannelName = locationAndChannel[2:5]
	default:
		_, err := client.Write([]byte(RES_ERR))
		return err
	}
	client.Channels = append(client.Channels, channel)

	_, err := client.Write([]byte(RES_OK))
	return err
}

// Fallback of "SELECT <...>" command, implements handler interface
func (*SELECT) Fallback(client *SeedLinkClient, provider SeedLinkProvider, consumer SeedLinkConsumer, args ...string) {
	client.Close()
}
