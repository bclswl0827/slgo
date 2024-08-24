package handlers

import (
	"fmt"
	"time"

	"github.com/bclswl0827/mseedio"
)

func SendSeedLinkPacket(client *SeedLinkClient, data SeedLinkDataPacket) error {
	// Create data chunks to adapt to SeedLink packet size
	var countGroup [][]int32
	if len(data.DataArr) > CHUNK_SIZE {
		for i := 0; i < len(data.DataArr); i += CHUNK_SIZE {
			if i+CHUNK_SIZE > len(data.DataArr) {
				countGroup = append(countGroup, data.DataArr[i:])
			} else {
				countGroup = append(countGroup, data.DataArr[i:i+CHUNK_SIZE])
			}
		}
	} else {
		countGroup = append(countGroup, data.DataArr)
	}

	dataSpanMs := 1000 / data.SampleRate
	for i, c := range countGroup {
		// Generate MiniSEED record
		var miniseed mseedio.MiniSeedData
		miniseed.Init(mseedio.STEIM2, mseedio.MSBFIRST)
		err := miniseed.Append(c, &mseedio.AppendOptions{
			ChannelCode:    data.Channel,
			StationCode:    client.Station,
			LocationCode:   client.Location,
			NetworkCode:    client.Network,
			SampleRate:     float64(data.SampleRate),
			SequenceNumber: fmt.Sprintf("%06d", client.Sequence),
			StartTime:      time.UnixMilli(data.Timestamp + int64(i*CHUNK_SIZE*dataSpanMs)).UTC(),
		})
		if err != nil {
			return err
		}

		// Get MiniSEED data bytes always in 512 bytes
		miniseed.Series[0].BlocketteSection.RecordLength = 9
		slData, err := miniseed.Encode(mseedio.OVERWRITE, mseedio.MSBFIRST)
		if err != nil {
			return err
		}

		// Prepend and send SeedLink sequence number
		slSeq := []byte(fmt.Sprintf("SL%06X", client.Sequence))
		_, err = client.Write(slSeq)
		if err != nil {
			return err
		}

		// Send SeedLink packet data
		_, err = client.Write(slData)
		if err != nil {
			return err
		}

		client.Sequence++
	}

	return nil
}
