package handlers

type BATCH struct{}

// Callback of "BATCH" command, implements handler interface
func (*BATCH) Callback(client *SeedLinkClient, provider SeedLinkProvider, consumer SeedLinkConsumer, args ...string) error {
	_, err := client.Write([]byte(RES_OK))
	return err
}

// Fallback of "BATCH" command, implements handler interface
func (*BATCH) Fallback(client *SeedLinkClient, provider SeedLinkProvider, consumer SeedLinkConsumer, args ...string) {
	client.Close()
}
