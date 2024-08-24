package handlers

type SELECT struct{}

// Callback of "SELECT <...>" command, implements handler interface
func (*SELECT) Callback(client *SeedLinkClient, provider SeedLinkProvider, consumer SeedLinkConsumer, args ...string) error {
	if len(args) < 1 {
		_, err := client.Write([]byte(RES_ERR))
		return err
	} else {
		if len(args[0]) < 5 {
			_, err := client.Write([]byte(RES_ERR))
			return err
		} else {
			client.Location = args[0][:2]
			client.Channels = append(client.Channels, args[0][2:5])
		}
	}
	_, err := client.Write([]byte(RES_OK))
	return err
}

// Fallback of "SELECT <...>" command, implements handler interface
func (*SELECT) Fallback(client *SeedLinkClient, provider SeedLinkProvider, consumer SeedLinkConsumer, args ...string) {
	client.Close()
}
