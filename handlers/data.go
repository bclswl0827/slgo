package handlers

import "strconv"

type DATA struct{}

// Callback of "DATA" command, implements handler interface
func (*DATA) Callback(client *SeedLinkClient, provider SeedLinkProvider, consumer SeedLinkConsumer, args ...string) error {
	client.StartTime = provider.GetCurrentTime()
	if len(args) > 0 {
		seq, err := strconv.ParseInt(args[0], 16, 64)
		if err != nil {
			client.Write([]byte(RES_ERR))
			return err
		}
		client.Sequence = seq + 1
	}
	_, err := client.Write([]byte(RES_OK))
	return err
}

// Fallback of "DATA" command, implements handler interface
func (*DATA) Fallback(client *SeedLinkClient, provider SeedLinkProvider, consumer SeedLinkConsumer, args ...string) {
	client.Close()
}
