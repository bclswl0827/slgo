package main

import (
	"log"
	"time"

	"github.com/bclswl0827/slgo"
	cmap "github.com/orcaman/concurrent-map/v2"
	messagebus "github.com/vardius/message-bus"
)

const (
	HOST = "0.0.0.0"
	PORT = 18000
)

func main() {
	messageBus := messagebus.New(65535)

	// Create a new ticker that will publish a message every second
	ticker := time.NewTicker(time.Second)
	go func() {
		for {
			<-ticker.C
			messageBus.Publish(TOPIC_NAME, &adcRawData{
				SampleRate: SAMPLE_RATE,
				Timestamp:  time.Now().UnixMilli(),
				Channel_1:  generateRandomArray(SAMPLE_RATE, -32768, 32768),
				Channel_2:  generateRandomArray(SAMPLE_RATE, -32768, 32768),
				Channel_3:  generateRandomArray(SAMPLE_RATE, -32768, 32768),
			})
		}
	}()

	log.Println("test this server with Swarm client: https://volcanoes.usgs.gov/software/swarm/download.shtml")
	log.Printf("starting SeedLink server on %s:%d", HOST, PORT)

	// Create a new SeedLink server with the provider, consumer, and hooks implementations
	server := slgo.New(
		&provider{
			startTime: time.Now().UTC(),
		},
		&consumer{
			messageBus:  messageBus,
			subscribers: cmap.New[eventHandler](),
		},
		&hooks{},
	)
	err := server.Start(HOST, PORT)
	if err != nil {
		log.Fatalln(err)
	}
}
