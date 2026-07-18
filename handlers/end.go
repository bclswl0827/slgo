package handlers

import "sync"

type END struct {
	DataType int
}

// Callback of "END" command, implements handler interface
func (e *END) Callback(client *SeedLinkClient, provider SeedLinkProvider, consumer SeedLinkConsumer, args ...string) error {
	var (
		station  = client.Station
		location = client.Location
		network  = client.Network
	)
	clientID := client.RemoteAddr().String()

	var (
		deliveryMutex sync.Mutex
		pending       []SeedLinkDataPacket
		historyReady  bool
		failed        bool
	)
	deliver := func(data SeedLinkDataPacket) error {
		newSeq, dataBytes, err := SendSeedLinkPacket(
			station, location, network, e.DataType, client.GetSequence(), data,
		)
		if err != nil {
			return err
		}
		if len(dataBytes) > 0 {
			if _, err = client.Write(dataBytes); err != nil {
				return err
			}
		}
		client.SetSequence(newSeq)
		return nil
	}

	// Subscribe before querying history so live packets arriving during the
	// query can be queued and delivered after the historical window.
	err := consumer.Subscribe(
		clientID,
		client.Channels,
		func(data SeedLinkDataPacket) {
			deliveryMutex.Lock()
			defer deliveryMutex.Unlock()
			if failed {
				return
			}
			if !historyReady {
				data.DataArr = append([]int32(nil), data.DataArr...)
				pending = append(pending, data)
				return
			}
			if err := deliver(data); err != nil {
				failed = true
				_ = client.Close()
			}
		},
	)
	if err != nil {
		_, _ = client.Write([]byte(RES_ERR))
		return err
	}
	client.Streaming = true

	var historyRecords []SeedLinkDataPacket
	if !client.StartTime.IsZero() {
		endTime := client.EndTime
		if endTime.IsZero() {
			endTime = provider.GetCurrentTime()
		}
		historyRecords, err = provider.QueryHistory(client.StartTime, endTime, client.Channels)
		if err != nil {
			deliveryMutex.Lock()
			failed = true
			deliveryMutex.Unlock()
			_ = consumer.Unsubscribe(clientID)
			client.Streaming = false
			_, _ = client.Write([]byte(RES_ERR))
			return err
		}
	}

	deliveryMutex.Lock()
	for _, dataPacket := range historyRecords {
		if err = deliver(dataPacket); err != nil {
			break
		}
	}
	for _, dataPacket := range pending {
		if err != nil {
			break
		}
		err = deliver(dataPacket)
	}
	pending = nil
	if err == nil {
		historyReady = true
	} else {
		failed = true
	}
	deliveryMutex.Unlock()
	if err != nil {
		_ = consumer.Unsubscribe(clientID)
		client.Streaming = false
		return err
	}

	return nil
}

// Fallback of "END" command, implements handler interface
func (*END) Fallback(client *SeedLinkClient, provider SeedLinkProvider, consumer SeedLinkConsumer, args ...string) {
	client.Close()
}
