package handlers

import (
	"errors"
	"time"
)

type END struct {
	DataType int
}

// Callback of "END" command, implements handler interface
func (e *END) Callback(client *SeedLinkClient, provider SeedLinkProvider, consumer SeedLinkConsumer, args ...string) error {
	if client.StartTime.IsZero() {
		client.Write([]byte(RES_ERR))
		return errors.New("start time not set")
	}

	// Query history data from database
	if !client.EndTime.IsZero() {
		historyRecords, err := provider.QueryHistory(client.StartTime, client.EndTime.Add(10*time.Second), client.Channels)
		if err != nil {
			client.Write([]byte(RES_ERR))
			return err
		}
		for _, dataPacket := range historyRecords {
			err = SendSeedLinkPacket(client, dataPacket, e.DataType)
			if err != nil {
				client.Write([]byte(RES_ERR))
				return err
			}
		}
	}

	// Subscribe to the message queue
	client.Streaming = true
	return consumer.Subscribe(
		client.RemoteAddr().String(),
		client.Channels,
		func(data SeedLinkDataPacket) {
			err := SendSeedLinkPacket(client, data, e.DataType)
			if err != nil {
				consumer.Unsubscribe(client.RemoteAddr().String())
				client.Write([]byte(RES_ERR))
				client.Close()
			}
		},
	)
}

// Fallback of "END" command, implements handler interface
func (*END) Fallback(client *SeedLinkClient, provider SeedLinkProvider, consumer SeedLinkConsumer, args ...string) {
	client.Close()
}
