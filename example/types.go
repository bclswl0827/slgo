package main

const (
	SAMPLE_RATE = 100
	TOPIC_NAME  = "seedlink-example"
)

type adcRawData struct {
	SampleRate int
	Timestamp  int64
	Channel_1  []int32
	Channel_2  []int32
	Channel_3  []int32
}

type eventHandler = func(data *adcRawData)
