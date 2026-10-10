package handlers

import (
	"bytes"
	"errors"
	"fmt"
	"time"

	"github.com/bclswl0827/mseedio"
)

func chunkInt32Slice(data []int32, chunkSize int) [][]int32 {
	var chunks [][]int32
	for i := 0; i < len(data); i += chunkSize {
		end := i + chunkSize
		if end > len(data) {
			end = len(data)
		}
		chunks = append(chunks, data[i:end])
	}
	return chunks
}

func SendSeedLinkPacket(station, location, network string, dataType int, sequence int64, data SeedLinkDataPacket) (newSequence int64, packetBuf []byte, err error) {
	next, packet, _, err := sendSeedLinkPacketAt(station, location, network, dataType, sequence, data, time.UnixMilli(data.Timestamp).UTC())
	return next, packet, err
}

// sampleTime keeps the sample index calculation consistent for packet writing
// and for trimming overlapping records at a stream boundary.
func sampleTime(start time.Time, sampleCount, sampleRate int) time.Time {
	return start.Add(time.Duration(sampleCount/sampleRate)*time.Second +
		time.Duration(sampleCount%sampleRate)*time.Second/time.Duration(sampleRate))
}

// MiniSEED fixed headers store time in 100-microsecond units. Round forward so
// a later chunk cannot be encoded as starting before the preceding chunk ends.
func miniSeedStartTime(t time.Time) time.Time {
	start := t.Truncate(100 * time.Microsecond)
	if start.Before(t) {
		start = start.Add(100 * time.Microsecond)
	}
	return start
}

func sendSeedLinkPacketAt(station, location, network string, dataType int, sequence int64, data SeedLinkDataPacket, startTime time.Time) (newSequence int64, packetBuf []byte, endTime time.Time, err error) {
	if sequence < 0 {
		return sequence, nil, time.Time{}, errors.New("sequence number must not be negative")
	}
	if data.SampleRate <= 0 {
		return sequence, nil, time.Time{}, errors.New("sample rate must be greater than zero")
	}
	if len(data.DataArr) == 0 {
		return sequence, nil, time.Time{}, errors.New("data packet contains no samples")
	}
	if data.Channel == "" {
		return sequence, nil, time.Time{}, errors.New("channel code is empty")
	}

	chunks := chunkInt32Slice(data.DataArr, CHUNK_SIZE)

	var buf bytes.Buffer

	for i, c := range chunks {
		var miniseed mseedio.MiniSeedData
		miniseed.Init(dataType, mseedio.MSBFIRST)

		chunkStart := miniSeedStartTime(sampleTime(startTime, i*CHUNK_SIZE, data.SampleRate))
		err := miniseed.Append(c, &mseedio.AppendOptions{
			ChannelCode:    data.Channel,
			StationCode:    station,
			LocationCode:   location,
			NetworkCode:    network,
			SampleRate:     float64(data.SampleRate),
			SequenceNumber: fmt.Sprintf("%06d", sequence%1000000),
			StartTime:      chunkStart,
		})
		if err != nil {
			return 0, nil, time.Time{}, err
		}

		// Force 512-byte record
		for i := 0; i < len(miniseed.Series); i++ {
			miniseed.Series[i].BlocketteSection.RecordLength = 9
		}
		slData, err := miniseed.Encode(mseedio.OVERWRITE, mseedio.MSBFIRST)
		if err != nil {
			return 0, nil, time.Time{}, err
		}

		if len(slData) != 512 {
			return sequence, nil, time.Time{}, fmt.Errorf("encoded miniSEED record has length %d, want 512", len(slData))
		}

		slSeq := fmt.Sprintf("SL%06X", uint64(sequence)&0xFFFFFF)
		buf.Write([]byte(slSeq))
		buf.Write(slData)

		endTime = sampleTime(chunkStart, len(c), data.SampleRate)
		sequence++
	}

	return sequence, buf.Bytes(), endTime, nil
}
