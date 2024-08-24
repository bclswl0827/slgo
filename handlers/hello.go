package handlers

import "fmt"

type HELLO struct{}

// Callback of "HELLO" command, implements SeedLinkCommandCallback interface
func (*HELLO) Callback(client *SeedLinkClient, provider SeedLinkProvider, consumer SeedLinkConsumer, args ...string) error {
	station := provider.GetOrganization()
	_, err := client.Write([]byte(fmt.Sprintf("%s\r\n%s\r\n", RELEASE, station)))
	return err
}

// Fallback of "HELLO" command, implements SeedLinkCommandCallback interface
func (*HELLO) Fallback(client *SeedLinkClient, provider SeedLinkProvider, consumer SeedLinkConsumer, args ...string) {
	client.Close()
}
