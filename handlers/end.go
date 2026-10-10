package handlers

import (
	"sort"
	"sync"
	"time"
)

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
		sentThrough   = make(map[string]time.Time)
	)
	deliver := func(data SeedLinkDataPacket) error {
		startTime := time.UnixMilli(data.Timestamp).UTC()
		if data.SampleRate > 0 && len(data.DataArr) > 0 {
			previousEnd := sentThrough[data.Channel]
			if previousEnd.Before(client.StartTime) {
				previousEnd = client.StartTime
			}
			if previousEnd.After(startTime) {
				// Millisecond timestamps can round a contiguous record back by less
				// than one millisecond. Preserve its samples and correct the time.
				if previousEnd.Sub(startTime) < time.Millisecond && data.SampleRate < 1000 {
					startTime = previousEnd
				} else {
					// Find the first sample at or after the end of this channel's
					// previous record. Search avoids overflowing rate * duration.
					skip := sort.Search(len(data.DataArr), func(i int) bool {
						return !sampleTime(startTime, i, data.SampleRate).Before(previousEnd)
					})
					if skip >= len(data.DataArr) {
						return nil
					}
					startTime = sampleTime(startTime, skip, data.SampleRate)
					data.DataArr = data.DataArr[skip:]
				}
			}
		}
		newSeq, dataBytes, recordEnd, err := sendSeedLinkPacketAt(
			station, location, network, e.DataType, client.GetSequence(), data, startTime,
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
		sentThrough[data.Channel] = recordEnd
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
		historyRecords = append([]SeedLinkDataPacket(nil), historyRecords...)
		sort.SliceStable(historyRecords, func(i, j int) bool {
			return historyRecords[i].Timestamp < historyRecords[j].Timestamp
		})
	}

	deliveryMutex.Lock()
	for _, dataPacket := range historyRecords {
		if err = deliver(dataPacket); err != nil {
			break
		}
	}
	sort.SliceStable(pending, func(i, j int) bool {
		return pending[i].Timestamp < pending[j].Timestamp
	})
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
