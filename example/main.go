package main

import (
	"log"

	"github.com/bclswl0827/slgo"
)

const (
	HOST = "0.0.0.0"
	PORT = 18000
)

func main() {
	provider := &provider{}
	consumer := &consumer{}
	hooks := &hooks{}

	server := slgo.New(provider, consumer, hooks)

	log.Printf("starting SeedLink server on %s:%d", HOST, PORT)
	log.Println("test this server with Swarm client: https://volcanoes.usgs.gov/software/swarm/download.shtml")

	err := server.Start(HOST, PORT)
	if err != nil {
		log.Fatalln(err)
	}
}
