package handlers

type STATION struct{}

// Callback of "STATION <...> <...>" command, implements handler interface
func (s *STATION) Callback(client *SeedLinkClient, provider SeedLinkProvider, consumer SeedLinkConsumer, args ...string) error {
	client.Station = s.truncate(args[0], 5)
	client.Network = s.truncate(args[1], 2)
	_, err := client.Write([]byte(RES_OK))
	return err
}

// Fallback of "STATION <...> <...>" command, implements handler interface
func (*STATION) Fallback(client *SeedLinkClient, provider SeedLinkProvider, consumer SeedLinkConsumer, args ...string) {
	client.Close()
}

func (*STATION) truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}

	return s[:n]
}
