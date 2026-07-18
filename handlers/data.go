package handlers

import "strconv"

type DATA struct{}

// Callback of "DATA" command, implements handler interface
func (*DATA) Callback(client *SeedLinkClient, provider SeedLinkProvider, consumer SeedLinkConsumer, args ...string) error {
	if len(args) > 2 {
		_, err := client.Write([]byte(RES_ERR))
		return err
	}

	sequence := client.GetSequence()
	if len(args) >= 1 {
		seq, err := strconv.ParseUint(args[0], 16, 24)
		if err != nil {
			_, writeErr := client.Write([]byte(RES_ERR))
			return writeErr
		}
		sequence = int64(seq)
	}

	startTime := provider.GetCurrentTime()
	if len(args) == 2 {
		parsed, err := parseSeedLinkTime(args[1])
		if err != nil {
			_, writeErr := client.Write([]byte(RES_ERR))
			return writeErr
		}
		startTime = parsed
	}
	if startTime.After(provider.GetCurrentTime()) {
		_, err := client.Write([]byte(RES_ERR))
		return err
	}

	client.SetSequence(sequence)
	client.StartTime = startTime
	client.EndTime = provider.GetCurrentTime()
	_, err := client.Write([]byte(RES_OK))
	return err
}

// Fallback of "DATA" command, implements handler interface
func (*DATA) Fallback(client *SeedLinkClient, provider SeedLinkProvider, consumer SeedLinkConsumer, args ...string) {
	client.Close()
}
