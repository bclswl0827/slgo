package handlers

type CAPABILITIES struct{}

// Callback of "CAPABILITIES" command, implements SeedLinkCommandCallback interface
func (*CAPABILITIES) Callback(client *SeedLinkClient, provider SeedLinkProvider, consumer SeedLinkConsumer, args ...string) error {
	_, err := client.Write([]byte(RES_OK))
	return err
}

// Fallback of "CAPABILITIES" command, implements SeedLinkCommandCallback interface
func (*CAPABILITIES) Fallback(client *SeedLinkClient, provider SeedLinkProvider, consumer SeedLinkConsumer, args ...string) {
	client.Close()
}
