package main

import (
	"errors"

	"github.com/bclswl0827/slgo/handlers"
	cmap "github.com/orcaman/concurrent-map/v2"
	messagebus "github.com/vardius/message-bus"
)

type consumer struct {
	messageBus  messagebus.MessageBus
	subscribers cmap.ConcurrentMap[string, eventHandler]
}

func (c *consumer) Subscribe(clientId string, channels []string, eventHandler func(handlers.SeedLinkDataPacket)) error {
	if _, ok := c.subscribers.Get(clientId); ok {
		return errors.New("this client has already subscribed")
	}
	handler := func(data *adcRawData) {
		// Match ADC channels to SeedLink channels
		for _, channel := range channels {
			switch channel {
			case "EHZ":
				eventHandler(handlers.SeedLinkDataPacket{
					Timestamp:  data.Timestamp,
					SampleRate: data.SampleRate,
					Channel:    channel,
					DataArr:    data.Channel_1,
				})
			case "EHE":
				eventHandler(handlers.SeedLinkDataPacket{
					Timestamp:  data.Timestamp,
					SampleRate: data.SampleRate,
					Channel:    channel,
					DataArr:    data.Channel_2,
				})
			case "EHN":
				eventHandler(handlers.SeedLinkDataPacket{
					Timestamp:  data.Timestamp,
					SampleRate: data.SampleRate,
					Channel:    channel,
					DataArr:    data.Channel_3,
				})
			}
		}
	}
	c.subscribers.Set(clientId, handler)
	c.messageBus.Subscribe(TOPIC_NAME, handler)
	return nil
}

func (c *consumer) Unsubscribe(clientId string) error {
	fn, ok := c.subscribers.Get(clientId)
	if !ok {
		return errors.New("this client has not subscribed")
	}
	c.messageBus.Unsubscribe(TOPIC_NAME, fn)
	c.subscribers.Remove(clientId)
	return nil
}
