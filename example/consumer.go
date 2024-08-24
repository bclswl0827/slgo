package main

import (
	"log"

	"github.com/bclswl0827/slgo/handlers"
)

type consumer struct{}

func (c *consumer) Subscribe(clientId string, eventHandler func(handlers.SeedLinkDataPacket)) error {
	return nil
}

func (c *consumer) Unsubscribe(clientId string) error {
	log.Printf("client %v unsubscribed", clientId)
	return nil
}
