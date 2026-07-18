package handlers

import "strings"

type STATION struct{}

// Callback of "STATION <...> <...>" command, implements handler interface
func (s *STATION) Callback(client *SeedLinkClient, provider SeedLinkProvider, consumer SeedLinkConsumer, args ...string) error {
	if len(args) < 1 || len(args) > 2 || args[0] == "" {
		_, err := client.Write([]byte(RES_ERR))
		return err
	}

	network := "*"
	if len(args) == 2 {
		network = args[1]
	}
	if !validCode(args[0], "?*-") || !validCode(network, "?*-") {
		_, err := client.Write([]byte(RES_ERR))
		return err
	}

	client.Station = s.truncate(args[0], 5)
	client.Network = s.truncate(network, 2)
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

func validCode(value, extra string) bool {
	return value != "" && strings.IndexFunc(value, func(r rune) bool {
		return !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || strings.ContainsRune(extra, r))
	}) == -1
}
