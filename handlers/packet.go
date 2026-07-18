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
	if sequence < 0 {
		return sequence, nil, errors.New("sequence number must not be negative")
	}
	if data.SampleRate <= 0 {
		return sequence, nil, errors.New("sample rate must be greater than zero")
	}
	if len(data.DataArr) == 0 {
		return sequence, nil, errors.New("data packet contains no samples")
	}
	if data.Channel == "" {
		return sequence, nil, errors.New("channel code is empty")
	}

	chunks := chunkInt32Slice(data.DataArr, CHUNK_SIZE)

	var buf bytes.Buffer

	for i, c := range chunks {
		var miniseed mseedio.MiniSeedData
		miniseed.Init(dataType, mseedio.MSBFIRST)

		sampleOffset := int64(i * CHUNK_SIZE)
		timeOffset := time.Duration(sampleOffset * int64(time.Second) / int64(data.SampleRate))
		startTime := time.UnixMilli(data.Timestamp).UTC().Add(timeOffset)
		err := miniseed.Append(c, &mseedio.AppendOptions{
			ChannelCode:    data.Channel,
			StationCode:    station,
			LocationCode:   location,
			NetworkCode:    network,
			SampleRate:     float64(data.SampleRate),
			SequenceNumber: fmt.Sprintf("%06d", sequence%1000000),
			StartTime:      startTime,
		})
		if err != nil {
			return 0, nil, err
		}

		// Force 512-byte record
		for i := 0; i < len(miniseed.Series); i++ {
			miniseed.Series[i].BlocketteSection.RecordLength = 9
		}
		slData, err := miniseed.Encode(mseedio.OVERWRITE, mseedio.MSBFIRST)
		if err != nil {
			return 0, nil, err
		}

		if len(slData) != 512 {
			return sequence, nil, fmt.Errorf("encoded miniSEED record has length %d, want 512", len(slData))
		}

		slSeq := fmt.Sprintf("SL%06X", uint64(sequence)&0xFFFFFF)
		buf.Write([]byte(slSeq))
		buf.Write(slData)

		sequence++
	}

	return sequence, buf.Bytes(), nil
}
