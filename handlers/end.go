package handlers

import "errors"

type END struct{}

// Callback of "END" command, implements handler interface
func (*END) Callback(client *SeedLinkClient, provider SeedLinkProvider, consumer SeedLinkConsumer, args ...string) error {
	if client.StartTime.IsZero() {
		client.Write([]byte(RES_ERR))
		return errors.New("start time not set")
	}

	// Query history data from database
	historyRecords, err := provider.QueryHistory(client.StartTime, client.EndTime, client.Channels)
	if err != nil {
		client.Write([]byte(RES_ERR))
		return err
	}
	for _, dataPacket := range historyRecords {
		err = SendSeedLinkPacket(client, dataPacket)
		if err != nil {
			return err
		}
	}

	// Subscribe to the message queue
	client.Streaming = true
	return consumer.Subscribe(
		client.RemoteAddr().String(),
		client.Channels,
		func(data SeedLinkDataPacket) {
			err := SendSeedLinkPacket(client, data)
			if err != nil {
				consumer.Unsubscribe(client.RemoteAddr().String())
				client.Close()
			}
		},
	)
}

// Fallback of "END" command, implements handler interface
func (*END) Fallback(client *SeedLinkClient, provider SeedLinkProvider, consumer SeedLinkConsumer, args ...string) {
	client.Close()
}
