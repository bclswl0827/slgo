package main

import (
	"log"

	"github.com/bclswl0827/slgo/handlers"
)

type hooks struct{}

func (h *hooks) OnConnection(client *handlers.SeedLinkClient) {
	log.Printf("client %v connected", client.RemoteAddr())
}

func (h *hooks) OnData(client *handlers.SeedLinkClient, data []byte) {
	log.Printf("%d bytes sent to client %v", len(data), client.RemoteAddr())
}

func (h *hooks) OnClose(client *handlers.SeedLinkClient) {
	log.Printf("client %v disconnected", client.RemoteAddr())
}

func (h *hooks) OnCommand(client *handlers.SeedLinkClient, command []string) {
	log.Printf("client %v issued command %v", client.RemoteAddr(), command)
}
